package service

import (
	opsee "github.com/opsee/basic/service"
	"github.com/opsee/keelhaul/bus"
	"github.com/opsee/keelhaul/config"
	"github.com/opsee/keelhaul/launcher"
	"github.com/opsee/keelhaul/router"
	"github.com/opsee/keelhaul/store"
	"google.golang.org/grpc"
)

var instanceSizes = map[string]bool{
	"t2.micro":    true,
	"t2.small":    true,
	"t2.medium":   true,
	"t2.large":    true,
	"m4.large":    true,
	"m4.xlarge":   true,
	"m4.2xlarge":  true,
	"m4.4xlarge":  true,
	"m4.10xlarge": true,
	"m3.medium":   true,
	"m3.large":    true,
	"m3.xlarge":   true,
	"m3.2xlarge":  true,
}

var regions = map[string]bool{
	"ap-northeast-1": true,
	"ap-southeast-1": true,
	"ap-southeast-2": true,
	"eu-central-1":   true,
	"eu-west-1":      true,
	"sa-east-1":      true,
	"us-east-1":      true,
	"us-west-1":      true,
	"us-west-2":      true,
}

type service struct {
	db         store.Store
	launcher   launcher.Launcher
	bus        bus.Bus
	router     router.Router
	config     *config.Config
	grpcServer *grpc.Server
	spanx      opsee.SpanxClient
}

func New(db store.Store, bus bus.Bus, launch launcher.Launcher, router router.Router, spanxclient opsee.SpanxClient, cfg *config.Config) *service {
	s := &service{
		db:       db,
		launcher: launch,
		bus:      bus,
		router:   router,
		spanx:    spanxclient,
		config:   cfg,
	}

	server := grpc.NewServer()
	opsee.RegisterKeelhaulServer(server, s)

	s.grpcServer = server

	return s
}
