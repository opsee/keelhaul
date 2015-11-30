package service

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awsutil"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/opsee/keelhaul/com"
	log "github.com/sirupsen/logrus"
)

type ScanVPCsRequest struct {
	AccessKey string   `json:"access_key"`
	SecretKey string   `json:"secret_key"`
	Regions   []string `json:"regions"`
}

type ScanVPCsResponse struct {
	Regions []*com.Region `json:"regions"`
}

func (r *ScanVPCsRequest) Validate() error {
	if r.AccessKey == "" {
		return errMissingAccessKey
	}

	if r.SecretKey == "" {
		return errMissingSecretKey
	}

	if len(r.Regions) < 1 {
		return errMissingRegion
	}

	for _, reg := range r.Regions {
		if regions[reg] != true {
			return fmt.Errorf("provided region is not valid: %s", reg)
		}
	}

	return nil
}

func (s *service) ScanVPCs(user *com.User, request *ScanVPCsRequest) (*ScanVPCsResponse, error) {
	logger := log.WithFields(log.Fields{
		"customer-id": user.CustomerID,
		"user-id":     user.ID,
	})

	creds := credentials.NewStaticCredentials(
		request.AccessKey,
		request.SecretKey,
		"",
	)

	vpcRegions := make([]*com.Region, len(request.Regions))

	for ri, region := range request.Regions {
		ec2Client := ec2.New(session.New(&aws.Config{
			Credentials: creds,
			Region:      aws.String(region),
			MaxRetries:  aws.Int(11),
		}))

		vpcOutput, err := ec2Client.DescribeVpcs(nil)
		if err != nil {
			return nil, err
		}

		vpcs := make([]*com.VPC, len(vpcOutput.Vpcs))
		for vi, v := range vpcOutput.Vpcs {
			vpc := &com.VPC{}
			awsutil.Copy(vpc, v)
			vpcs[vi] = vpc
		}

		subnetOutput, err := ec2Client.DescribeSubnets(nil)
		if err != nil {
			return nil, err
		}

		subnets := make([]*com.Subnet, len(subnetOutput.Subnets))
		for si, s := range subnetOutput.Subnets {
			subnet := &com.Subnet{}
			awsutil.Copy(subnet, s)
			subnets[si] = subnet
		}

		accountOutput, err := ec2Client.DescribeAccountAttributes(&ec2.DescribeAccountAttributesInput{
			AttributeNames: []*string{
				aws.String("supported-platforms"),
			},
		})
		if err != nil {
			return nil, err
		}

		supportedPlatforms := make([]*string, 0)
		for _, a := range accountOutput.AccountAttributes {
			if *a.AttributeName == "supported-platforms" {
				for _, v := range a.AttributeValues {
					supportedPlatforms = append(supportedPlatforms, v.AttributeValue)
				}
			}
		}

		vpcRegions[ri] = &com.Region{
			Region:             region,
			SupportedPlatforms: supportedPlatforms,
			VPCs:               vpcs,
			Subnets:            subnets,
		}
	}

	// let's save this data, but we'll have to ignore errors
	go func() {
		for _, region := range vpcRegions {
			region.CustomerID = user.CustomerID

			err := s.db.PutRegion(region)
			if err != nil {
				logger.WithError(err).Errorf("error saving region: %#v", *region)
			}
		}
	}()

	return &ScanVPCsResponse{
		Regions: vpcRegions,
	}, nil
}
