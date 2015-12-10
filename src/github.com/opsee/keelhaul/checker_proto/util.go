package checker_proto

import (
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/golang/protobuf/proto"
	"reflect"
)

var registry = make(map[string]reflect.Type)

func init() {
	registry["HttpCheck"] = reflect.TypeOf(HttpCheck{})
}

func MarshalAny(i interface{}) (*Any, error) {
	msg, ok := i.(proto.Message)
	if !ok {
		err := fmt.Errorf("Unable to convert to proto.Message: %v", i)
		logrus.WithFields(logrus.Fields{"service": "checker", "event": "marshalling error"}).Error(err.Error())
		return nil, err
	}
	bytes, err := proto.Marshal(msg)

	if err != nil {
		logrus.WithFields(logrus.Fields{"service": "checker", "event": "marshalling error"}).Error(err.Error())
		return nil, err
	}

	return &Any{
		TypeUrl: reflect.ValueOf(i).Elem().Type().Name(),
		Value:   bytes,
	}, nil
}

func UnmarshalAny(any *Any) (interface{}, error) {
	class := any.TypeUrl
	bytes := any.Value

	instance := reflect.New(registry[class]).Interface()
	err := proto.Unmarshal(bytes, instance.(proto.Message))
	if err != nil {
		logrus.WithFields(logrus.Fields{"service": "checker", "event": "unmarshall returned error", "error": "couldn't unmarshall *Any"}).Error(err.Error())
		return nil, err
	}
	logrus.WithFields(logrus.Fields{"service": "checker", "event": "unmarshal successful"}).Info("unmarshaled Any to: ", instance)

	return instance, nil
}
