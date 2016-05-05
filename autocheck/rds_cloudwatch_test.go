package autocheck

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/opsee/basic/schema"
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
		checks: writeChecks("invalid-db", "db.invalid.xtiny"),
	},
	{
		rds: &rds.DBInstance{
			DBInstanceIdentifier: aws.String("opsee-test-db"),
			DBInstanceClass:      aws.String("db.m3.xlarge"),
		},
		checks: writeChecks("opsee-test-db", "db.m3.xlarge"),
	},
}

func TestRDSGenerate(t *testing.T) {
	assert := assert.New(t)

	for _, r := range rdsTests {
		cz, err := NewTarget(r.rds).Generate()
		assert.NoError(err)
		assert.EqualValues(r.checks, cz)
	}
}

func writeChecks(dbName string, dbclass string) []*schema.Check {
	instanceMem := GetInstanceClassMemory(dbclass)
	cwMetrics := []struct {
		checkName string
		rel       string
		op        float64
		dispName  string
	}{
		{
			checkName: "CPUUtilization",
			rel:       "lessThan",
			op:        95.000,
			dispName:  "CPU Utilization",
		},
		{
			checkName: "DatabaseConnections",
			rel:       "lessThan",
			op:        (instanceMem / 12582880) * 0.85,
			dispName:  "Connection Count",
		},
		{
			checkName: "FreeableMemory",
			rel:       "greaterThan",
			op:        instanceMem * 0.10,
			dispName:  "Available Memory",
		},
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

		rdsCheck.Metrics = append(rdsCheck.Metrics, &schema.CloudWatchMetric{
			Namespace: "AWS/RDS",
			Name:      m.checkName,
		})

		check.Assertions = append(check.Assertions, &schema.Assertion{
			Key:          "cloudwatch",
			Relationship: m.rel,
			Operand:      fmt.Sprintf("%.3f", m.op),
			Value:        m.checkName,
		})
	}

	checkSpec, _ := opsee_types.MarshalAny(rdsCheck)
	check.CheckSpec = checkSpec

	return []*schema.Check{check}
}
