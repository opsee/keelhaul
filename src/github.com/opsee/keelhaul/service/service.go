package service

import (
	"errors"
	"github.com/opsee/basic/com"
	"github.com/opsee/basic/spanx"
	"github.com/opsee/keelhaul/bus"
	"github.com/opsee/keelhaul/config"
	"github.com/opsee/keelhaul/launcher"
	"github.com/opsee/keelhaul/router"
	"github.com/opsee/keelhaul/store"
)

var (
	errUnauthorized     = errors.New("unauthorized.")
	errMissingAccessKey = errors.New("missing access_key.")
	errMissingSecretKey = errors.New("missing secret_key.")
	errMissingRegion    = errors.New("missing region.")
	errUnknown          = errors.New("unknown error.")
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

type Service interface {
	ScanVPCs(*com.User, *ScanVPCsRequest) (*ScanVPCsResponse, error)
	LaunchBastions(*com.User, *LaunchBastionsRequest) (*LaunchBastionsResponse, error)
}

type service struct {
	db       store.Store
	launcher launcher.Launcher
	bus      bus.Bus
	router   router.Router
	spanx    spanx.Client
	config   *config.Config
}

func New(db store.Store, bus bus.Bus, launch launcher.Launcher, router router.Router, cfg *config.Config) *service {
	return &service{
		db:       db,
		launcher: launch,
		bus:      bus,
		router:   router,
		spanx:    spanx.New(cfg.SpanxEndpoint),
		config:   cfg,
	}
}
