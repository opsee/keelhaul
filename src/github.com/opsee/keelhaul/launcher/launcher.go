package launcher

import (
	"github.com/aws/aws-sdk-go/aws/session"
	etcd "github.com/coreos/etcd/client"
	"github.com/opsee/basic/com"
	"github.com/opsee/keelhaul/bus"
	"github.com/opsee/keelhaul/config"
	"github.com/opsee/keelhaul/router"
	"github.com/opsee/keelhaul/store"
)

type Launcher interface {
	LaunchBastion(*session.Session, *com.User, string, string, string, string) (*Launch, error)
}

type launcher struct {
	db     store.Store
	etcd   etcd.KeysAPI
	bus    bus.Bus
	router router.Router
	config *config.Config
}

func New(db store.Store, router router.Router, etcdKAPI etcd.KeysAPI, bus bus.Bus, cfg *config.Config) *launcher {
	return &launcher{
		db:     db,
		router: router,
		etcd:   etcdKAPI,
		bus:    bus,
		config: cfg,
	}
}

func (l *launcher) LaunchBastion(sess *session.Session, user *com.User, vpcID, subnetID, subnetRouting, instanceType string) (*Launch, error) {
	launch := NewLaunch(l.db, l.router, l.etcd, l.config, sess, user)
	go l.watchLaunch(launch)

	// this is done synchronously so that we can return the bastion id
	err := launch.CreateBastion(vpcID, subnetID, subnetRouting, instanceType)
	if err != nil {
		return nil, err
	}

	go launch.Launch()

	return launch, nil
}

func (l *launcher) watchLaunch(launch *Launch) {
	for event := range launch.EventChan {
		l.bus.Publish(event.Message)
	}

	if launch.Err != nil {
		l.NotifyError(launch)
	} else {
		launch.CheckRequestFactory.CheckRequestPool.DrainRequests(true)
		l.NotifySuccess(launch)
	}
}
