package autocheck

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/opsee/basic/schema"
	opsee_cloudwatch "github.com/opsee/basic/schema/aws/cloudwatch"
	opsee_types "github.com/opsee/protobuf/opseeproto/types"
	"github.com/stretchr/testify/assert"
	"testing"
)

var rdsTests = []struct {
	rds    *rds.DBInstance
	checks []*schema.Check
}{
	{
		rds: &rds.DBInstance{
			DBInstanceIdentifier: aws.String("invalid-db"),
			DBInstanceClass:      aws.String("db.invalid.xtiny"),
		},
		checks: writeChecks("invalid-db", "db.invalid.xtiny", 95.000, 0.0, 0.0, nil),
	},
	{
		rds: &rds.DBInstance{
			DBInstanceIdentifier: aws.String("opsee-test-db"),
			DBInstanceClass:      aws.String("db.m3.xlarge"),
		},
		checks: writeChecks("opsee-test-db", "db.m3.xlarge",
			95.000,
			((15.0*1e9)/12582880.0)*0.85,
			(15.0*1e9)*0.1,
			nil),
	},
}

var rdsAlarmTests = []struct {
	rds    *rds.DBInstance
	checks []*schema.Check
}{
	{
		rds: &rds.DBInstance{
			DBInstanceIdentifier: aws.String("opsee-test2-db"),
			DBInstanceClass:      aws.String("db.m3.xlarge"),
		},
		checks: writeChecks("opsee-test2-db", "db.m3.xlarge",
			75.000,
			((15.0*1e9)/12582880.0)*0.85,
			(15.0*1e9)*0.1,
			aws.Float64(0.001)),
	},
}

func TestRDSGenerate(t *testing.T) {
	assert := assert.New(t)

	for _, r := range rdsTests {
		cz, err := NewTarget(r.rds).Generate()
		assert.NoError(err)
		assert.EqualValues(r.checks, cz)
	}

	for _, r := range rdsAlarmTests {
		cz, err := NewTargetWithAlarms(r.rds, writeAlarms()).Generate()
		assert.NoError(err)
		assert.EqualValues(r.checks, cz)
	}
}

func writeAlarms() []*opsee_cloudwatch.MetricAlarm {
	return []*opsee_cloudwatch.MetricAlarm{
		&opsee_cloudwatch.MetricAlarm{
			AlarmName:          aws.String("awsrds-devpg-CPU-Utilization"),
			MetricName:         aws.String("CPUUtilization"),
			StateValue:         aws.String("OK"),
			Namespace:          aws.String("AWS/RDS"),
			Statistic:          aws.String("Average"),
			ComparisonOperator: aws.String("GreaterThanOrEqualToThreshold"),
			Threshold:          aws.Float64(75.00),
			Period:             aws.Int64(300),
			Dimensions: []*opsee_cloudwatch.Dimension{
				&opsee_cloudwatch.Dimension{
					Name:  aws.String("DBInstanceIdentifier"),
					Value: aws.String("opsee-test2-db"),
				},
			},
		},
		&opsee_cloudwatch.MetricAlarm{
			AlarmName:          aws.String("awsrds-devpg-Read-Latency"),
			MetricName:         aws.String("ReadLatency"),
			StateValue:         aws.String("OK"),
			Namespace:          aws.String("AWS/RDS"),
			Statistic:          aws.String("Average"),
			ComparisonOperator: aws.String("LessThanThreshold"),
			Threshold:          aws.Float64(0.001),
			Period:             aws.Int64(300),
			Dimensions: []*opsee_cloudwatch.Dimension{
				&opsee_cloudwatch.Dimension{
					Name:  aws.String("DBInstanceIdentifier"),
					Value: aws.String("opsee-test2-db"),
				},
			},
		},
	}
}

func writeChecks(dbName string, dbclass string, cpuThresh float64, dbConns float64, memThresh float64, readThresh *float64) []*schema.Check {
	cwMetrics := []struct {
		checkName string
		rel       string
		op        float64
		dispName  string
	}{
		{
			checkName: "CPUUtilization",
			rel:       "lessThan",
			op:        cpuThresh,
			dispName:  "CPU Utilization",
		},
		{
			checkName: "DatabaseConnections",
			rel:       "lessThan",
			op:        dbConns,
			dispName:  "Connection Count",
		},
		{
			checkName: "FreeableMemory",
			rel:       "greaterThan",
			op:        memThresh,
			dispName:  "Available Memory",
		},
	}

	if readThresh != nil {
		cwMetrics = append(cwMetrics, struct {
			checkName string
			rel       string
			op        float64
			dispName  string
		}{
			checkName: "ReadLatency",
			rel:       "lessThan",
			op:        aws.Float64Value(readThresh),
			dispName:  "Read Latency",
		})
	}

	rdsCheck := &schema.CloudWatchCheck{}

	check := &schema.Check{
		Name:     fmt.Sprintf("RDS metrics for %s (auto)", dbName),
		Interval: int32(60),
		Target: &schema.Target{
			Name: dbName,
			Type: "dbinstance",
			Id:   dbName,
		},
	}

	for _, m := range cwMetrics {
		if m.op == 0.0 {
			continue
		}

		stringOp := fmt.Sprintf("%.3f", m.op)
		if m.checkName == "DatabaseConnections" {
			stringOp = fmt.Sprintf("%d", int(m.op))
		}

		rdsCheck.Metrics = append(rdsCheck.Metrics, &schema.CloudWatchMetric{
			Namespace: "AWS/RDS",
			Name:      m.checkName,
		})

		check.Assertions = append(check.Assertions, &schema.Assertion{
			Key:          "cloudwatch",
			Relationship: m.rel,
			Operand:      stringOp,
			Value:        m.checkName,
		})
	}

	checkSpec, _ := opsee_types.MarshalAny(rdsCheck)
	check.CheckSpec = checkSpec

	return []*schema.Check{check}
}
