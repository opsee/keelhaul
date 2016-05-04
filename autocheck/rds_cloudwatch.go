package autocheck

import (
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/opsee/basic/schema"
	opsee_types "github.com/opsee/protobuf/opseeproto/types"
)

var (
	errNoDB      = errors.New("no rds db")
	maxConnRatio = 0.85
	minMemRatio  = 0.10
	maxCPUUtil   = 95
)

type RDSCloudWatch struct {
	*rds.DBInstance
}

func (rc RDSCloudWatch) Generate() ([]*schema.Check, error) {
	dbinst := rc.DBInstance
	if dbinst == nil {
		return nil, errNoDB
	}

	var (
		dbName = aws.StringValue(dbinst.DBInstanceIdentifier)
		name   = fmt.Sprintf("RDS metrics for %s (auto)", dbName)
		// RDS DB instance max connections are proporitional to instance class resources, i.e.,
		// 		max_connections = DBInstanceClassMemoryBytes / 12582880
		maxConnections = GetInstanceClassMemory(*rc.DBInstance.DBInstanceClass) / 12582880
		minFreeMem     = GetInstanceClassMemory(*rc.DBInstance.DBInstanceClass) * minMemRatio

		metrics    []*schema.CloudWatchMetric
		assertions []*schema.Assertion
	)

	metrics = append(metrics, &schema.CloudWatchMetric{
		Namespace: "AWS/RDS",
		Name:      "CPUUtilization",
	})
	assertions = append(assertions, &schema.Assertion{
		Key:          "cloudwatch",
		Relationship: "lessThan",
		Operand:      fmt.Sprintf("%.3f", maxCPUUtil),
		Value:        "CPUUtilization",
	})

	if maxConnections > 0 {
		metrics = append(metrics, &schema.CloudWatchMetric{
			Namespace: "AWS/RDS",
			Name:      "DatabaseConnections",
		})
		assertions = append(assertions, &schema.Assertion{
			Key:          "cloudwatch",
			Relationship: "lessThan",
			Operand:      fmt.Sprintf("%.3f", (maxConnections * maxConnRatio)),
			Value:        "DatabaseConnections",
		})
	}

	if minFreeMem > 0 {
		metrics = append(metrics, &schema.CloudWatchMetric{
			Namespace: "AWS/RDS",
			Name:      "FreeableMemory",
		})
		assertions = append(assertions, &schema.Assertion{
			Key:          "cloudwatch",
			Relationship: "greaterThan",
			Operand:      fmt.Sprintf("%.3f", minFreeMem),
			Value:        "FreeableMemory",
		})
	}

	clwCheck := &schema.CloudWatchCheck{
		Metrics: metrics,
	}
	checkSpec, err := opsee_types.MarshalAny(clwCheck)
	if err != nil {
		return nil, err
	}
	check := &schema.Check{
		Name:     name,
		Interval: int32(60),
		Target: &schema.Target{
			Name: dbName,
			Type: "dbinstance",
			Id:   dbName,
		},
		CheckSpec:  checkSpec,
		Assertions: assertions,
	}

	return []*schema.Check{check}, nil
}

func GetInstanceClassMemory(dbInstClass string) float64 {
	instClassMemGB := 0.0

	switch dbInstClass {
	case "db.t1.micro":
		instClassMemGB = 0.615
	case "db.m1.small":
		instClassMemGB = 1.7
	case "db.m4.large":
		instClassMemGB = 8
	case "db.m4.xlarge":
		instClassMemGB = 16
	case "db.m4.2xlarge":
		instClassMemGB = 32
	case "db.m4.4xlarge":
		instClassMemGB = 64
	case "db.m4.10xlarge":
		instClassMemGB = 160
	case "db.r3.large":
		instClassMemGB = 15
	case "db.r3.xlarge":
		instClassMemGB = 30.5
	case "db.r3.2xlarge":
		instClassMemGB = 61
	case "db.r3.4xlarge":
		instClassMemGB = 122
	case "db.r3.8xlarge":
		instClassMemGB = 244
	case "db.t2.micro":
		instClassMemGB = 1
	case "db.t2.small":
		instClassMemGB = 2
	case "db.t2.medium":
		instClassMemGB = 4
	case "db.t2.large":
		instClassMemGB = 8
	case "db.m3.medium":
		instClassMemGB = 3.75
	case "db.m3.large":
		instClassMemGB = 7.5
	case "db.m3.xlarge":
		instClassMemGB = 15
	case "db.m3.2xlarge":
		instClassMemGB = 30
	case "db.m2.xlarge":
		instClassMemGB = 17.1
	case "db.m2.2xlarge":
		instClassMemGB = 34.2
	case "db.m2.4xlarge":
		instClassMemGB = 68.4
	case "db.cr1.8xlarge":
		instClassMemGB = 244
	}

	return instClassMemGB * 1e9
}
