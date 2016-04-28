package launcher

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"sort"
	"strings"
	"text/template"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/ec2rolecreds"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/cenkalti/backoff"
	etcd "github.com/coreos/etcd/client"
	"github.com/opsee/basic/com"
	"github.com/opsee/keelhaul/bus"
	"golang.org/x/net/context"
	"time"
)

type getBastionConfig struct{}

func (s getBastionConfig) Execute(launch *Launch) {
	response, err := launch.etcd.Get(context.Background(), launch.config.BastionConfigKey, &etcd.GetOptions{
		Recursive: true,
		Sort:      true,
		Quorum:    true,
	})
	if err != nil {
		launch.error(err, &bus.Message{
			Command: commandLaunchBastion,
			Message: "failed fetching bastion config from etcd",
		})
		return
	}

	bastionConfig := &BastionConfig{}
	err = json.Unmarshal([]byte(response.Node.Value), bastionConfig)
	if err != nil {
		launch.error(err, &bus.Message{
			Command: commandLaunchBastion,
			Message: "failed unmarshaling bastion config",
		})
		return
	}

	bastionConfig.ModifiedIndex = response.Node.ModifiedIndex
	launch.bastionConfig = bastionConfig
	launch.event(&bus.Message{
		State:   stateInProgress,
		Command: commandLaunchBastion,
		Message: "generated bastion config",
	})
}

type ImageList []*ec2.Image

func (l ImageList) Len() int           { return len(l) }
func (l ImageList) Swap(i, j int)      { l[i], l[j] = l[j], l[i] }
func (l ImageList) Less(i, j int) bool { return *l[i].CreationDate > *l[j].CreationDate }

type getLatestImageID struct{}

func (s getLatestImageID) Execute(launch *Launch) {
	// We use our own access-key and secret-key here, because for whatever
	// reason, customers can't find our AMIs like this even when they're public.
	creds := credentials.NewChainCredentials(
		[]credentials.Provider{
			&ec2rolecreds.EC2RoleProvider{
				Client: ec2metadata.New(session.New()),
			},
			&credentials.EnvProvider{},
		},
	)

	ec2client := ec2.New(session.New(&aws.Config{
		Credentials: creds,
		MaxRetries:  aws.Int(11),
		Region:      launch.session.Config.Region,
	}))

	tag := launch.ImageTag
	if tag == "" {
		tag = launch.bastionConfig.Tag
	}

	imageOutput, err := ec2client.DescribeImages(&ec2.DescribeImagesInput{
		Owners: []*string{
			aws.String(launch.bastionConfig.OwnerID),
		},
		Filters: []*ec2.Filter{
			{
				Name:   aws.String("tag:release"),
				Values: []*string{aws.String(tag)},
			},
			{
				Name:   aws.String("is-public"),
				Values: []*string{aws.String("true")},
			},
		},
	})

	if err != nil {
		launch.error(err, &bus.Message{
			Command: commandLaunchBastion,
			Message: "failed to get list of bastion images",
		})
		return
	}

	if len(imageOutput.Images) == 0 {
		launch.error(
			fmt.Errorf("No images with ownerID=%s and tag:release=%s found.", launch.bastionConfig.OwnerID, launch.bastionConfig.Tag),
			&bus.Message{
				Command: commandLaunchBastion,
				Message: "failed to get list of bastion images",
			})
		return
	}

	// sort in descending order
	sort.Sort(ImageList(imageOutput.Images))
	launch.ImageID = *imageOutput.Images[0].ImageId
	launch.event(&bus.Message{
		State:   stateInProgress,
		Command: commandLaunchBastion,
		Message: fmt.Sprintf("got latest stable bastion image: %s", launch.ImageID),
	})
}

type createTopic struct{}

func (s createTopic) Execute(launch *Launch) {
	var (
		topic *sns.CreateTopicOutput
		err   error
	)

	err = backoff.Retry(func() error {
		topic, err = launch.snsClient.CreateTopic(&sns.CreateTopicInput{
			Name: aws.String("opsee-stack-" + launch.User.CustomerId),
		})

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
		launch.error(err, &bus.Message{
			Command: commandLaunchBastion,
			Message: "failed creating topic",
		})
	} else {
		launch.createTopicOutput = topic
		launch.event(&bus.Message{
			State:   stateInProgress,
			Command: commandLaunchBastion,
			Message: fmt.Sprintf("created sns topic: %s", *topic.TopicArn),
		})
	}
}

type createQueue struct{}

func (s createQueue) Execute(launch *Launch) {
	var (
		queue *sqs.CreateQueueOutput
		err   error
	)

	err = backoff.Retry(func() error {
		queue, err = launch.sqsClient.CreateQueue(&sqs.CreateQueueInput{
			QueueName: aws.String("opsee-bastion-launch-sqs" + launch.Bastion.ID),
		})

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
		launch.error(err, &bus.Message{
			Command: commandLaunchBastion,
			Message: "failed creating sqs queue",
		})
	} else {
		launch.createQueueOutput = queue
		launch.event(&bus.Message{
			State:   stateInProgress,
			Command: commandLaunchBastion,
			Message: fmt.Sprintf("created sqs queue: %s", *queue.QueueUrl),
		})
	}
}

type getQueueAttributes struct{}

func (s getQueueAttributes) Execute(launch *Launch) {
	var (
		queueAttributes *sqs.GetQueueAttributesOutput
		err             error
	)

	err = backoff.Retry(func() error {
		queueAttributes, err = launch.sqsClient.GetQueueAttributes(&sqs.GetQueueAttributesInput{
			QueueUrl: launch.createQueueOutput.QueueUrl,
			AttributeNames: []*string{
				aws.String("QueueArn"),
			},
		})

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
		launch.error(err, &bus.Message{
			Command: commandLaunchBastion,
			Message: "failed to get sqs queue attributes",
		})
		return
	}

	_, ok := queueAttributes.Attributes["QueueArn"]
	if !ok {
		launch.error(
			fmt.Errorf("no queue ARN found in queue attributes"),
			&bus.Message{
				Command: commandLaunchBastion,
				Message: "failed to get queue attributes",
			},
		)
		return
	}

	launch.getQueueAttributesOutput = queueAttributes
	launch.event(&bus.Message{
		State:   stateInProgress,
		Command: commandLaunchBastion,
		Message: "got sqs queue attributes",
	})
}

const policyStr = `{
    "Version":"2012-10-17",
    "Statement":[
        {
            "Sid":"{{ .policyID }}",
            "Effect":"Allow",
            "Principal":"*",
            "Action":"sqs:SendMessage",
            "Resource":"{{ .queueARN }}",
            "Condition":{
                "ArnEquals":{
                    "aws:SourceArn":"{{ .topicARN }}"
                }
            }
        }
    ]
}`

var policyTmpl = template.Must(template.New("policy").Parse(policyStr))

type setQueueAttributes struct{}

func (s setQueueAttributes) Execute(launch *Launch) {
	buf := bytes.NewBuffer([]byte{})
	err := policyTmpl.Execute(buf, map[string]string{
		"policyID": launch.Bastion.ID,
		"queueARN": *launch.getQueueAttributesOutput.Attributes["QueueArn"],
		"topicARN": *launch.createTopicOutput.TopicArn,
	})

	if err != nil {
		launch.error(err, &bus.Message{
			Command: commandLaunchBastion,
			Message: "failed to generate sqs policy",
		})
		return
	}

	var (
		sqa *sqs.SetQueueAttributesOutput
	)

	err = backoff.Retry(func() error {
		sqa, err = launch.sqsClient.SetQueueAttributes(&sqs.SetQueueAttributesInput{
			QueueUrl: launch.createQueueOutput.QueueUrl,
			Attributes: map[string]*string{
				"Policy": aws.String(buf.String()),
			},
		})

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
		launch.error(err, &bus.Message{
			Command: commandLaunchBastion,
			Message: "failed setting sqs queue attributes",
		})
	} else {
		launch.setQueueAttributesOutput = sqa
		launch.event(&bus.Message{
			State:   stateInProgress,
			Command: commandLaunchBastion,
			Message: "set sqs queue attributes",
		})
	}
}

type subscribe struct{}

func (s subscribe) Execute(launch *Launch) {
	var (
		subscribeOutput *sns.SubscribeOutput
		err             error
	)

	err = backoff.Retry(func() error {
		subscribeOutput, err = launch.snsClient.Subscribe(&sns.SubscribeInput{
			Protocol: aws.String("sqs"),
			TopicArn: launch.createTopicOutput.TopicArn,
			Endpoint: launch.getQueueAttributesOutput.Attributes["QueueArn"],
		})

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
		launch.error(err, &bus.Message{
			Command: commandLaunchBastion,
			Message: "failed subscribing to sns topic",
		})
	} else {
		launch.subscribeOutput = subscribeOutput
		launch.event(&bus.Message{
			State:   stateInProgress,
			Command: commandLaunchBastion,
			Message: "subscribed to sns topic",
		})
	}
}

type createStack struct{}

func (s createStack) Execute(launch *Launch) {
	userdata, err := launch.GenerateUserData()
	if err != nil {
		launch.error(err, &bus.Message{
			Command: commandLaunchBastion,
			Message: "failed to generate bastion userdata",
		})
		return
	}

	var templateFile string
	if launch.ImageTag == "beta" {
		templateFile = "/bastion-beta-cf.template"
	} else {
		templateFile = launch.config.BastionCFTemplate
	}

	templateBytes, err := ioutil.ReadFile(templateFile)
	if err != nil {
		launch.error(err, &bus.Message{
			Command: commandLaunchBastion,
			Message: "failed to read bastion cloudformation template",
		})
		return
	}

	associateIP := "False"
	if launch.Bastion.SubnetRouting == com.RoutingStatePublic {
		associateIP = "True"
	}

	awsAccount := &com.Account{
		CustomerID: launch.User.CustomerId,
	}

	stackParameters := []*cloudformation.Parameter{
		{
			ParameterKey:   aws.String("ImageId"),
			ParameterValue: aws.String(launch.ImageID),
		},
		{
			ParameterKey:   aws.String("InstanceType"),
			ParameterValue: aws.String(launch.Bastion.InstanceType),
		},
		{
			ParameterKey:   aws.String("UserData"),
			ParameterValue: aws.String(base64.StdEncoding.EncodeToString(userdata)),
		},
		{
			ParameterKey:   aws.String("VpcId"),
			ParameterValue: aws.String(launch.Bastion.VPCID),
		},
		{
			ParameterKey:   aws.String("SubnetId"),
			ParameterValue: aws.String(launch.Bastion.SubnetID),
		},
		{
			ParameterKey:   aws.String("AssociatePublicIpAddress"),
			ParameterValue: aws.String(associateIP),
		},
		{
			ParameterKey:   aws.String("CustomerId"),
			ParameterValue: aws.String(launch.User.CustomerId),
		},
		{
			ParameterKey:   aws.String("BastionId"),
			ParameterValue: aws.String(launch.Bastion.ID),
		},
	}

	if launch.ImageTag != "beta" {
		stackParameters = append(stackParameters, &cloudformation.Parameter{
			ParameterKey:   aws.String("OpseeRole"),
			ParameterValue: aws.String(awsAccount.RoleName()),
		})
	}

	var (
		stack *cloudformation.CreateStackOutput
	)

	err = backoff.Retry(func() error {
		stack, err = launch.cloudformationClient.CreateStack(&cloudformation.CreateStackInput{
			StackName:    aws.String("opsee-stack-" + launch.User.CustomerId),
			TemplateBody: aws.String(string(templateBytes)),
			Capabilities: []*string{
				aws.String("CAPABILITY_IAM"),
			},
			Parameters: stackParameters,
			Tags: []*cloudformation.Tag{
				{
					Key:   aws.String("Name"),
					Value: aws.String("Opsee Stack"),
				},
				{
					Key:   aws.String("vendor"),
					Value: aws.String("Opsee"),
				},
				{
					Key:   aws.String("opsee:customer-id"),
					Value: aws.String(launch.User.CustomerId),
				},
			},
			NotificationARNs: []*string{
				launch.createTopicOutput.TopicArn,
			},
		})

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
		launch.error(err, &bus.Message{
			Command: commandLaunchBastion,
			Message: "failed creating cloudformation stack",
		})
	} else {
		launch.createStackOutput = stack
		launch.event(&bus.Message{
			State:   stateInProgress,
			Command: commandLaunchBastion,
			Message: "launched cloudformation stack",
		})
	}
}

type bastionLaunchingState struct{}

func (s bastionLaunchingState) Execute(launch *Launch) {
	err := launch.db.UpdateBastion(launch.Bastion.Launch(*launch.createStackOutput.StackId, launch.ImageID))
	if err != nil {
		launch.error(err, &bus.Message{
			Command: commandLaunchBastion,
			Message: "failed saving bastion object",
		})
	}
}

type consumeSQS struct{}

func (s consumeSQS) Execute(launch *Launch) {
	var (
		state  = "launching"
		reason string
	)

	msgInput := &sqs.ReceiveMessageInput{
		QueueUrl:        launch.createQueueOutput.QueueUrl,
		WaitTimeSeconds: aws.Int64(20),
	}

	for {
		messages, err := launch.sqsClient.ReceiveMessage(msgInput)
		if err != nil {
			launch.error(err, &bus.Message{
				Command: commandLaunchBastion,
				Message: "failed receiving messages from sqs queue",
			})
			return
		}

		for _, message := range messages.Messages {
			msg := make(map[string]interface{})
			err = json.Unmarshal([]byte(*message.Body), &msg)

			if err != nil {
				launch.error(err, &bus.Message{
					Command: commandLaunchBastion,
					Message: "failed decoding message from sqs queue",
				})
				return
			}

			_, err = launch.sqsClient.DeleteMessage(&sqs.DeleteMessageInput{
				QueueUrl:      launch.createQueueOutput.QueueUrl,
				ReceiptHandle: message.ReceiptHandle,
			})

			if err != nil {
				launch.error(err, &bus.Message{
					Command: commandLaunchBastion,
					Message: "failed deleting message from sqs queue",
				})
				return
			}

			m, ok := msg["Message"].(string)
			if !ok {
				continue
			}

			cfMessage := make(cfMessage)
			err = parseCloudFormation(&cfMessage, m)
			if err != nil {
				launch.error(err, &bus.Message{
					Command: commandLaunchBastion,
					Message: "failed parsing cloudformation message",
				})
				return
			}

			state = cfMessage.state("opsee-stack-" + launch.User.CustomerId)
			if state == cfFailed {
				reason, _ = cfMessage["ResourceStatusReason"].(string)
			}

			if state == cfRollback {
				launch.error(
					fmt.Errorf(reason),
					&bus.Message{
						Command:    commandLaunchBastion,
						Message:    "cloudformation failed to launch",
						Attributes: cfMessage,
					},
				)
				return
			}

			if state == cfComplete {
				launch.event(&bus.Message{
					State:      stateInProgress,
					Command:    commandLaunchBastion,
					Message:    "cloudformation stack launch complete",
					Attributes: cfMessage,
				})
				return
			}

			launch.event(&bus.Message{
				State:      stateInProgress,
				Command:    commandLaunchBastion,
				Message:    "launching cloudformation stack",
				Attributes: cfMessage,
			})
		}
	}
}

type cfMessage map[string]interface{}

const (
	cfLaunching = "launching"
	cfFailed    = "failed"
	cfRollback  = "rollback"
	cfComplete  = "complete"
)

func (cf cfMessage) state(stackName string) string {
	var (
		status string
		typ    string
		id     string
	)

	status, _ = cf["ResourceStatus"].(string)
	typ, _ = cf["ResourceType"].(string)
	id, _ = cf["LogicalResourceId"].(string)

	if status == "CREATE_FAILED" {
		return cfFailed
	}

	if typ == "AWS::CloudFormation::Stack" && id == stackName {
		switch status {
		case "CREATE_COMPLETE":
			return cfComplete
		case "ROLLBACK_COMPLETE":
			return cfRollback
		}
	}

	return cfLaunching
}

func parseCloudFormation(cf *cfMessage, message string) error {
	lines := strings.Split(message, "\n")
	for _, line := range lines {
		fields := strings.SplitN(line, "=", 2)
		if len(fields) != 2 {
			continue
		}

		(*cf)[fields[0]] = strings.Trim(fields[1], "'")
	}

	if len(*cf) == 0 {
		return fmt.Errorf("cloudformation message not parsed: %#v", *cf)
	}

	return nil
}

type bastionActiveState struct{}

func (s bastionActiveState) Execute(launch *Launch) {
	var (
		instanceID *string
		groupID    *string
		nextToken  *string
	)

	for {
		stackResourcesOutput, err := launch.cloudformationClient.ListStackResources(&cloudformation.ListStackResourcesInput{
			StackName: aws.String("opsee-stack-" + launch.User.CustomerId),
			NextToken: nextToken,
		})

		if err != nil {
			launch.error(err, &bus.Message{
				Command: commandLaunchBastion,
				Message: "failed retrieving launched stack info",
			})
			return
		}

		for _, s := range stackResourcesOutput.StackResourceSummaries {
			switch *s.ResourceType {
			case "AWS::EC2::SecurityGroup":
				groupID = s.PhysicalResourceId
			}
		}

		nextToken = stackResourcesOutput.NextToken
		if nextToken == nil {
			break
		}
	}

	instanceID = aws.String("not used")
	err := launch.db.UpdateBastion(launch.Bastion.Activate(*instanceID, *groupID))
	if err != nil {
		launch.error(err, &bus.Message{
			Command: commandLaunchBastion,
			Message: "failed saving bastion object",
		})
		return
	}

	launch.event(&bus.Message{
		State:   stateComplete,
		Command: commandLaunchBastion,
		Message: "bastion activation complete",
	})
}
