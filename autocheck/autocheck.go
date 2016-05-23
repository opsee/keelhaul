package autocheck

import (
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/opsee/basic/schema"
	opsee_cloudwatch "github.com/opsee/basic/schema/aws/cloudwatch"
)

type Target interface {
	Generate() ([]*schema.Check, error)
}

type EmptyTarget struct{}

func (t EmptyTarget) Generate() ([]*schema.Check, error) {
	return []*schema.Check{}, nil
}

func NewTarget(obj interface{}) Target {
	switch o := obj.(type) {
	case *elb.LoadBalancerDescription:
		return LoadBalancer{o}
	case *ec2.Instance:
		return EC2CloudWatch{o}
	case *rds.DBInstance:
		return RDSCloudWatch{o, nil}
	default:
		return EmptyTarget{}
	}
}

func NewTargetWithAlarms(obj interface{}, alarms interface{}) Target {
	switch o := obj.(type) {
	case *rds.DBInstance:
		switch a := alarms.(type) {
		case []*opsee_cloudwatch.MetricAlarm:
			return RDSCloudWatch{o, a}
		}
	}
	return EmptyTarget{}
}
