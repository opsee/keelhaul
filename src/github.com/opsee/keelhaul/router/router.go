package router

import (
	"encoding/json"
	etcd "github.com/coreos/etcd/client"
	"github.com/opsee/keelhaul/com"
	"golang.org/x/net/context"
	"path"
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

func New(etcd etcd.KeysAPI) Router {
	return &router{
		etcd: etcd,
	}
}

func (c *router) GetServices(bastion *com.Bastion) (map[string]interface{}, error) {
	response, err := c.etcd.Get(context.Background(), path.Join(basePath, bastion.CustomerID, bastion.ID), &etcd.GetOptions{
		Recursive: true,
		Sort:      true,
		Quorum:    true,
	})
	if err != nil {
		return nil, err
	}

	services := make(map[string]interface{})
	err = json.Unmarshal([]byte(response.Node.Value), &services)
	if err != nil {
		return nil, err
	}

	return services, nil
}
