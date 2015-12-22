package router

import (
	"encoding/json"
	"errors"
	"github.com/cenkalti/backoff"
	etcd "github.com/coreos/etcd/client"
	"github.com/opsee/keelhaul/com"
	"golang.org/x/net/context"
	"path"
	"time"
)

type Router interface {
	GetServices(*com.Bastion) (map[string]interface{}, error)
}

type router struct {
	etcd etcd.KeysAPI
}

const (
	basePath = "/opsee.co/routes"
)

var (
	ErrNotFound = errors.New("bastion not found")
)

func New(etcd etcd.KeysAPI) Router {
	return &router{
		etcd: etcd,
	}
}

func (c *router) GetServices(bastion *com.Bastion) (map[string]interface{}, error) {
	var (
		services map[string]interface{}
		err      error
	)

	backoff.Retry(func() error {
		services, err = c.getServices(bastion)
		if err != nil && err != ErrNotFound {
			return err
		}

		return nil

	}, &backoff.ExponentialBackOff{
		InitialInterval:     100 * time.Millisecond,
		RandomizationFactor: 0.5,
		Multiplier:          1.5,
		MaxInterval:         time.Second,
		MaxElapsedTime:      5 * time.Second,
		Clock:               &systemClock{},
	})

	if err != nil {
		return nil, err
	}

	return services, nil
}

func (c *router) getServices(bastion *com.Bastion) (map[string]interface{}, error) {
	response, err := c.etcd.Get(context.Background(), path.Join(basePath, bastion.CustomerID, bastion.ID), &etcd.GetOptions{
		Recursive: true,
		Sort:      true,
		Quorum:    true,
	})

	if err != nil {
		return nil, remapErrors(err)
	}

	services := make(map[string]interface{})

	err = json.Unmarshal([]byte(response.Node.Value), &services)
	if err != nil {
		return nil, err
	}

	return services, nil
}

type systemClock struct{}

func (s *systemClock) Now() time.Time {
	return time.Now()
}

func remapErrors(err error) error {
	if etcdErr, ok := err.(etcd.Error); ok {
		switch etcdErr.Code {
		case etcd.ErrorCodeKeyNotFound:
			return ErrNotFound
		}
	}

	return err
}
