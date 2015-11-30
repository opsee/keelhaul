package main

import (
	etcd "github.com/coreos/etcd/client"
	"github.com/opsee/keelhaul/bus"
	"github.com/opsee/keelhaul/config"
	"github.com/opsee/keelhaul/launcher"
	"github.com/opsee/keelhaul/router"
	"github.com/opsee/keelhaul/service"
	"github.com/opsee/keelhaul/store"
	log "github.com/sirupsen/logrus"
	"os"
	"time"
)

func main() {
	cfg := &config.Config{
		PublicHost:        mustEnvString("KEELHAUL_ADDRESS"),
		PostgresConn:      mustEnvString("POSTGRES_CONN"),
		EtcdAddr:          mustEnvString("ETCD_ADDR"),
		BastionConfigKey:  mustEnvString("BASTION_CONFIG_KEY"),
		BastionCFTemplate: mustEnvString("BASTION_CF_TEMPLATE"),
		VapeEndpoint:      mustEnvString("VAPE_ENDPOINT"),
		FieriEndpoint:     mustEnvString("FIERI_ENDPOINT"),
		SlackEndpoint:     mustEnvString("SLACK_ENDPOINT"),
		NSQDAddr:          mustEnvString("NSQD_HOST"),
		NSQTopic:          mustEnvString("NSQ_TOPIC"),
		NSQLookupds:       mustEnvString("NSQLOOKUPD_ADDRS"),
	}

	db, err := store.NewPostgres(cfg.PostgresConn)
	if err != nil {
		log.Fatalf("Error while initializing postgres: ", err)
	}

	etcdClient, err := etcd.New(etcd.Config{
		Endpoints:               []string{cfg.EtcdAddr},
		Transport:               etcd.DefaultTransport,
		HeaderTimeoutPerRequest: time.Second,
	})
	if err != nil {
		log.Fatalf("couldn't initialize etcd client: ", err)
	}
	etcdKeysAPI := etcd.NewKeysAPI(etcdClient)

	bus, err := bus.New(cfg)
	if err != nil {
		log.Fatalf("couldn't initialize bus: ", err)
	}
	bus.Start()

	router := router.New(etcdKeysAPI)

	launcher := launcher.New(db, router, etcdKeysAPI, bus, cfg)

	svc := service.New(db, bus, launcher, router, cfg)
	svc.StartHTTP(cfg.PublicHost)

	bus.Stop()
}

func mustEnvString(envVar string) string {
	out := os.Getenv(envVar)
	if out == "" {
		log.Fatal(envVar, "must be set")
	}
	return out
}
