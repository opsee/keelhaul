package checkgen

import (
	"encoding/json"
	"github.com/sirupsen/logrus"
)

// TODO, this could get out of sync with the API.
//{"check-id":"14sWFvKfdEeOS3ZlxF1Oej","notifications":[{"key":"code","operand":200,"relationship":"equal"}]}
type Notifications struct {
	CheckId            string          `json:"check-id"`
	CheckNotifications []*Notification `json:"notifications,array"`
}

func (notifications *Notifications) MarshalToString() (string, error) {
	out, err := json.Marshal(*notifications)

	if err != nil {
		logrus.WithFields(logrus.Fields{"module": "checkgen", "event": "MarshalToString", "Error": err}).Error("Couldn't unmarshal json for notification.")
		return string(out), err
	}

	return string(out), err
}

type Notification struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}
