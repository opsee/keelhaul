package checkgen

import (
	"encoding/json"
	"github.com/opsee/keelhaul/checker"
	"github.com/sirupsen/logrus"
)

type ResponseHandler interface {
	HandleResponse(response *RequestPoolResponse) *RequestPoolResponse
}

type DefaultResponseHandler struct{}

func (defaultResponseHandler *DefaultResponseHandler) HandleResponse(response *RequestPoolResponse) *RequestPoolResponse {
	return response
}

type CreateCheckResponseHandler struct {
	CheckAssertions     Assertions
	CheckNotifications  Notifications
	CheckRequestFactory *CheckRequestFactory
}

func (createCheckResponseHandler *CreateCheckResponseHandler) HandleResponse(response *RequestPoolResponse) *RequestPoolResponse {
	r := response.Response
	check := &checker.Check{}
	json.NewDecoder(r.Body).Decode(check)

	// build an assertion request and do it
	createCheckResponseHandler.CheckAssertions.CheckId = check.Id
	req, err := createCheckResponseHandler.CheckRequestFactory.BuildAssertionsRequest(&createCheckResponseHandler.CheckAssertions)
	if err != nil {
		logrus.WithFields(logrus.Fields{"module": "checkgen", "event": "HandleResponse", "RetVal": req, "Error": err}).Error("Couldn't BuildAssertionsRequest(", createCheckResponseHandler.CheckAssertions, ")")
		return &RequestPoolResponse{Response: nil, Err: err}
	}

	res := req.DoRequest()
	if res.Err != nil {
		return res
	}

	// assertion passed so build a notifications request and do it
	out, err := createCheckResponseHandler.CheckNotifications.MarshalToString()
	if err != nil {
		logrus.WithFields(logrus.Fields{"module": "checkgen", "event": "HandleResponse", "RetVal": out, "Error": err}).Error("Couldn't CheckNotifications.MarshalToString()")
		return &RequestPoolResponse{Response: nil, Err: err}
	}

	createCheckResponseHandler.CheckNotifications.CheckId = check.Id
	req, err = createCheckResponseHandler.CheckRequestFactory.BuildNotificationsRequest(&createCheckResponseHandler.CheckNotifications)
	if err != nil {
		logrus.WithFields(logrus.Fields{"module": "checkgen", "event": "HandleResponse", "RetVal": req, "Error": err}).Error("Couldn't BuildNotificationsRequest(", createCheckResponseHandler.CheckNotifications, ")")
		return &RequestPoolResponse{Response: nil, Err: err}
	}

	return req.DoRequest()
}
