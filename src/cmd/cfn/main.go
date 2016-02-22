package main

import (
	"database/sql"
	"encoding/base64"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	// "os/exec"
	"sort"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	// "github.com/aws/aws-sdk-go/aws/awsutil"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/ec2rolecreds"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/opsee/basic/com"
	"github.com/opsee/basic/schema"
	"github.com/opsee/keelhaul/launcher"
)

var (
	config               *launcher.BastionConfig
	bastion              *com.Bastion
	user                 *schema.User
	userId               int
	templatePath         string
	region               string
	public               string
	doDelete             bool
	cloudformationClient *cloudformation.CloudFormation
	role                 string
	creds                = credentials.NewChainCredentials(
		[]credentials.Provider{
			&ec2rolecreds.EC2RoleProvider{
				Client: ec2metadata.New(session.New()),
			},
			&credentials.EnvProvider{},
		},
	)
)

func init() {
	flag.StringVar(&templatePath, "template_path", "", "Path to CloudFormation template")
	flag.StringVar(&region, "region", "us-west-1", "AWS Region")
	flag.StringVar(&public, "public", "True", "True or False - Whether to associate a public IP")
	flag.StringVar(&role, "role", "", "Opsee IAM Role name")
	flag.BoolVar(&doDelete, "delete", true, "Delete stack after creation")

	config = &launcher.BastionConfig{
		Tag:         "stable",
		OwnerID:     "933693344490",
		VPNRemote:   "bastion.opsee.com",
		DNSServer:   "169.254.169.253",
		NSQDHost:    "nsqd.in.opsee.com",
		BartnetHost: "bartnet.in.opsee.com",
		AuthType:    "BASIC_TOKEN",
	}

	bastion = &com.Bastion{
		ID:           "",
		CustomerID:   "",
		UserID:       0,
		ImageID:      sql.NullString{},
		InstanceType: "t2.micro",
		VPCID:        "",
		SubnetID:     "",
		Password:     "",
	}

	flag.StringVar(&bastion.ID, "bastion_id", "", "Bastion ID")
	flag.StringVar(&bastion.VPCID, "vpc_id", "", "VPC ID")
	flag.StringVar(&bastion.SubnetID, "subnet_id", "", "Subnet ID")
	flag.StringVar(&bastion.Password, "password", "", "Bastion VPN Password")

	user = &schema.User{}

	flag.StringVar(&user.CustomerId, "customer_id", "", "Customer ID")
	flag.IntVar(&userId, "user_id", 0, "User ID")
	flag.StringVar(&user.Email, "email", "", "Customer Email")
	user.Id = int32(userId)
}

func ptos(s *string) string {
	if s != nil {
		return *s
	}
	return ""
}

func getLatestAMI() (string, error) {
	ec2client := ec2.New(session.New(&aws.Config{
		Credentials: creds,
		MaxRetries:  aws.Int(11),
		Region:      aws.String(region),
	}))

	imageOutput, err := ec2client.DescribeImages(&ec2.DescribeImagesInput{
		Owners: []*string{
			aws.String(config.OwnerID),
		},
		Filters: []*ec2.Filter{
			{
				Name:   aws.String("tag:release"),
				Values: []*string{aws.String(config.Tag)},
			},
			{
				Name:   aws.String("is-public"),
				Values: []*string{aws.String("true")},
			},
		},
	})

	if err != nil {
		log.Print("Error getting latest bastion AMI")
		return "", err
	}

	sort.Sort(launcher.ImageList(imageOutput.Images))
	imageID := *imageOutput.Images[0].ImageId

	return imageID, nil
}

func launch(imageID string) (string, error) {
	userdata, err := config.GenerateUserData(user, bastion)
	if err != nil {
		log.Print("Unable to generate user data.")
		return "", err
	}

	templateBytes, err := ioutil.ReadFile(templatePath)
	if err != nil {
		log.Print("Unable to read template file.")
		return "", err
	}

	stackParameters := []*cloudformation.Parameter{
		{
			ParameterKey:   aws.String("ImageId"),
			ParameterValue: aws.String(imageID),
		},
		{
			ParameterKey:   aws.String("InstanceType"),
			ParameterValue: aws.String(bastion.InstanceType),
		},
		{
			ParameterKey:   aws.String("UserData"),
			ParameterValue: aws.String(base64.StdEncoding.EncodeToString(userdata)),
		},
		{
			ParameterKey:   aws.String("VpcId"),
			ParameterValue: aws.String(bastion.VPCID),
		},
		{
			ParameterKey:   aws.String("SubnetId"),
			ParameterValue: aws.String(bastion.SubnetID),
		},
		{
			ParameterKey:   aws.String("AssociatePublicIpAddress"),
			ParameterValue: aws.String(public),
		},
		{
			ParameterKey:   aws.String("BastionId"),
			ParameterValue: aws.String(bastion.ID),
		},
		{
			ParameterKey:   aws.String("CustomerId"),
			ParameterValue: aws.String(bastion.CustomerID),
		},
		{
			ParameterKey:   aws.String("OpseeRole"),
			ParameterValue: aws.String(role),
		},
	}

	createStackResponse, err := cloudformationClient.CreateStack(&cloudformation.CreateStackInput{
		StackName:    aws.String("opsee-bastion-" + bastion.ID),
		TemplateBody: aws.String(string(templateBytes)),
		Capabilities: []*string{
			aws.String("CAPABILITY_IAM"),
		},
		Parameters: stackParameters,
		Tags: []*cloudformation.Tag{
			{
				Key:   aws.String("Name"),
				Value: aws.String("Opsee Bastion " + bastion.ID),
			},
		},
	})
	if err != nil {
		log.Print("Error creating CloudFormation stack")
		return "", err
	}

	// fmt.Print(awsutil.Prettify(createStackResponse))
	stackID := createStackResponse.StackId
	var stackState string
	var stack *cloudformation.Stack
	var fmtString = "%-32s %-20s %-64s\n"
	for {
		// cmd := exec.Command("clear")
		// cmd.Stdout = os.Stdout
		// cmd.Run()
		fmt.Println()

		stackDescriptionResponse, err := cloudformationClient.DescribeStacks(&cloudformation.DescribeStacksInput{
			StackName: stackID,
		})
		if err != nil {
			log.Print("Error describing stacks")
			return "", err
		}
		stack = stackDescriptionResponse.Stacks[0]
		stackState = *stack.StackStatus

		switch stackState {
		case "CREATE_COMPLETE":
			return *stackID, nil
		case "ROLLBACK_COMPLETE":
			return "", fmt.Errorf(stackState)
		default:
			stackResourcesResponse, err := cloudformationClient.DescribeStackResources(&cloudformation.DescribeStackResourcesInput{
				StackName: stackID,
			})
			if err != nil {
				return "", err
			}

			// fmt.Println(awsutil.Prettify(stackResourcesResponse))

			fmt.Printf(fmtString, "RESOURCE TYPE", "RESOURCE STATUS", "RESOURCE STATUS REASON")
			for _, res := range stackResourcesResponse.StackResources {
				fmt.Printf(fmtString, ptos(res.ResourceType), ptos(res.ResourceStatus), ptos(res.ResourceStatusReason))
			}
		}

		time.Sleep(5 * time.Second)
	}
}

func delete(stackID string) error {
	_, err := cloudformationClient.DeleteStack(&cloudformation.DeleteStackInput{
		StackName: aws.String(stackID),
	})

	if err != nil {
		return err
	}

	log.Print("Deleting stack...")

	for {
		describeStackResponse, err := cloudformationClient.DescribeStacks(&cloudformation.DescribeStacksInput{
			StackName: aws.String(stackID),
		})

		if err != nil {
			return err
		}

		stack := describeStackResponse.Stacks[0]

		if stack == nil {
			return nil
		}

		if ptos(stack.StackStatus) == "DELETE_COMPLETE" {
			log.Print("Delete complete.")
			return nil
		}

		time.Sleep(5 * time.Second)
	}
}

func main() {
	flag.Parse()
	cloudformationClient = cloudformation.New(session.New(), aws.NewConfig().WithCredentials(creds).WithRegion(region))

	ami, err := getLatestAMI()
	if err != nil {
		log.Fatal(err)
	}

	stackID, err := launch(ami)
	if err != nil {
		log.Print(err)
	}

	if stackID == "" {
		os.Exit(1)
	}

	if doDelete {
		err = delete(stackID)
		if err != nil {
			log.Fatal(err)
		}
	}
}
