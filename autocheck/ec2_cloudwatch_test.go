package autocheck

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/opsee/basic/schema"
	opsee_types "github.com/opsee/protobuf/opseeproto/types"
	"github.com/stretchr/testify/assert"
	"testing"
)

var ec2Tests = []struct {
	ec2    *ec2.Instance
	checks []*schema.Check
}{
	{
		ec2: &ec2.Instance{
			InstanceId: aws.String("i-e48b6a39"),
		},
		checks: writeEC2Checks("i-e48b6a39"),
	},
}

func TestEC2Generate(t *testing.T) {
	assert := assert.New(t)

	for _, r := range ec2Tests {
		cz, err := NewTarget(r.ec2).Generate()
		assert.NoError(err)
		assert.EqualValues(r.checks, cz)
	}
}

func writeEC2Checks(instID string) []*schema.Check {
	checks := make([]*schema.Check, 0)
	cwMetrics := []struct {
		checkName string
		rel       string
		op        float64
		dispName  string
	}{
		{
			checkName: "CPUUtilization",
			rel:       "lessThan",
			op:        0.950,
			dispName:  "CPU Utilization",
		},
	}

	for _, m := range cwMetrics {
		if m.op == 0.0 {
			continue
		}
		ec2Check := &schema.CloudWatchCheck{
			Metrics: []*schema.CloudWatchMetric{
				&schema.CloudWatchMetric{
					Namespace: "AWS/EC2",
					Name:      m.checkName,
				},
			},
		}
		checkSpec, _ := opsee_types.MarshalAny(ec2Check)
		check := &schema.Check{
			Name:     fmt.Sprintf("EC2 (%s) %s (auto)", instID, m.dispName),
			Interval: int32(60),
			Target: &schema.Target{
				Name: instID,
				Type: "cloudwatch",
				Id:   instID,
			},
			CheckSpec: checkSpec,
			Assertions: []*schema.Assertion{
				{
					Key:          "cloudwatch",
					Relationship: m.rel,
					Operand:      fmt.Sprintf("%.3f", m.op),
					Value:        m.checkName,
				},
			},
		}
		checks = append(checks, check)
	}

	return checks
}
