package tracker

import (
	"encoding/json"
	"path"
	"regexp"
	"time"

	etcd "github.com/coreos/etcd/client"
	"github.com/opsee/keelhaul/notifier"
	"github.com/opsee/keelhaul/store"
	log "github.com/opsee/logrus"
	"github.com/satori/go.uuid"
	"golang.org/x/net/context"
)

const (
	routePath        = "/opsee.co/routes"
	trackerKeyPath   = "/opsee.co/config/keelhaul/trackerLock"
	leaderCheckRate  = time.Duration(30) * time.Second
	serviceTTL       = time.Duration(60) * time.Second
	minUpdateDelay   = time.Duration(180) * time.Second // matches TTL for routes
	updateBatchSize  = 128
	inactiveInterval = "3 minutes"
	uuidFormat       = `^[a-z0-9]{8}-[a-z0-9]{4}-[1-5][a-z0-9]{3}-[a-z0-9]{4}-[a-z0-9]{12}$`
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
	notifier       notifier.Notifier
}

func New(db store.Store, etcdKAPI etcd.KeysAPI, notifier notifier.Notifier) *tracker {
	selfID := uuid.NewV4().String()
	return &tracker{
		db:       db,
		etcd:     etcdKAPI,
		quit:     make(chan struct{}, 1),
		selfID:   selfID,
		notifier: notifier,
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
		log.WithError(err).Error("etcd read failed")
		return
	}

	bastBatch := make([]string, 0, updateBatchSize)
	custBatch := make([]string, 0, updateBatchSize)
	for _, custNode := range response.Node.Nodes {
		custID := path.Base(custNode.Key)
		if !checkUUID(custID) {
			log.WithError(err).Warnf("invalid custID: %s", custID)
			continue
		}
		for _, bastNode := range custNode.Nodes {
			bastID := path.Base(bastNode.Key)
			if !checkUUID(bastID) {
				log.WithError(err).Warnf("invalid bastID: %s", bastID)
				continue
			}

			// right herr we check if the checker service has registered
			services := make(map[string]interface{})
			err := json.Unmarshal([]byte(bastNode.Value), &services)
			if err != nil {
				log.WithError(err).Warnf("couldn't unmarshal services for bastion: %s", bastID)
				continue
			}

			if _, ok := services["checker"]; !ok {
				continue
			}

			bastBatch = append(bastBatch, bastID)
			custBatch = append(custBatch, custID)
			if len(bastBatch) == updateBatchSize {
				err = t.db.UpdateTrackingSeen(bastBatch, custBatch)
				if err != nil {
					log.WithError(err).Error("tracking table update failed")
					return
				}
				bastBatch = make([]string, 0, updateBatchSize)
				custBatch = make([]string, 0, updateBatchSize)
			}
		}
	}
	if len(bastBatch) > 0 {
		err = t.db.UpdateTrackingSeen(bastBatch, custBatch)
		if err != nil {
			log.WithError(err).Error("tracking table update failed")
			return
		}
	}

	states, err := t.db.GetPendingTrackingStates(inactiveInterval)
	if err != nil {
		log.WithError(err).Error("failed to list tracking states")
		return
	}
	for _, s := range states.States {
		var err error
		if s.Status == "active" {
			log.Infof("attempting notify for %s", s.ID)
			s.Status = "inactive"
			err = t.notifier.NotifySlackBastionState(false, s.CustomerID, map[string]interface{}{
				"bastion_id":    s.ID,
				"customer_id":   s.CustomerID,
				"current_state": s.Status,
				"last_seen":     s.LastSeen.Local().Format(time.RFC1123),
			})
			if err == nil {
				err = t.db.UpdateTrackingState(s.ID, "inactive")
			}
		} else {
			log.Infof("attempting notify for %s", s.ID)
			s.Status = "active"
			err = t.notifier.NotifySlackBastionState(true, s.CustomerID, map[string]interface{}{
				"bastion_id":    s.ID,
				"customer_id":   s.CustomerID,
				"current_state": s.Status,
				"last_seen":     s.LastSeen.Local().Format(time.RFC1123),
			})
			if err == nil {
				err = t.db.UpdateTrackingState(s.ID, "active")
			}
		}
		if err != nil {
			log.WithError(err).Error("failed to update tracking state and/or notify")
		}
	}
}

func checkUUID(uuid string) bool {
	uuidExp := regexp.MustCompile(uuidFormat)
	if !uuidExp.MatchString(uuid) {
		return false
	}
	return true
}
