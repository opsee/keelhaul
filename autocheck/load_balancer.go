package autocheck

import (
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/opsee/basic/schema"
	opsee_types "github.com/opsee/protobuf/opseeproto/types"
	"regexp"
	"strconv"
	"strings"
)

var (
	errNoLoadBalancerDescription = errors.New("no load balancer description")
	errNoLoadBalancerName        = errors.New("no load balancer name")
	errNoHealthCheck             = errors.New("no health check")

	elbTargetRegexp = regexp.MustCompile(`^(HTTPS?):(\d+)(/.*)$`)
)

type LoadBalancer struct {
	*elb.LoadBalancerDescription
}

func (l LoadBalancer) Generate() ([]*schema.Check, error) {
	lbd := l.LoadBalancerDescription
	if lbd == nil {
		return nil, errNoLoadBalancerDescription
	}

	if lbd.HealthCheck == nil {
		return nil, errNoHealthCheck
	}

	if lbd.LoadBalancerName == nil {
		return nil, errNoLoadBalancerName
	}

	var (
		checks        = make([]*schema.Check, 0)
		lbName        = aws.StringValue(lbd.LoadBalancerName)
		target        = aws.StringValue(lbd.HealthCheck.Target)
		targetMatches = elbTargetRegexp.FindStringSubmatch(target)
	)

	if len(targetMatches) < 4 {
		return checks, nil
	}

	var (
		protocol = strings.ToLower(targetMatches[1])
		path     = targetMatches[3]
	)

	port, err := strconv.ParseInt(targetMatches[2], 10, 32)
	if err != nil {
		return nil, fmt.Errorf("error parsing port %s", targetMatches[2])
	}

	switch protocol {
	case "http", "https":
		name := fmt.Sprintf("%s %s (auto)", protocol, lbName)

		httpCheck := &schema.HttpCheck{
			Name:     name,
			Path:     path,
			Protocol: protocol,
			Port:     int32(port),
			Verb:     "GET",
		}

		checkSpec, err := opsee_types.MarshalAny(httpCheck)
		if err != nil {
			return nil, err
		}

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
			// Spec <--- TODO: fill this out later when using cats
		}

		checks = append(checks, check)
	default:
	}

	return checks, nil
}
