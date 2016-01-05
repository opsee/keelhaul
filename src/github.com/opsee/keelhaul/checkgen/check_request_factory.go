package checkgen

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"github.com/opsee/keelhaul/com"
	"github.com/opsee/keelhaul/config"
	"github.com/opsee/keelhaul/util"
	"github.com/sirupsen/logrus"
	"net/http"
)

type CheckRequestFactory struct {
	Config           *config.Config
	User             *com.User
	CheckRequestPool *RequestPool
	CheckFactories   map[string]ChecksFactory
}

func NewCheckRequestFactoryWithConfig(config *config.Config, user *com.User) *CheckRequestFactory {
	return &CheckRequestFactory{
		CheckFactories: map[string]ChecksFactory{
			"LoadBalancerDescription": &ELBCheckFactory{},
		},
		CheckRequestPool: NewRequestPool(),
		Config:           config,
		User:             user,
	}
}

func NewCheckRequestFactory() *CheckRequestFactory {
	return &CheckRequestFactory{
		CheckFactories: map[string]ChecksFactory{
			"LoadBalancerDescription": &ELBCheckFactory{},
		},
		CheckRequestPool: NewRequestPool(),
		Config:           &config.Config{},
		User:             &com.User{},
	}
}

func (checkRequestFactory *CheckRequestFactory) ProduceCheckRequests(awsobj *com.AWSObject) error {
	if awsobj == nil {
		return fmt.Errorf("Nil pointer to AWSObject.  Cannot Produce CheckRequest")
	}
	if awsobj.Owner == nil {
		awsobj.Owner = checkRequestFactory.User
	}
	if checkFactory, ok := checkRequestFactory.CheckFactories[awsobj.Type]; ok {
		checksChan := checkFactory.ProduceChecks(awsobj)
		for newCheck := range checksChan {
			createCheckRequest, err := checkRequestFactory.buildCheckRequest(newCheck)
			if err != nil {
				continue
			}
			logrus.Info("Write to channel")
			select {
			case checkRequestFactory.CheckRequestPool.RequestsChan <- *createCheckRequest:
				logrus.Info("Wrote check to RequestPool")
			default:
				logrus.Warn("RequestPool full.")
			}
			logrus.Info("/Write to channel")
		}
	}
	return fmt.Errorf("No suitable factory found to produce %s", awsobj.Type)
}

func (checkRequestFactory *CheckRequestFactory) BuildNotificationsRequest(notifications *Notifications) (*RequestPoolRequest, error) {
	notificationsJson, err := notifications.MarshalToString()
	if err != nil {
		return nil, err
	}

	req, err := checkRequestFactory.getAuthenticatedRequest(fmt.Sprintf("%s/notifications", checkRequestFactory.Config.BeavisEndpoint), []byte(notificationsJson))
	if err != nil {
		return nil, err
	}

	return &RequestPoolRequest{Id: util.RandomString(8), Request: req, ResponseHandler: &DefaultResponseHandler{}}, nil
}

func (checkRequestFactory *CheckRequestFactory) BuildAssertionsRequest(assertions *Assertions) (*RequestPoolRequest, error) {
	assertionsJson, err := assertions.MarshalToString()
	if err != nil {
		return nil, err
	}

	req, err := checkRequestFactory.getAuthenticatedRequest(fmt.Sprintf("%s/assertions", checkRequestFactory.Config.BeavisEndpoint), []byte(assertionsJson))
	if err != nil {
		return nil, err
	}

	return &RequestPoolRequest{Id: util.RandomString(8), Request: req, ResponseHandler: &DefaultResponseHandler{}}, nil
}

func (checkRequestFactory *CheckRequestFactory) buildCheckRequest(check *CheckFactoryCheck) (*RequestPoolRequest, error) {
	pb := check.CheckJson

	req, err := checkRequestFactory.getAuthenticatedRequest(fmt.Sprintf("%s/checks", checkRequestFactory.Config.BartnetEndpoint), []byte(pb))
	if err != nil {
		return nil, err
	}

	responseHandler := &CreateCheckResponseHandler{
		CheckAssertions:     check.CheckAssertions,
		CheckNotifications:  check.CheckNotifications,
		CheckRequestFactory: checkRequestFactory,
	}

	return &RequestPoolRequest{Id: util.RandomString(8), Request: req, ResponseHandler: responseHandler}, nil
}

func (checkRequestFactory *CheckRequestFactory) getAuthenticatedRequest(target string, data []byte) (*http.Request, error) {
	dataBuf := bytes.NewBuffer(data)
	req, err := http.NewRequest("POST", target, dataBuf)
	if err != nil {
		return nil, err
	}

	userdata := fmt.Sprintf("{\"email\":\"%s\",\"customer_id\":\"%s\"}", checkRequestFactory.User.Email, checkRequestFactory.User.CustomerID)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(userdata)))
	return req, nil
}
