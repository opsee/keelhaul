package autocheck

import (
	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/opsee/basic/schema"
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
	default:
		return EmptyTarget{}
	}
}
