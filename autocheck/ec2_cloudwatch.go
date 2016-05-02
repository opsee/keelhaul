package autocheck

import (
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/opsee/basic/schema"
	opsee_types "github.com/opsee/protobuf/opseeproto/types"
)

var (
	errNoInstance = errors.New("no ec2 instance")
	maxEC2CPUUtil = 0.95
)

type EC2CloudWatch struct {
	*ec2.Instance
}

func (ec EC2CloudWatch) Generate() ([]*schema.Check, error) {
	instance := ec.Instance
	if instance == nil {
		return nil, errNoInstance
	}

	var (
		checks = make([]*schema.Check, 0, 3)
		instID = aws.StringValue(instance.InstanceId)
	)

	// CPU Util check
	name := fmt.Sprintf("EC2 (%s) CPU Utilization (auto)", instID)
	clwCheck := &schema.CloudWatchCheck{
		Metrics: []*schema.CloudWatchMetric{
			&schema.CloudWatchMetric{
				Namespace: "AWS/EC2",
				Name:      "CPUUtilization",
			},
		},
	}
	checkSpec, err := opsee_types.MarshalAny(clwCheck)
	if err != nil {
		return nil, err
	}
	check := &schema.Check{
		Name:     name,
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
				Relationship: "lessThan",
				Operand:      fmt.Sprintf("%.3f", maxEC2CPUUtil),
				Value:        "CPUUtilization",
			},
		},
	}
	checks = append(checks, check)

	return checks, nil
}
