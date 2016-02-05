package store

import (
	"github.com/opsee/basic/com"
)

type Store interface {
	PutBastion(*com.Bastion) error
	UpdateBastion(*com.Bastion) error
	PutRegion(*com.Region) error

	GetBastion(*GetBastionRequest) (*GetBastionResponse, error)
	ListBastions(*ListBastionsRequest) (*ListBastionsResponse, error)

	UpdateTracking([]string) error
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
