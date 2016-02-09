package tracker

import (
	"path"
	"time"

	etcd "github.com/coreos/etcd/client"
	"github.com/opsee/keelhaul/store"
	"github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"
)

const (
	routePath       = "/opsee.co/routes"
	trackerKeyPath  = "/opsee.co/config/keelhaul/trackerLock"
	leaderCheckRate = time.Duration(30) * time.Second
	serviceTTL      = time.Duration(60) * time.Second
	minUpdateDelay  = time.Duration(180) * time.Second // matches TTL for routes
	updateBatchSize = 128
)

type tracker struct {
	db             store.Store
	etcd           etcd.KeysAPI
	quit           chan struct{}
	isServing      bool
	lastUpdateTime time.Time
	selfID         string
	offerOptions   etcd.SetOptions
	contOptions    etcd.SetOptions
}

func New(db store.Store, etcdKAPI etcd.KeysAPI) *tracker {
	selfID := uuid.NewV4().String()
	return &tracker{
		db:     db,
		etcd:   etcdKAPI,
		quit:   make(chan struct{}, 1),
		selfID: selfID,
		offerOptions: etcd.SetOptions{
			PrevExist: etcd.PrevNoExist,
			TTL:       serviceTTL,
			Dir:       false,
		},
		contOptions: etcd.SetOptions{
			PrevValue: selfID,
			TTL:       serviceTTL,
			Dir:       false,
		},
	}
}

func (t *tracker) Start() {
	go func() {
		leaderTicker := time.NewTicker(leaderCheckRate)
		t.offerService()
		if t.isServing && t.isTimeToUpdate() {
			t.updateSeen()
			t.lastUpdateTime = time.Now()
		}
		for {
			select {
			case <-leaderTicker.C:
				t.offerService()
				if t.isServing && t.isTimeToUpdate() {
					t.updateSeen()
					t.lastUpdateTime = time.Now()
				}
			case <-t.quit:
				leaderTicker.Stop()
				return
			}
		}
	}()
}

func (t *tracker) Stop() {
	t.quit <- struct{}{}
}

func (t *tracker) isTimeToUpdate() bool {
	return (time.Now().Sub(t.lastUpdateTime) >= minUpdateDelay)
}

/*
attempt become leader by setting key only if empty and continue setting it with compare-and-swap.
when not leader then peroidically attempt to become leader
(would be more responsive to use a watcher but the delay here is acceptable)
*/
func (t *tracker) offerService() {
	var setOpts etcd.SetOptions
	if t.isServing {
		setOpts = t.contOptions
	} else {
		setOpts = t.offerOptions
	}

	_, err := t.etcd.Set(context.Background(), trackerKeyPath, t.selfID, &setOpts)
	if err == nil {
		t.isServing = true
	} else {
		t.isServing = false
		// temporarily log all
		log.WithError(err).WithFields(log.Fields{
			"id":        t.selfID,
			"PrevExist": setOpts.PrevExist,
			"PrevValue": setOpts.PrevValue}).Warn("etcd Set error")

		switch err.(type) {
		default:
			log.WithError(err).Error("unexpected etcd error")
		case etcd.Error:
			switch err.(etcd.Error).Code {
			case etcd.ErrorCodeTestFailed:
				return
			case etcd.ErrorCodeNodeExist:
				return
			default:
				log.WithError(err).Error("unexpected etcd error")
			}
		}
	}
}

func (t *tracker) updateSeen() {
	log.WithFields(log.Fields{"id": t.selfID}).Info("tracker thought leader")
	/* TODO find better approach for reads
	possibly consume heartbeats from nsq
	this reads the entire tree including the values.  there will be badness as this scales  */
	response, err := t.etcd.Get(context.Background(), routePath, &etcd.GetOptions{
		Recursive: true,
		Quorum:    false, // shouldn't need consistent reads here
	})
	if err != nil {
		log.Error(err)
		return
	}

	bastBatch := make([]string, 0, updateBatchSize)
	for _, node := range response.Node.Nodes {
		for _, norde := range node.Nodes {
			bastBatch = append(bastBatch, path.Base(norde.Key))
			if len(bastBatch) == updateBatchSize {
				err = t.db.UpdateTracking(bastBatch)
				if err != nil {
					log.Error(err)
				}
				bastBatch = make([]string, 0, updateBatchSize)
			}
		}
	}
	if len(bastBatch) > 0 {
		err = t.db.UpdateTracking(bastBatch)
		if err != nil {
			log.Error(err)
		}
	}

	// TODO update new status in postgres and fire notifications
}
