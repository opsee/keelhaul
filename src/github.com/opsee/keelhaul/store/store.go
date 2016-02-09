package store

import (
	"github.com/opsee/basic/com"
	"time"
)

type Store interface {
	PutBastion(*com.Bastion) error
	UpdateBastion(*com.Bastion) error
	PutRegion(*com.Region) error

	GetBastion(*GetBastionRequest) (*GetBastionResponse, error)
	ListBastions(*ListBastionsRequest) (*ListBastionsResponse, error)

	UpdateTrackingSeen([]string, []string) error
	ListTrackingStates(string) (*TrackingStateResponse, error)
	UpdateTrackingState(string, string) error
}

type TrackingState struct {
	ID         string    `json:"bastion_id"`
	CustomerID string    `json:"customer_id" db:"customer_id"`
	Status     string    `json:"current_state"`
	LastSeen   time.Time `json:"last_seen" db:"last_seen"`
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
