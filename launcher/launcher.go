package launcher

import (
	"github.com/aws/aws-sdk-go/aws/session"
	etcd "github.com/coreos/etcd/client"
	"github.com/opsee/basic/schema"
	"github.com/opsee/basic/service"
	"github.com/opsee/keelhaul/bus"
	"github.com/opsee/keelhaul/config"
	"github.com/opsee/keelhaul/notifier"
	"github.com/opsee/keelhaul/router"
	"github.com/opsee/keelhaul/store"
	"time"
)

var MagicExgid = "127a7354-290e-11e6-b178-2bc1f6aefc14"

type systemClock struct{}

func (s *systemClock) Now() time.Time {
	return time.Now()
}

type Launcher interface {
	LaunchBastion(*session.Session, *schema.User, string, string, string, string, string, string, string) (*Launch, error)
}

type launcher struct {
	db       store.Store
	etcd     etcd.KeysAPI
	bus      bus.Bus
	router   router.Router
	spanx    service.SpanxClient
	config   *config.Config
	notifier notifier.Notifier
	bezos    service.BezosClient
}

func New(db store.Store, router router.Router, etcdKAPI etcd.KeysAPI, bus bus.Bus, notifier notifier.Notifier, spanxclient service.SpanxClient, bezos service.BezosClient, cfg *config.Config) (*launcher, error) {
	return &launcher{
		db:       db,
		router:   router,
		etcd:     etcdKAPI,
		bus:      bus,
		spanx:    spanxclient,
		config:   cfg,
		notifier: notifier,
		bezos:    bezos,
	}, nil
}

func (l *launcher) LaunchBastion(sess *session.Session, user *schema.User, executionGroupId, region, vpcID, subnetID, subnetRouting, instanceType, imageTag string) (*Launch, error) {
	launch := NewLaunch(l.db, l.router, l.etcd, l.spanx, l.bezos, l.config, sess, user)
	go l.watchLaunch(launch)

	// this is done synchronously so that we can return the bastion id
	err := launch.CreateBastion(executionGroupId, region, vpcID, subnetID, subnetRouting, instanceType)
	if err != nil {
		return nil, err
	}

	go launch.Launch(imageTag)

	return launch, nil
}

func (l *launcher) watchLaunch(launch *Launch) {
	for event := range launch.EventChan {
		l.bus.Publish(event.Message)
	}

	if launch.Err != nil {
		l.notifier.NotifyError(int(launch.User.Id), launch.NotifyVars())
	} else {
		l.notifier.NotifySuccess(int(launch.User.Id), launch.NotifyVars())
	}
}
