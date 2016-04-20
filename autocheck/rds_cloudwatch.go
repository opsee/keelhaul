package autocheck

import (
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/opsee/basic/schema"
	"regexp"
	"strconv"
	"strings"
)

var (
	errNoDB = errors.New("no rds db")
	CloudWatchStatisticsPeriod = 60
)

type RDSCloudWatch struct {
	*rds.DBInstance
	connCountThreshold int
	cpuThresh int
	memThresh int
}

func NewRDSCloudWatch(dbinstance *rds.DBInstance, cloudwatchClient *cloudwatch.CloudWatch) (*RDSCloudWatch, error) {
	rdsCloudWatch := RDSCloudWatch{dbinstance}
	rdsCloudWatch.runConnCountChecks(cloudwatchClient)
}

func (rc RDSCloudWatch) runConnCountChecks(cloudwatchClient *cloudwatch.CloudWatch) error {
	req := &CloudWatchRequest{
		target: rc.DBInstance.DBInstanceIdentifier,
		metrics: []*string{
			"DatabaseConnections",
		}
		namespace: "AWS/RDS",
		statisticsIntervalSecs: 300,
		statisticsPeriod: 60,
		statistics: []*string {
			"Average"
		}
	}

	resp, err := req.getStats(cloudwatchClient)
	// XXX
}

func (rc RDSCloudWatch) Generate() ([]*schema.Check, error) {
	db := rc.DBInstance
	if db == nil {
		return nil, errNoDB
	}

	var (
		checks = make([]*schema.Check, 0)
		dbName = aws.StringValue(db.DBInstanceIdentifier)
	)

	name := fmt.Sprintf("RDS %s (auto)", dbName)

	clwCheck := &schema.CloudWatchCheck{
		Metrics: []*schema.CloudWatchMetric{
			&schema.CloudWatchMetric{
				Namespace: "AWS/RDS",
				Name: "DatabaseConnections",
			},
		},
	}
	checkSpec, err := schema.MarshalAny(clwCheck)
	if err != nil {
		return nil, err
	}

	check := &schema.Check{
		Name:     name,
		Interval: int32(30),
		Target: &schema.Target{
			Name: dbName,
			Type: "cloudwatch",
			Id:   dbName,
		},
		CheckSpec: checkSpec,
		Assertions: []*schema.Assertion{
			{
				Key:          "cloudwatch",
				Relationship: "equal",
				Operand:      rc.connCountThreshold, // XXX
				Value:        "DatabaseConnections",
			},
		},
	}

	checks = append(checks, check)

	return checks, nil
}
