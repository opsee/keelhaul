package service

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	opsee "github.com/opsee/basic/service"
	"github.com/opsee/keelhaul/scanner"
	"github.com/opsee/spanx/spanxcreds"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	"regexp"
)

var (
	vpcRegexp = regexp.MustCompile(`(?i)vpc`)
)

func (s *service) ScanVpcs(ctx context.Context, req *opsee.ScanVpcsRequest) (*opsee.ScanVpcsResponse, error) {
	if req.User == nil {
		return nil, errMissingUser
	}

	err := req.User.Validate()
	if err != nil {
		return nil, err
	}

	logger := log.WithFields(log.Fields{
		"customer-id": req.User.CustomerId,
		"user-id":     req.User.Id,
	})

	logger.Info("scan vpcs request")

	if req.Region == "" {
		return nil, errMissingRegion
	}

	sess := session.New(&aws.Config{
		Credentials: spanxcreds.NewSpanxCredentials(req.User, s.spanx),
		Region:      aws.String(req.Region),
		MaxRetries:  aws.Int(11),
	})

	scannedRegion, err := scanner.ScanRegion(req.Region, sess)
	if err != nil {
		return nil, err
	}

	hasVPC := false
	for _, sp := range scannedRegion.SupportedPlatforms {
		if vpcRegexp.MatchString(sp) {
			hasVPC = true
			break
		}
	}

	logger.Infof("region has VPC support: %t", hasVPC)

	// let's save this data, but we'll have to ignore errors
	go func() {
		scannedRegion.CustomerId = req.User.CustomerId

		err := s.db.PutRegion(scannedRegion)
		if err != nil {
			logger.WithError(err).Errorf("error saving region: %#v", *scannedRegion)
		}
	}()

	return &opsee.ScanVpcsResponse{scannedRegion}, nil
}
