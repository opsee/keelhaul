package service

import (
	"github.com/opsee/basic/com"
	"github.com/opsee/basic/schema"
	opsee "github.com/opsee/basic/service"
	"github.com/opsee/keelhaul/router"
	"github.com/opsee/keelhaul/store"
	log "github.com/opsee/logrus"
	"golang.org/x/net/context"
)

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
