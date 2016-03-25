package launcher

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/cenkalti/backoff"
	opseeawscredentials "github.com/opsee/basic/schema/aws/credentials"
	"github.com/opsee/basic/service"
	"github.com/opsee/keelhaul/bus"
	"golang.org/x/net/context"
	"time"
)

type putRole struct{}

func (s putRole) Execute(launch *Launch) {
	credValue, err := launch.session.Config.Credentials.Get()
	if err != nil {
		launch.error(
			err,
			&bus.Message{
				Command: commandLaunchBastion,
				Message: "error fetching credentials from session",
			},
		)

		return
	}

	stscreds, err := launch.spanx.PutRole(context.Background(), &service.PutRoleRequest{
		User: launch.User,
		Credentials: &opseeawscredentials.Value{
			AccessKeyID:     aws.String(credValue.AccessKeyID),
			SecretAccessKey: aws.String(credValue.SecretAccessKey),
		},
	})
	if err != nil {
		launch.error(
			err,
			&bus.Message{
				Command: commandLaunchBastion,
				Message: "error provisioning opsee role",
			},
		)

		return
	}

	// now graduate to our new session that uses the opsee role
	launch.session = launch.session.Copy(&aws.Config{
		Credentials: credentials.NewStaticCredentials(
			stscreds.Credentials.GetAccessKeyID(),
			stscreds.Credentials.GetSecretAccessKey(),
			stscreds.Credentials.GetSessionToken(),
		),
	})

	launch.sqsClient = sqs.New(launch.session)
	launch.snsClient = sns.New(launch.session)
	launch.cloudformationClient = cloudformation.New(launch.session)

	// test to see if our role is actually usable by doing something with it
	ec2Client := ec2.New(launch.session)

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
			&bus.Message{
				Command: commandLaunchBastion,
				Message: "error using new opsee provisioned role",
			},
		)

		return
	}

	launch.event(&bus.Message{
		State:   stateInProgress,
		Command: commandLaunchBastion,
		Message: "created global opsee role",
	})
}

