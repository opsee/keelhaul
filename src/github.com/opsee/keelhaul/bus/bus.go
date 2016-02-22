package bus

import (
	"encoding/json"
	"fmt"
	"github.com/nsqio/go-nsq"
	"github.com/opsee/basic/schema"
	"github.com/opsee/keelhaul/config"
	"github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
	"strings"
	"time"
)

type Bus interface {
	Start()
	Publish(*Message) error
	Subscribe(*schema.User) *Subscription
	Unsubscribe(*Subscription)
	Stop() error
}

type bus struct {
	config          *config.Config
	channels        map[string]map[*Subscription]bool
	subscribeChan   chan *Subscription
	unsubscribeChan chan *Subscription
	messageChan     chan *Message
	producer        *nsq.Producer
	consumer        *nsq.Consumer
}

type Subscription struct {
	CustomerID string
	Chan       chan *Message
}

func New(cfg *config.Config) (*bus, error) {
	b := &bus{
		config:          cfg,
		channels:        make(map[string]map[*Subscription]bool),
		subscribeChan:   make(chan *Subscription),
		unsubscribeChan: make(chan *Subscription),
		messageChan:     make(chan *Message),
	}

	channel := uuid.NewV4().String() + "#ephemeral"

	lookupdAddrs := make([]string, 0)
	for _, a := range strings.Split(cfg.NSQLookupds, ",") {
		lookupdAddrs = append(lookupdAddrs, strings.Trim(a, " "))
	}

	consumer, err := nsq.NewConsumer(cfg.NSQTopic, channel, nsq.NewConfig())
	if err != nil {
		return nil, err
	}

	consumer.AddConcurrentHandlers(b, 4)
	consumer.ConnectToNSQLookupds(lookupdAddrs)

	producer, err := nsq.NewProducer(cfg.NSQDAddr, nsq.NewConfig())
	if err != nil {
		return nil, err
	}

	b.consumer = consumer
	b.producer = producer

	return b, nil
}

func (b *bus) Start() {
	go func() {
		for {
			select {
			case sub := <-b.subscribeChan:
				_, ok := b.channels[sub.CustomerID]
				if !ok {
					b.channels[sub.CustomerID] = make(map[*Subscription]bool)
				}

				b.channels[sub.CustomerID][sub] = true

			case sub := <-b.unsubscribeChan:
				channel, ok := b.channels[sub.CustomerID]
				if !ok {
					return
				}

				_, ok = channel[sub]
				if ok {
					close(sub.Chan)
					delete(channel, sub)
				}

				if len(channel) == 0 {
					delete(b.channels, sub.CustomerID)
				}

			case msg := <-b.messageChan:
				channel, ok := b.channels[msg.CustomerID]
				if ok {
					for sub, _ := range channel {
						sub.Chan <- msg
					}
				}
			}
		}
	}()
}

func (b *bus) Publish(msg *Message) error {
	err := msg.Validate()
	if err != nil {
		return err
	}

	msgBytes, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	return b.producer.Publish(b.config.NSQTopic, msgBytes)
}

func (b *bus) Subscribe(user *schema.User) *Subscription {
	sub := &Subscription{
		CustomerID: user.CustomerId,
		Chan:       make(chan *Message),
	}

	b.subscribeChan <- sub

	return sub
}

func (b *bus) Unsubscribe(sub *Subscription) {
	b.unsubscribeChan <- sub
}

func (b *bus) HandleMessage(m *nsq.Message) error {
	message := &Message{}
	err := json.Unmarshal(m.Body, message)
	if err != nil {
		b.handleError(m, err)
		return nil
	}

	err = message.Validate()
	if err != nil {
		b.handleError(m, err)
		return nil
	}

	b.messageChan <- message

	return nil
}

func (b *bus) Stop() error {
	b.consumer.Stop()

	var err error

	select {
	case <-b.consumer.StopChan:
		err = nil
	case <-time.After(15 * time.Second):
		err = fmt.Errorf("timed out waiting for consumer shutdown")
	}

	return err
}

func (b *bus) handleError(m *nsq.Message, err error) {
	log.WithError(err).Warn("error processing nsq message")
}
