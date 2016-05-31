package launcher

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/opsee/awscan"
	opsee_cloudwatch "github.com/opsee/basic/schema/aws/cloudwatch"
	"github.com/opsee/basic/service"
	"github.com/opsee/keelhaul/bus"
	opsee_types "github.com/opsee/protobuf/opseeproto/types"
	"golang.org/x/net/context"
	"net/http"
	"reflect"
	"time"
)

type vpcDiscovery struct{}

const (
	instanceErrorThreshold = 0.3
	defaultTTL             = time.Second * time.Duration(5)
	maxPages               = 10
)

func (s vpcDiscovery) Execute(launch *Launch) {
	var (
		instances   = make(map[string]bool)
		dbInstances = make(map[string]bool)
		disco       = awscan.NewDiscoverer(awscan.NewScanner(launch.session, launch.Bastion.VPCID))
		httpclient  = &http.Client{}
	)

	launch.event(&bus.Message{
		State:   stateInProgress,
		Command: commandDiscovery,
		Message: "starting vpc environment discovery",
	})

	// fetch all cloudwatch alarms up-front for use in autocheck creation
	cwAlarms := fetchAlarms(launch)

	for event := range disco.Discover() {
		if event.Err != nil {
			switch event.Err.(*awscan.DiscoveryError).Type {
			case awscan.InstanceType, awscan.DBInstanceType:
				launch.VPCEnvironment.InstanceErrorCount++
			case awscan.SecurityGroupType, awscan.DBSecurityGroupType, awscan.AutoScalingGroupType, awscan.LoadBalancerType:
				launch.VPCEnvironment.GroupErrorCount++
			default:
				continue
			}

			s.handleError(event.Err, launch)
		} else {
			messageType := reflect.ValueOf(event.Result).Elem().Type().Name()
			messageBody, err := json.Marshal(event.Result)
			if err != nil {
				s.handleError(err, launch)
				continue
			}

			request, err := http.NewRequest("POST", launch.config.FieriEndpoint+"/entity/"+messageType, bytes.NewBuffer(messageBody))
			if err != nil {
				s.handleError(err, launch)
				continue
			}

			request.Header.Set("Content-Type", "application/json")
			request.Header.Set("Accept", "application/json")
			request.Header.Set("Customer-Id", launch.User.CustomerId)

			resp, err := httpclient.Do(request)
			if err != nil {
				s.handleError(err, launch)
				continue
			}

			resp.Body.Close()
			if resp.StatusCode != 201 {
				s.handleError(fmt.Errorf("bad response from fieri: %s", resp.Status), launch)
				continue
			}

			switch messageType {
			case awscan.InstanceType:
				// we'll have to de-dupe instances so use a ghetto set (map)
				i, ok := event.Result.(*ec2.Instance)
				if !ok {
					launch.VPCEnvironment.InstanceErrorCount++
					s.handleError(fmt.Errorf("failed ec2 instance type assertion"), launch)
					continue
				}

				instances[*i.InstanceId] = true
				launch.VPCEnvironment.InstanceCount = card(instances)
				// disable until frontend is ready (mike)
				//launch.Autochecks.AddTarget(event.Result)

			case awscan.DBInstanceType:
				// we'll have to de-dupe instances so use a ghetto set (map)
				i, ok := event.Result.(*rds.DBInstance)
				if !ok {
					launch.VPCEnvironment.InstanceErrorCount++
					s.handleError(fmt.Errorf("failed rds db instance type assertion"), launch)
					continue
				}
				dbInstances[*i.DBInstanceIdentifier] = true
				launch.VPCEnvironment.DBInstanceCount = card(dbInstances)
				rdsAlarms := filterAlarms(cwAlarms, "AWS/RDS")
				if rdsAlarms != nil {
					launch.Autochecks.AddTargetWithAlarms(event.Result, rdsAlarms)
				} else {
					launch.Autochecks.AddTarget(event.Result)
				}

			case awscan.SecurityGroupType:
				launch.VPCEnvironment.SecurityGroupCount++

			case awscan.DBSecurityGroupType:
				launch.VPCEnvironment.DBSecurityGroupCount++

			case awscan.AutoScalingGroupType:
				launch.VPCEnvironment.AutoscalingGroupCount++

			case awscan.LoadBalancerType:
				launch.Autochecks.AddTarget(event.Result)
				launch.VPCEnvironment.LoadBalancerCount++
			}

			launch.logger.WithField("resource-type", messageType).Info("resource discovery")
		}
	}

	if launch.VPCEnvironment.tooManyErrors() {
		launch.error(launch.VPCEnvironment.LastError, &bus.Message{
			Command: commandDiscovery,
			Message: "too many discovery errors",
		})
		return
	}

	launch.event(&bus.Message{
		State:   stateComplete,
		Command: commandDiscovery,
		Message: "vpc environment discovery complete",
	})
}

// we have a custom error handler for this stage, since errors may be recoverable
func (s vpcDiscovery) handleError(err error, launch *Launch) {
	launch.VPCEnvironment.LastError = err
	launch.logger.WithError(err).Error("vpc discovery error, potentially ignoring")
}

func (v *VPCEnvironment) tooManyErrors() bool {
	total := v.InstanceCount + v.DBInstanceCount
	if total == 0 {
		return v.GroupErrorCount > 0
	}

	return v.GroupErrorCount > 0 ||
		float64(v.InstanceErrorCount)/float64(total) > instanceErrorThreshold
}

func card(m map[string]bool) int {
	i := 0
	for _, _ = range m {
		i++
	}
	return i
}

func fetchAlarms(launch *Launch) []*opsee_cloudwatch.MetricAlarm {
	var (
		next      *string
		cwAlarms  = make([]*opsee_cloudwatch.MetricAlarm, 0)
		timestamp = &opsee_types.Timestamp{}
	)
	timestamp.Scan(time.Now().UTC().Add(defaultTTL * -1))
	for i := 0; i < maxPages; i++ {
		params := &opsee_cloudwatch.DescribeAlarmsInput{
			NextToken: next,
		}
		resp, err := launch.bezos.Get(
			context.Background(),
			&service.BezosRequest{
				User:   launch.User,
				Region: *launch.session.Config.Region,
				VpcId:  launch.Bastion.VPCID,
				MaxAge: timestamp,
				Input:  &service.BezosRequest_Cloudwatch_DescribeAlarmsInput{params},
			})
		if err != nil {
			launch.logger.WithError(err).Error("describe alarms request error")
			return cwAlarms
		}
		output := resp.GetCloudwatch_DescribeAlarmsOutput()
		if output == nil {
			launch.logger.WithError(err).Error("describe alarms output error")
			return cwAlarms
		}
		cwAlarms = append(cwAlarms, output.MetricAlarms...)
		if output.NextToken != nil {
			next = output.NextToken
			continue
		}
		return cwAlarms
	}

	launch.logger.WithField("cust", launch.User.CustomerId).Info("describe alarms max pages reached")
	return cwAlarms
}

func filterAlarms(alarms []*opsee_cloudwatch.MetricAlarm, namespace string) []*opsee_cloudwatch.MetricAlarm {
	filtered := make([]*opsee_cloudwatch.MetricAlarm, 0, len(alarms))

	for _, a := range alarms {
		if *a.Namespace == namespace {
			filtered = append(filtered, a)
		}
	}

	return filtered
}
