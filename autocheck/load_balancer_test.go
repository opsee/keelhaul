package autocheck

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/opsee/basic/schema"
	opsee_types "github.com/opsee/protobuf/opseeproto/types"
	"github.com/stretchr/testify/assert"
	"testing"
)

var elbtests = []struct {
	elb    *elb.LoadBalancerDescription
	checks []*schema.Check
}{
	{
		elb: &elb.LoadBalancerDescription{
			LoadBalancerName: aws.String("no checks"),
			HealthCheck: &elb.HealthCheck{
				Target: aws.String("TCP:9000"),
			},
		},
		checks: []*schema.Check{},
	},
	{
		elb: &elb.LoadBalancerDescription{
			LoadBalancerName: aws.String("me http"),
			HealthCheck: &elb.HealthCheck{
				Target: aws.String("HTTP:9000/topepe"),
			},
		},
		checks: testChecks("me http", "http me http (auto)", "http", "/topepe", 9000),
	},
	{
		elb: &elb.LoadBalancerDescription{
			LoadBalancerName: aws.String("me secure load ranger"),
			HealthCheck: &elb.HealthCheck{
				Target: aws.String("HTTPS:9/topepeeee"),
			},
		},
		checks: testChecks("me secure load ranger", "https me secure load ranger (auto)", "https", "/topepeeee", 9),
	},
}

func TestELBGenerate(t *testing.T) {
	assert := assert.New(t)

	for _, t := range elbtests {
		checks, err := NewTarget(t.elb).Generate()
		assert.NoError(err)
		assert.EqualValues(t.checks, checks)
	}
}

func testChecks(lbName, name, protocol, path string, port int32) []*schema.Check {
	httpCheck := &schema.HttpCheck{
		Name:     name,
		Path:     path,
		Protocol: protocol,
		Port:     port,
		Verb:     "GET",
	}

	checkSpec, _ := opsee_types.MarshalAny(httpCheck)
	check := &schema.Check{
		Name:     name,
		Interval: int32(30),
		Target: &schema.Target{
			Name: lbName,
			Type: "elb",
			Id:   lbName,
		},
		CheckSpec: checkSpec,
		Assertions: []*schema.Assertion{
			{
				Key:          "code",
				Relationship: "equal",
				Operand:      "200",
			},
		},
	}

	return []*schema.Check{check}
}
