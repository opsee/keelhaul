package autocheck

import (
	"fmt"
	"github.com/opsee/basic/clients/bartnet"
	"github.com/opsee/basic/clients/hugs"
	"github.com/opsee/basic/schema"
)

type bartnetSink struct {
	bartnetClient bartnet.Client
	hugsClient    hugs.Client
	user          *schema.User
	defaultNotifs []*hugs.Notification
}

func NewBartnetSink(bartnetEndpoint, hugsEndpoint string, user *schema.User) *bartnetSink {
	return &bartnetSink{
		bartnetClient: bartnet.New(bartnetEndpoint),
		hugsClient:    hugs.New(hugsEndpoint),
		user:          user,
	}
}

func (s *bartnetSink) Send(check *schema.Check) error {
	checkResp, err := s.bartnetClient.CreateCheck(s.user, check)
	if err != nil {
		return err
	}

	if checkResp.Id == "" {
		return fmt.Errorf("error getting check id from bartnet %#v", checkResp)
	}

	if s.defaultNotifs == nil {
		notifs, err := s.hugsClient.ListNotificationsDefault(s.user)
		if err != nil || len(notifs) == 0 {
			// just default to the user's email
			notifs = []*hugs.Notification{
				{
					Type:  "email",
					Value: s.user.Email,
				},
			}
		}

		s.defaultNotifs = notifs
	}

	return s.hugsClient.CreateNotifications(s.user, &hugs.NotificationRequest{
		CheckId:       checkResp.Id,
		Notifications: s.defaultNotifs,
	})
}
