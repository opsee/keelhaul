package checkgen

import (
	"encoding/json"
	"github.com/sirupsen/logrus"
)

// TODO, this could get out of sync with the API.
//{"check-id":"14sWFvKfdEeOS3ZlxF1Oej","assertions":[{"key":"code","operand":200,"relationship":"equal"}]}
type Assertions struct {
	CheckId         string       `json:"check-id"`
	CheckAssertions []*Assertion `json:"assertions,array"`
}

func (assertions *Assertions) MarshalToString() (string, error) {
	out, err := json.Marshal(*assertions)

	if err != nil {
		logrus.WithFields(logrus.Fields{"module": "checkgen", "event": "MarshalToString", "Error": err}).Error("Couldn't unmarshal assertion json.")
		return string(out), err
	}

	return string(out), err
}

type Assertion struct {
	Key          string `json:"key"`
	Operand      int    `json:"operand"`
	Relationship string `json:"relationship"`
}
