package checkgen

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/opsee/awscan"
	"github.com/opsee/keelhaul/com"
	"github.com/sirupsen/logrus"
	"os"
	"reflect"
	"testing"
)

type vpcDiscovery struct{}

const (
	instanceErrorThreshold = 0.3
)

func TestCanRunLocally(t *testing.T) {
	var (
		creds = credentials.NewStaticCredentials(
			os.Getenv("AWS_ACCESS_KEY_ID"),
			os.Getenv("AWS_SECRET_KEY"),
			"",
		)
		sess = session.New(&aws.Config{
			Credentials: creds,
			Region:      aws.String("us-west-1"),
			MaxRetries:  aws.Int(11),
		})

		disco               = awscan.NewDiscoverer(awscan.NewScanner(sess, "vpc-79b1491c"))
		checkRequestFactory = NewCheckRequestFactory()
	)

	checkRequestFactory.Config.BeavisEndpoint = os.Getenv("BEAVIS_ENDPOINT")
	checkRequestFactory.Config.BartnetEndpoint = os.Getenv("BARTNET_ENDPOINT")
	checkRequestFactory.User.Email = os.Getenv("CUSTOMER_EMAIL")
	checkRequestFactory.User.CustomerID = os.Getenv("CUSTOMER_ID")

	for event := range disco.Discover() {
		if event.Err == nil {
			messageType := reflect.ValueOf(event.Result).Elem().Type().Name()

			switch messageType {
			case awscan.LoadBalancerType:
				logrus.Print("found elb")
				checkRequestFactory.ProduceCheckRequests(&com.AWSObject{Type: messageType, Object: event.Result, Owner: &com.User{Email: os.Getenv("CUSTOMER_EMAIL"), CustomerID: os.Getenv("CUSTOMER_ID")}})
			}
		} else {
			logrus.Print(event.Err)
		}
		logrus.Print("processed event")
	}
	logrus.Print("draining request pool (send requests to bartnet)")
	close(checkRequestFactory.CheckRequestPool.RequestsChan)
	checkRequestFactory.CheckRequestPool.DrainRequests(true)
	logrus.Print("sent ", checkRequestFactory.CheckRequestPool.SuccessfulRequests, " requests succesfully.")
}
