package service

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/opsee/basic/com"
	"github.com/opsee/basic/schema"
	opsee "github.com/opsee/basic/service"
	"github.com/opsee/keelhaul/router"
	"github.com/opsee/keelhaul/store"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"
)

type LaunchBastionsRequest struct {
	AccessKey    string          `json:"access_key"`
	SecretKey    string          `json:"secret_key"`
	Regions      []*LaunchRegion `json:"regions"`
	InstanceSize string          `json:"instance_size"`
}

type LaunchRegion struct {
	Region string       `json:"region"`
	VPCs   []*LaunchVPC `json:"vpcs"`
}

type LaunchVPC struct {
	ID            string `json:"id"`
	SubnetID      string `json:"subnet_id"`
	SubnetRouting string `json:"subnet_routing"`
	BastionID     string `json:"bastion_id"`
}

type LaunchBastionsResponse struct {
	Regions []*LaunchRegion `json:"regions"`
}

func (r *LaunchBastionsRequest) Validate() error {
	if r.AccessKey == "" {
		return fmt.Errorf("access_key is required.")
	}

	if r.SecretKey == "" {
		return fmt.Errorf("secret_key is required.")
	}

	if len(r.Regions) < 1 {
		return fmt.Errorf("must specify at least one region.")
	}

	for _, reg := range r.Regions {
		if regions[reg.Region] != true {
			return fmt.Errorf("provided region is not valid: %s", reg)
		}

		for _, vpc := range reg.VPCs {
			if vpc.ID == "" {
				return fmt.Errorf("vpc id is required.")
			}

			if vpc.SubnetID == "" {
				return fmt.Errorf("vpc subnet_id is required.")
			}

			if vpc.SubnetRouting == "" {
				return fmt.Errorf("vpc subnet_routing is required.")
			}

			_, ok := com.RoutingPreference[vpc.SubnetRouting]
			if !ok {
				return fmt.Errorf("vpc subnet_routing: %s is not valid.", vpc.SubnetRouting)
			}
		}
	}

	if instanceSizes[r.InstanceSize] != true {
		return fmt.Errorf("provided instance_size is not valid: %s", r.InstanceSize)
	}

	return nil
}

func (s *service) LaunchBastions(user *schema.User, request *LaunchBastionsRequest) (*LaunchBastionsResponse, error) {
	creds := credentials.NewStaticCredentials(
		request.AccessKey,
		request.SecretKey,
		"",
	)

	for _, region := range request.Regions {
		sess := session.New(&aws.Config{
			Credentials: creds,
			Region:      aws.String(region.Region),
			MaxRetries:  aws.Int(11),
		})

		for _, vpc := range region.VPCs {
			launch, err := s.launcher.LaunchBastion(sess, user, region.Region, vpc.ID, vpc.SubnetID, vpc.SubnetRouting, request.InstanceSize)
			if err != nil {
				// multiple region launch isn't going to be atomic rn
				return nil, err
			}

			vpc.BastionID = launch.Bastion.ID
		}
	}

	return &LaunchBastionsResponse{request.Regions}, nil
}

type ListBastionsRequest struct {
	State []string `json:"state"`
}

type ListBastionsResponse struct {
	Bastions []*com.Bastion `json:"bastions"`
}

func (r *ListBastionsRequest) Validate() error {
	if len(r.State) == 0 {
		r.State = com.BastionStates
	}

	return nil
}

func (s *service) ListBastions(user *schema.User, request *ListBastionsRequest) (*ListBastionsResponse, error) {
	response, err := s.db.ListBastions(&store.ListBastionsRequest{
		CustomerID: user.CustomerId,
		State:      request.State,
	})

	if err != nil {
		log.WithError(err).WithFields(log.Fields{"customer_id": user.CustomerId, "state": request.State}).Errorf("error querying database")
		return nil, err
	}

	for _, bastion := range response.Bastions {
		logger := log.WithFields(log.Fields{"customer_id": user.CustomerId, "bastion_id": bastion.ID})
		services, err := s.router.GetServices(bastion)
		_, ok := services["checker"]

		if err != nil {
			if err != router.ErrNotFound {
				logger.WithError(err).Error("bastion router error")
			} else {
				logger.Debug("bastion not found in router")
			}
		} else {
			if ok {
				bastion.Connected = true
			} else {
				logger.Debug("bastion found but checker not running")
			}
		}
	}

	return &ListBastionsResponse{Bastions: response.Bastions}, nil
}

func (s *service) AuthenticateBastion(ctx context.Context, request *opsee.AuthenticateBastionRequest) (*opsee.AuthenticateBastionResponse, error) {
	response, err := s.db.GetBastion(&store.GetBastionRequest{
		ID:    request.Id,
		State: "active",
	})

	if err != nil {
		log.WithError(err).WithField("bastion_id", request.Id).Error("not found in database")
		return nil, errUnauthorized
	}

	err = response.Bastion.Authenticate(request.Password)
	if err != nil {
		log.WithError(err).WithField("bastion_id", request.Id).Error("bcrypt comparison failed")
		return nil, errUnauthorized
	}

	log.WithField("bastion_id", request.Id).Info("bastion auth successful")

	return &opsee.AuthenticateBastionResponse{true}, nil
}
