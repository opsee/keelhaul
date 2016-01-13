package launcher

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/cenkalti/backoff"
	"github.com/opsee/basic/com"
	"time"
)

type putRole struct{}

func (s putRole) Execute(launch *Launch) {
	credValue, err := launch.session.Config.Credentials.Get()
	if err != nil {
		launch.error(
			err,
			&com.Message{
				Command: commandLaunchBastion,
				Message: "error fetching credentials from session",
			},
		)

		return
	}

	stscreds, err := launch.spanx.PutRole(launch.User, credValue.AccessKeyID, credValue.SecretAccessKey)
	if err != nil {
		launch.error(
			err,
			&com.Message{
				Command: commandLaunchBastion,
				Message: "error provisioning opsee role",
			},
		)

		return
	}

	// test to see if our role is actually usable by doing something with it
	ec2Client := ec2.New(
		launch.session.Copy(&aws.Config{
			Credentials: credentials.NewStaticCredentials(
				stscreds.AccessKeyID,
				stscreds.SecretAccessKey,
				stscreds.SessionToken,
			),
		}),
	)

	err = backoff.Retry(func() error {
		_, err = ec2Client.DescribeVpcs(nil)
		if err != nil {
			return err
		}

		return nil

	}, &backoff.ExponentialBackOff{
		InitialInterval:     100 * time.Millisecond,
		RandomizationFactor: 0.5,
		Multiplier:          1.5,
		MaxInterval:         time.Second,
		MaxElapsedTime:      5 * time.Second,
		Clock:               &systemClock{},
	})

	if err != nil {
		launch.error(
			err,
			&com.Message{
				Command: commandLaunchBastion,
				Message: "error using new opsee provisioned role",
			},
		)

		return
	}

	launch.event(&com.Message{
		State:   stateInProgress,
		Command: commandLaunchBastion,
		Message: "created global opsee role",
	})
}

type systemClock struct{}

func (s *systemClock) Now() time.Time {
	return time.Now()
}
