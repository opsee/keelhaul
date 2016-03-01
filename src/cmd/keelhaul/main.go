package main

import (
	"io/ioutil"
	"os"
	"time"

	etcd "github.com/coreos/etcd/client"
	"github.com/opsee/keelhaul/bus"
	"github.com/opsee/keelhaul/config"
	"github.com/opsee/keelhaul/launcher"
	"github.com/opsee/keelhaul/notifier"
	"github.com/opsee/keelhaul/router"
	"github.com/opsee/keelhaul/service"
	"github.com/opsee/keelhaul/store"
	"github.com/opsee/keelhaul/tracker"
	"github.com/opsee/vaper"
	log "github.com/sirupsen/logrus"
)

func main() {
	cfg := &config.Config{
		PublicHost:                 mustEnvString("KEELHAUL_ADDRESS"),
		PostgresConn:               mustEnvString("KEELHAUL_POSTGRES_CONN"),
		EtcdAddr:                   mustEnvString("KEELHAUL_ETCD_ADDR"),
		BastionConfigKey:           mustEnvString("KEELHAUL_BASTION_CONFIG_KEY"),
		BastionCFTemplate:          mustEnvString("KEELHAUL_BASTION_CF_TEMPLATE"),
		VapeEmailEndpoint:          mustEnvString("KEELHAUL_VAPE_EMAIL_ENDPOINT"),
		VapeUserInfoEndpoint:       mustEnvString("KEELHAUL_VAPE_USERINFO_ENDPOINT"),
		VapeKey:                    mustEnvString("KEELHAUL_VAPE_KEYFILE"),
		FieriEndpoint:              mustEnvString("KEELHAUL_FIERI_ENDPOINT"),
		LaunchesSlackEndpoint:      mustEnvString("KEELHAUL_LAUNCHES_SLACK_ENDPOINT"),
		LaunchesErrorSlackEndpoint: mustEnvString("KEELHAUL_LAUNCHES_ERROR_SLACK_ENDPOINT"),
		TrackerSlackEndpoint:       mustEnvString("KEELHAUL_TRACKER_SLACK_ENDPOINT"),
		NSQDAddr:                   mustEnvString("KEELHAUL_NSQD_HOST"),
		NSQTopic:                   mustEnvString("KEELHAUL_NSQ_TOPIC"),
		NSQLookupds:                mustEnvString("KEELHAUL_NSQLOOKUPD_ADDRS"),
		BartnetEndpoint:            mustEnvString("KEELHAUL_BARTNET_ENDPOINT"),
		BeavisEndpoint:             mustEnvString("KEELHAUL_BEAVIS_ENDPOINT"),
		HugsEndpoint:               mustEnvString("KEELHAUL_HUGS_ENDPOINT"),
		SpanxEndpoint:              mustEnvString("KEELHAUL_SPANX_ENDPOINT"),
	}

	key, err := ioutil.ReadFile(cfg.VapeKey)
	if err != nil {
		log.Error("Unable to read keyfile:", cfg.VapeKey)
		log.Fatal(err)
	}
	vaper.Init(key)

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

	notifier := notifier.New(cfg)

	launcher, err := launcher.New(db, router, etcdKeysAPI, bus, notifier, cfg)
	if err != nil {
		log.Fatalf("couldn't initialize launcher: ", err)
	}

	tracker := tracker.New(db, etcdKeysAPI, notifier)
	tracker.Start()

	certfile := mustEnvString("KEELHAUL_CERT")
	certkeyfile := mustEnvString("KEELHAUL_CERT_KEY")

	svc := service.New(db, bus, launcher, router, cfg)
	svc.StartMux(cfg.PublicHost, certfile, certkeyfile)

	tracker.Stop()
	bus.Stop()
}

func mustEnvString(envVar string) string {
	out := os.Getenv(envVar)
	if out == "" {
		log.Fatal(envVar, "must be set")
	}
	return out
}
