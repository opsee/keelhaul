package launcher

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/opsee/awscan"
	"github.com/opsee/keelhaul/com"
	"net/http"
	"reflect"
)

type vpcDiscovery struct{}

const (
	instanceErrorThreshold = 0.3
)

func (s vpcDiscovery) Execute(launch *Launch) {
	var (
		instances   = make(map[string]bool)
		dbInstances = make(map[string]bool)
		disco       = awscan.NewDiscoverer(awscan.NewScanner(launch.session, launch.Bastion.VPCID))
		httpclient  = &http.Client{}
	)

	launch.event(&com.Message{
		State:   stateInProgress,
		Command: commandDiscovery,
		Message: "starting vpc environment discovery",
	})

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
			request.Header.Set("Customer-Id", launch.User.CustomerID)

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

			case awscan.SecurityGroupType:
				launch.VPCEnvironment.SecurityGroupCount++

			case awscan.DBSecurityGroupType:
				launch.VPCEnvironment.DBSecurityGroupCount++

			case awscan.AutoScalingGroupType:
				launch.VPCEnvironment.AutoscalingGroupCount++

			case awscan.LoadBalancerType:
				launch.VPCEnvironment.LoadBalancerCount++
			}

			launch.logger.WithField("resource-type", messageType).Info("resource discovery")
		}
	}

	if launch.VPCEnvironment.tooManyErrors() {
		launch.error(launch.VPCEnvironment.LastError, &com.Message{
			Command: commandDiscovery,
			Message: "too many discovery errors",
		})
		return
	}

	launch.event(&com.Message{
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
