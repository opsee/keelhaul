package store

import (
	"time"

	"github.com/opsee/basic/com"
	"github.com/opsee/basic/schema"

	opsee "github.com/opsee/basic/service"
)

type Store interface {
	PutBastion(*com.Bastion) error
	UpdateBastion(*com.Bastion) error
	PutRegion(*schema.Region) error
	DeprecatedPutRegion(*com.Region) error

	GetBastion(*GetBastionRequest) (*GetBastionResponse, error)
	ListBastions(*ListBastionsRequest) (*ListBastionsResponse, error)

	UpdateTrackingSeen([]string, []string) error
	GetPendingTrackingStates(string) (*TrackingStateResponse, error)
	ListTrackingStates(int, int) (*TrackingStateResponse, error)
	ListBastionStates([]string, ...*opsee.Filter) (*TrackingStateResponse, error)
	UpdateTrackingState(string, string) error
}

type TrackingState struct {
	ID         string    `json:"bastion_id"`
	CustomerID string    `json:"customer_id" db:"customer_id"`
	Status     string    `json:"current_state"`
	LastSeen   time.Time `json:"last_seen" db:"last_seen"`
	Region     string    `json:"region" db:"region"`
	VpcId      string    `json:"vpc_id" db:"vpc_id"`
}

type TrackingStateResponse struct {
	States []*TrackingState
}

type GetBastionRequest struct {
	ID    string
	State string
}

type GetBastionResponse struct {
	Bastion *com.Bastion
}

type ListBastionsRequest struct {
	CustomerID string
	State      []string
}

type ListBastionsResponse struct {
	Bastions []*com.Bastion
}
