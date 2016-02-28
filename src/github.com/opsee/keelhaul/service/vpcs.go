package service

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/opsee/basic/com"
	"github.com/opsee/basic/schema"
	opsee "github.com/opsee/basic/service"
	"github.com/opsee/keelhaul/scanner"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	"regexp"
)

var (
	vpcRegexp = regexp.MustCompile(`(?i)vpc`)
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

func (s *service) DeprecatedScanVPCs(user *schema.User, request *ScanVPCsRequest) (*ScanVPCsResponse, error) {
	logger := log.WithFields(log.Fields{
		"customer-id": user.CustomerId,
		"user-id":     user.Id,
	})

	creds := credentials.NewStaticCredentials(
		request.AccessKey,
		request.SecretKey,
		"",
	)

	vpcRegions := make([]*com.Region, len(request.Regions))

	for ri, region := range request.Regions {
		r, err := scanner.ScanRegion(region, session.New(&aws.Config{
			Credentials: creds,
			Region:      aws.String(region),
			MaxRetries:  aws.Int(11),
		}))

		if err != nil {
			logger.WithError(err).Errorf("error scanning region: %s", region)
			return nil, err
		}

		hasVPC := false
		for _, sp := range r.SupportedPlatforms {
			if vpcRegexp.MatchString(aws.StringValue(sp)) {
				hasVPC = true
				break
			}
		}

		logger.Infof("region has VPC support: %t", hasVPC)

		vpcRegions[ri] = r
	}

	// let's save this data, but we'll have to ignore errors
	go func() {
		for _, region := range vpcRegions {
			region.CustomerID = user.CustomerId

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

func (s *service) ScanVpcs(ctx context.Context, req *opsee.ScanVpcsRequest) (*opsee.ScanVpcsResponse, error) {
	return nil, nil
}
