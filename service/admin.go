package service

import (
	"github.com/opsee/basic/schema"
	opsee "github.com/opsee/basic/service"
	opsee_types "github.com/opsee/protobuf/opseeproto/types"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"
)

func (s *service) ListBastionStates(ctx context.Context, req *opsee.ListBastionStatesRequest) (*opsee.ListBastionStatesResponse, error) {
	bs, err := s.db.ListBastionStates(req.CustomerIds, req.Filters...)
	if err != nil {
		log.WithError(err).Error("failed to list bastion states")
		return nil, err
	}

	bastionStates := make([]*schema.BastionState, len(bs.States))
	for i, s := range bs.States {
		ts := &opsee_types.Timestamp{}
		err = ts.Scan(s.LastSeen)
		if err != nil {
			log.WithError(err).Error("failed scanning bastion_state last_seen timestamp")
			return nil, err
		}

		bastionStates[i] = &schema.BastionState{
			Id:         s.ID,
			CustomerId: s.CustomerID,
			Status:     s.Status,
			LastSeen:   ts,
			Region:     s.Region,
			VpcId:      s.VpcId,
		}
	}

	return &opsee.ListBastionStatesResponse{BastionStates: bastionStates}, nil
}
