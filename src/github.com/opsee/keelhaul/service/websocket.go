package service

import (
	"encoding/base64"
	"encoding/json"
	"github.com/opsee/keelhaul/com"
	"github.com/opsee/vaper"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/websocket"
	"time"
)

func (s *service) websocketHandler() func(ws *websocket.Conn) {
	return func(ws *websocket.Conn) {
		defer ws.Close()

		message := &com.Message{}
		err := websocket.JSON.Receive(ws, message)
		if err != nil {
			log.WithError(err).Error("problem decoding websocket subscription message")
			return
		}

		switch message.Command {
		case "authenticate":
			token, ok := message.Attributes["token"].(string)
			if !ok {
				log.Errorf("no token send in authenticate request: %#v", message.Attributes)
				return
			}

			decodedToken, err := vaper.Unmarshal(token)
			if err != nil {
				log.Errorf("couldn't unmarshal token: %s", token)
				return
			}

			user := &com.User{}
			err = decodedToken.Reify(user)
			if err != nil {
				log.WithError(err).Error("failed to decode bearer user token")
				return
			}

			log.WithFields(log.Fields{
				"customer-id": user.CustomerID,
				"user-id":     user.ID,
			}).Info("authenticated user via websocket")

			sub := s.bus.Subscribe(user)
			defer s.bus.Unsubscribe(sub)

			heartbeat := time.NewTicker(10 * time.Second)
			defer heartbeat.Stop()

			bastionBeat := time.NewTicker(30 * time.Second)
			defer bastionBeat.Stop()

			// initial messages
			err = websocket.JSON.Send(ws, s.bastionMessage(user))
			if err != nil {
				log.WithError(err).Error("can't sent to websocket")
				return
			}

			for {
				select {
				case <-bastionBeat.C:
					err = websocket.JSON.Send(ws, s.bastionMessage(user))
					if err != nil {
						log.WithError(err).Error("can't sent to websocket")
						return
					}

				case bmsg := <-sub.Chan:
					err = websocket.JSON.Send(ws, bmsg)
					if err != nil {
						log.WithError(err).Error("can't sent to websocket")
						return
					}

				case t := <-heartbeat.C:
					err = websocket.JSON.Send(ws, &com.Message{
						Command:    "heartbeat",
						State:      "ok",
						CustomerID: user.CustomerID,
						Attributes: map[string]interface{}{
							"time": t,
						},
					})
					if err != nil {
						log.WithError(err).Error("can't sent to websocket")
						return
					}
				}
			}
		}
	}
}

func (s *service) bastionMessage(user *com.User) *com.Message {
	var msg *com.Message

	bastionsResponse, err := s.ListBastions(user, &ListBastionsRequest{
		State: []string{"active"},
	})

	if err != nil {
		log.WithError(err).Error("failed listing bastions")

		msg = &com.Message{
			Command:    "bastions",
			State:      "error",
			CustomerID: user.CustomerID,
			Message:    "error listing bastions",
		}
	} else {
		msg = &com.Message{
			Command:    "bastions",
			State:      "ok",
			CustomerID: user.CustomerID,
			Attributes: map[string]interface{}{
				"bastions": bastionsResponse.Bastions,
			},
		}
	}

	return msg
}

func decodeBasicToken(token string, user *com.User) error {
	jsonblob, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		return err
	}

	err = json.Unmarshal(jsonblob, user)
	if err != nil {
		return err
	}

	err = user.Validate()
	if err != nil {
		return err
	}

	return nil
}
