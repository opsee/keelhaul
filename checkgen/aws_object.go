package checkgen

import (
	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/opsee/basic/schema"
	"reflect"
)

var (
	AWSTypeFactory = make(map[string]reflect.Type)
)

func init() {
	AWSTypeFactory[reflect.TypeOf(elb.LoadBalancerDescription{}).Name()] = reflect.TypeOf(elb.LoadBalancerDescription{})
}

type AWSObject struct {
	Type   string
	Object interface{}
	Owner  *schema.User
}
