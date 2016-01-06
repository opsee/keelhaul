package checkgen

import (
	"net/http"
	"sync"
	"sync/atomic"

	"github.com/sirupsen/logrus"
)

type RequestPoolRequest struct {
	Id              string
	Request         *http.Request
	ResponseHandler ResponseHandler
}

func (requestPoolRequest *RequestPoolRequest) DoRequest() *RequestPoolResponse {
	client := &http.Client{}
	logrus.WithFields(logrus.Fields{"module": "checkgen", "event": "DoRequest", "Request": requestPoolRequest.Request}).Info("Do it")
	resp, err := client.Do(requestPoolRequest.Request)
	if resp != nil {
		defer resp.Body.Close()
	}
	logrus.WithFields(logrus.Fields{"module": "checkgen", "event": "DoRequest", "Response": resp, "Error": err}).Info("Server returned response.")

	if err != nil {
		return &RequestPoolResponse{Err: err, Response: resp}
	} else {
		return requestPoolRequest.ResponseHandler.HandleResponse(&RequestPoolResponse{Err: nil, Response: resp})
	}
}

type RequestPoolResponse struct {
	Err      error
	Response *http.Response
}

type RequestPool struct {
	RequestsChan       chan RequestPoolRequest
	SuccessfulRequests int
}

func NewRequestPool() *RequestPool {
	return &RequestPool{
		RequestsChan:       make(chan RequestPoolRequest, 256),
		SuccessfulRequests: 0,
	}
}

func (requestPool *RequestPool) DrainRequests(send bool) int {
	var checks uint32
	if send {
		var wg sync.WaitGroup
		wg.Add(len(requestPool.RequestsChan))
		for r := range requestPool.RequestsChan {
			go func(r RequestPoolRequest) {
				logrus.Info("Sending request ", r)
				defer wg.Done()
				response := r.DoRequest()
				if response.Err == nil {
					atomic.AddUint32(&checks, 1)
				}
			}(r)
		}
		wg.Wait()
	}

	requestPool.SuccessfulRequests = int(checks)
	return int(checks)
}
