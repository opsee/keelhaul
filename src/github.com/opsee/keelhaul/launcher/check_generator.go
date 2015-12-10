package launcher

import (
	"bytes"
	"fmt"
	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/golang/protobuf/proto"
	"github.com/opsee/keelhaul/auth"
	"github.com/opsee/keelhaul/checker"
	"github.com/opsee/keelhaul/com"
	"github.com/opsee/keelhaul/util"
	"math/rand"
	"net/http"
	"time"
)

type AWSType string

const (
	ELB_DESCRIPTION = "LoadBalancerDescription"
)

func (awstype AWSType) String() string {
	switch awstype {
	case ELB_DESCRIPTION:
		return "LoadBalancerDescription"
	}

	return "UKNOWN_AWS_TYPE"
}

func GetAWSTypeByString(awstype string) (AWSType, error) {
	switch awstype {
	case "LoadBalancerDescription":
		return ELB_DESCRIPTION, nil
	}

	return UNKNOWN_AWS_TYPE, fmt.Errorf("Unknown aws type: %s", tokentype)
}

type AWSObjectType int

const (
	PROTO AWSObjectType = iota
	JSON  AWSObjectType = iota
)

type AWSObject struct {
	Type       AWSType
	ObjectType AWSObjectType
	Object     []byte
}

type CreateCheckRequest struct {
	Request *http.Request
}

type CreateCheckResponse struct {
	Err           error
	ResponseValue *http.Response
}

type RequestPool struct {
	Requests  map[string]*CreateCheckRequest
	Responses map[string]*Response
}

func (requestPool *RequestPool) AddRequest(id string, req *http.Request) (*http.Request, error) {
	if req == nil {
		return nil, fmt.Errorf("Nil pointer to http.Request object")
	}
	requestPool.Requests[id] = req

	return req, nil
}

func (requestPool *RequestPool) DrainRequests(send bool) *map[string]*CheckRequestResponse {
	for k := range requestPool.Requests {
		if send {
			go func() {
				client := &http.Client{}
				resp, err = client.Do(requestPool.Requests[k].Request)
				if err != nil {
					Responses[k] = &CheckRequestResponse{Err: err.Err, ResponseValue: resp}
					return
				} else {
					Responses[k] = &CheckRequestResponse{Err: nil, ResponseValue: resp}
					delete(requestPool.Requests, k)
				}
			}()
		} else {
			delete(requestPool.Requests, k)
		}
	}

	return req, nil
}

type CheckRequestFactory struct {
	User                *com.User
	ConcreteRequestPool *RequestPool
	ConcreteFactories   map[AWSType]ChecksFactory
}

func (checkRequestFactory *CheckRequestFactory) ProduceRequest(obj *AWSObject) (*http.Request, error) {
	if obj == nil {
		return nil, fmt.Errorf("Nil pointer to AWSObject.  Cannot Produce CheckRequest")
	}
	if concreteFactory, ok := Factories[*awsobj.Type]; ok {
		checksChan := concreteFactory.ProduceCheckRequest(awsobj), nil

		for _, check := range checksChan {
			select {
			case newCheck := <-checksChan:
				createCheckRequest := checkRequestFactory.buildRequest(newCheck)
				checkRequestFactory.ConcreteRequestPool.AddRequest(util.RandomString(8, "asdfkjhqwerpoiu12340987"))
			default:
				break
			}
		}
		requestPool.addRequest()

	}
	return nil, fmt.Errorf("No suitable factory found to produce %s", *AWSObject.Type)
}

func (checkRequestFactory *CheckRequestFactory) buildRequest(check *checker.Check) *CreateCheckRequest {
	pb, err := proto.Marshal(check)
	if err != nil {
		return nil, err
	}

	auth := &auth.BastionAuthTokenRequest{
		TokenType:      tokenType,
		CustomerEmail:  checkRequestFactory.User.email,
		CustomerID:     checkRequestFactory.User.Id,
		TargetEndpoint: "https://bartnet.in.opsee.com/checks",
	}

	if token, err := cache.GetToken(request); err != nil || token == nil {
		logrus.WithFields(logrus.Fields{"service": "checker", "Error": err.Error()}).Fatal("Error check request creation")
		return nil, err
	} else {
		theauth, header := token.AuthHeader()
		logrus.WithFields(logrus.Fields{"service": "checker", "Auth header:": theauth + " " + header}).Info("Creating check request.")

		req, err := http.NewRequest("POST", request.TargetEndpoint, nil)
		if err != nil {
			logrus.WithFields(logrus.Fields{"service": "checker", "error": err, "response": resp}).Warn("Couldn't sychronize checks")
			return nil, err
		} else {
			req.Header.Set("Content-Type", "application/x-protobuf")
			req.Header.Set(theauth, header)
			logrus.WithFields(logrus.Fields{"service": "checker", "Auth header:": theauth + " " + header}).Info("Created check request.")
			return &CreateCheckRequest{Request: req}, nil
		}
	}
	return nil, fmt.Errorf("Check Request creation failed.")
}

type ChecksFactory interface {
	ProduceCheckRequests(awsobj *AWSObject) chan *checker.Check
}

type ELBCheckFactory struct{}

func (elbFactory *ELBCheckFactory) ProduceCheckRequests(awsobj *AWSObject) chan *checker.Check {

	lb := elb.LoadBalancerDescription{}

	// unmarshal awsobj based on objtype (prombly json)

	// get listeners

	requests := make(chan checker.Check, len(lb.ListenerDescriptions))

	for i, listenerDescription := range lb.ListenerDescriptions {
		switch listenerDescription.Protocol {

		case "HTTP":
			target := &checker_proto.Target{
				Name: lb.LoadBalancerName,
				Type: "elb",
			}

			httpcheck := &checker_proto.HttpCheck{
				Name:     util.RandomString(5, "asdfglkhpoiuqwerAEMWX"),
				Path:     "/",
				Protocol: listenDescription.Protocol,
				Port:     listenerDesription.InstancePort,
				Verb:     "GET",
				Headers:  []*Header{},
			}

			spec, _ := checker_proto.MarshalAny

			check = &Check{
				Id:        "",
				Interval:  30,
				Target:    target,
				CheckSpec: spec,
			}

			requests <- check
		}
	}
	return requests
}
