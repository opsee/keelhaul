package service

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/opsee/keelhaul/com"
	"github.com/opsee/keelhaul/store"
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
	ID        string `json:"id"`
	SubnetID  string `json:"subnet_id"`
	BastionID string `json:"bastion_id"`
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
				return fmt.Errorf("vpc_id is missing.")
			}
		}
	}

	if instanceSizes[r.InstanceSize] != true {
		return fmt.Errorf("provided instance_size is not valid: %s", r.InstanceSize)
	}

	return nil
}

func (s *service) LaunchBastions(user *com.User, request *LaunchBastionsRequest) (*LaunchBastionsResponse, error) {
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
			launch, err := s.launcher.LaunchBastion(sess, user, vpc.ID, vpc.SubnetID, request.InstanceSize)
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

func (s *service) ListBastions(user *com.User, request *ListBastionsRequest) (*ListBastionsResponse, error) {
	response, err := s.db.ListBastions(&store.ListBastionsRequest{
		CustomerID: user.CustomerID,
		State:      request.State,
	})

	if err != nil {
		return nil, err
	}

	for _, bastion := range response.Bastions {
		_, err = s.router.GetServices(bastion)
		if err == nil {
			bastion.Connected = true
		}
	}

	return &ListBastionsResponse{Bastions: response.Bastions}, nil
}

type AuthenticateBastionRequest struct {
	ID       string `json:"id"`
	Password string `json:"password"`
}

type AuthenticateBastionResponse struct{}

func (r *AuthenticateBastionRequest) Validate() error {
	if r.ID == "" || r.Password == "" {
		return fmt.Errorf("must provide id and password")
	}

	return nil
}

func (s *service) AuthenticateBastion(request *AuthenticateBastionRequest) (*AuthenticateBastionResponse, error) {
	response, err := s.db.GetBastion(&store.GetBastionRequest{
		ID:    request.ID,
		State: "active",
	})

	if err != nil {
		return nil, errUnauthorized
	}

	err = response.Bastion.Authenticate(request.Password)
	if err != nil {
		return nil, errUnauthorized
	}

	return &AuthenticateBastionResponse{}, nil
}
