package checkgen

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/golang/protobuf/jsonpb"
	"github.com/hoisie/mustache"
	"github.com/opsee/basic/com"
	"github.com/opsee/keelhaul/checker"
	"github.com/sirupsen/logrus"
)

var checkNameTemplate *mustache.Template

func init() {
	tmpl, err := mustache.ParseString(`{{protocol_name}} {{target_name}} (auto)`)
	if err != nil {
		panic(err)
	}
	checkNameTemplate = tmpl
}

// TODO: hack to get this working.  Use protobuf instead as this could get out of sync with Checks.
type CheckFactoryCheck struct {
	Check              checker.Check
	CheckJson          string
	CheckAssertions    Assertions
	CheckNotifications Notifications
}

type ChecksFactory interface {
	ProduceChecks(awsobj *com.AWSObject) chan *CheckFactoryCheck
}

type ELBCheckFactory struct{}

func (elbFactory *ELBCheckFactory) ProduceChecks(awsobj *com.AWSObject) chan *CheckFactoryCheck {
	lb, ok := awsobj.Object.(*elb.LoadBalancerDescription)
	checks := make(chan *CheckFactoryCheck, 1)
	if !ok {
		logrus.WithFields(logrus.Fields{"module": "checkgen", "event": "ProduceChecks", "RetVal": lb, "OK": ok}).Error("Couldn't assert type *elb.LoadBalancerDescription on aws.Object")
		return checks
	} else {
		if lb.HealthCheck != nil {

			protocol := *lb.HealthCheck.Target

			targetProtocol := strings.ToLower(protocol[:strings.Index(protocol, ":")])
			switch targetProtocol {

			case "http", "https":
				targetPort := protocol[strings.Index(protocol, ":")+1 : strings.Index(protocol, "/")]
				targetPortInt, err := strconv.Atoi(targetPort)
				if err != nil {
					logrus.WithFields(logrus.Fields{"module": "checkgen", "event": "ProduceChecks", "targetPort": targetPort}).Warn("Couldn't convert target port to int.")
					return checks
				}

				targetPath := protocol[strings.Index(protocol, "/"):]
				if string(targetPath[0]) != "/" {
					logrus.WithFields(logrus.Fields{"module": "checkgen", "event": "ProduceChecks", "targetPath": targetPort}).Warn("Invalid target path")
					return checks
				}

				target := &checker.Target{
					Name: *lb.LoadBalancerName,
					Type: "elb",
					Id:   *lb.LoadBalancerName, // TODO, workaround bartnet/beavis issue.  fail on null id. any name ok?
				}

				templatevars := map[string]interface{}{
					"protocol_name": targetProtocol,
					"target_name":   target.Name,
				}

				httpcheck := &checker.HttpCheck{
					Name:     checkNameTemplate.Render(templatevars),
					Path:     targetPath,
					Protocol: targetProtocol,
					Port:     int32(targetPortInt),
					Verb:     "GET",
				}

				marshaler := jsonpb.Marshaler{}
				targetjson, _ := marshaler.MarshalToString(target)
				specjson, _ := marshaler.MarshalToString(httpcheck)
				checkJson := fmt.Sprintf("{\"name\":\"%s\", \"target\":%s, \"check_spec\":{\"type_url\":\"HttpCheck\",\"value\":%s},\"interval\": 30}", httpcheck.Name, targetjson, specjson)

				spec, _ := checker.MarshalAny(httpcheck)

				check := checker.Check{
					Interval:  30,
					Target:    target,
					CheckSpec: spec,
				}

				ass := Assertions{
					CheckAssertions: []*Assertion{
						&Assertion{Key: "code", Operand: 200, Relationship: "equal"},
					},
				}

				notis := Notifications{
					CheckNotifications: []*Notification{
						&Notification{Type: "email", Value: awsobj.Owner.Email},
					},
				}
				jsonCheck := &CheckFactoryCheck{Check: check, CheckJson: checkJson, CheckAssertions: ass, CheckNotifications: notis}
				logrus.WithFields(logrus.Fields{"module": "checkgen", "event": "ProduceChecks", "protocol": targetProtocol, "CheckFactoryCheck": jsonCheck}).Info("Generated CheckFactoryCheck.")
				checks <- jsonCheck
			default:
				logrus.WithFields(logrus.Fields{"module": "checkgen", "event": "ProduceChecks", "protocol": targetProtocol}).Warn("Unsupported check protocol")
			}
		}
	}
	close(checks)
	return checks
}
