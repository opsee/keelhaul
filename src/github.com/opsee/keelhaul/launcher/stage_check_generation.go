package launcher

import (
	"bytes"
	"fmt"
	//"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/elb"
	//"github.com/aws/aws-sdk-go/service/rds"
	"github.com/golang/protobuf/proto/jsonpb"
	"github.com/opsee/keelhaul/checker_proto"
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

type Response struct {
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

// todo, make sure drain is complete, actually use like a pool?
func (requestPool *RequestPool) DrainRequests(send bool) *map[string]*Response {
	for k := range Requests {
		if send {
			go func() {
				client := &http.Client{}
				resp, err = client.Do(Requests[k])
				if err != nil {
					Responses[k] = &Response{Err: err.Err, ResponseValue: resp}
					return
				} else {
					Responses[k] = &Response{Err: nil, ResponseValue: resp}
					delete(Requests, k)
				}
			}()
		} else {
			delete(Requests, k)
		}
	}

	return req, nil
}

type RequestFactory struct {
	ConcreteRequestPool *RequestPool
	ConcreteFactories   map[AWSType]CheckRequestFactory
}

func (checkRequestFactory *RequestFactory) ProduceRequest(obj *AWSObject) (*http.Request, error) {
	if obj == nil {
		return nil, fmt.Errorf("Nil pointer to AWSObject.  Cannot Produce CheckRequest")
	}
	if concreteFactory, ok := Factories[*awsobj.Type]; ok {
		request := concreteFactory.ProduceCheckRequest(awsobj), nil
		requestPool.addRequest(util.RandomString(10, "asdfPOIUqwerzxcv"), request)
	}
	return nil, fmt.Errorf("No suitable factory found to produce %s", *AWSObject.Type)
}

func (checkRequestFactory *RequestFactory) ProduceRequest(awsobj *AWSObject) (*http.Request, error) {
	if awsobj == nil {
		return nil, fmt.Errorf("Nil pointer to AWSObject.  Cannot Produce CheckRequest")
	}
	if concreteFactory, ok := Factories[*awsobj.Type]; ok {
		request := concreteFactory.ProduceCheckRequest(awsobj), nil
		requestPool.addRequest(util.RandomString(10, "asdfPOIUqwerzxcv"), request)
	}
	return nil, fmt.Errorf("No suitable factory found to produce %s", *AWSObject.Type)
}

type CheckRequestFactory interface {
	ProduceCheckRequest(awsobj *AWSObject) (*http.Request, error)
}

type ELBCheckFactory struct{}

func (elbFactory *ELBCheckFactory) ProduceCheckRequest(awsobj *AWSObject) []*CreateCheckRequest {

	lb := elb.LoadBalancerDescription{}

	// unmarshal awsobj based on objtype (prombly json)

	// get listeners

	requests := make([]CreateCheckRequest, len(lb.ListenerDescriptions))

	for listenerDescription := range lb.ListenerDescriptions {
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

		}
	}
}
