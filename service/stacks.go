package service

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	opsee "github.com/opsee/basic/service"
	"golang.org/x/net/context"
)

func (s *service) LaunchStack(ctx context.Context, req *opsee.LaunchStackRequest) (*opsee.LaunchStackResponse, error) {
	if req.User == nil {
		return nil, errMissingUser
	}

	err := req.User.Validate()
	if err != nil {
		return nil, err
	}

	if req.Region == "" {
		return nil, errMissingRegion
	}

	if req.VpcId == "" {
		return nil, errMissingVpc
	}

	if req.SubnetId == "" {
		return nil, errMissingSubnet
	}

	if req.SubnetRouting == "" {
		return nil, errMissingSubnetRouting
	}

	if req.InstanceSize == "" {
		req.InstanceSize = "t2.micro"
	}

	// get creds from spanx
	stscreds, err := s.spanx.GetCredentials(ctx, &opsee.GetCredentialsRequest{
		User: req.User,
	})

	if err != nil {
		return nil, err
	}

	sess := session.New(&aws.Config{
		Credentials: credentials.NewStaticCredentials(
			stscreds.Credentials.GetAccessKeyID(),
			stscreds.Credentials.GetSecretAccessKey(),
			stscreds.Credentials.GetSessionToken(),
		),
		Region:     aws.String(req.Region),
		MaxRetries: aws.Int(11),
	})

	_, err = s.launcher.LaunchBastion(sess, req.User, req.Region, req.VpcId, req.SubnetId, req.SubnetRouting, req.InstanceSize, "stable")
	if err != nil {
		return nil, err
	}

	return &opsee.LaunchStackResponse{}, nil
}
