package autocheck

import (
	"github.com/opsee/basic/clients/bartnet"
	"github.com/opsee/basic/schema"
)

type bartnetSink struct {
	bartnetClient bartnet.Client
	hugsClient    *hugsClient
	user          *schema.User
}

func NewBartnetSink(bartnetEndpoint, hugsEndpoint string, user *schema.User) *bartnetSink {
	return &bartnetSink{
		bartnetClient: bartnet.New(bartnetEndpoint),
		hugsClient:    newHugsClient(hugsEndpoint),
		user:          user,
	}
}

func (s *bartnetSink) Send(check *schema.Check) error {
	checkResp, err := s.bartnetClient.CreateCheck(s.user, check)
	if err != nil {
		return err
	}

	return s.hugsClient.CreateNotifications(s.user, &NotificationRequest{
		CheckId: checkResp.Id,
		Notifications: []*Notification{
			{
				Type:  "email",
				Value: s.user.Email,
			},
		},
	})
}
