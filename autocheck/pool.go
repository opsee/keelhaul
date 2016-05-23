package autocheck

import (
	"github.com/opsee/basic/schema"
	log "github.com/sirupsen/logrus"
)

type Sink interface {
	Send(*schema.Check) error
}

type Pool struct {
	Sink         Sink
	Logger       *log.Entry
	targets      []Target
	successCount int
}

func NewPool(sink Sink, logger *log.Entry) *Pool {
	return &Pool{
		Sink:   sink,
		Logger: logger,
	}
}

func (p *Pool) AddTarget(obj interface{}) {
	p.targets = append(p.targets, NewTarget(obj))
}

func (p *Pool) AddTargetWithAlarms(obj interface{}, alarms interface{}) {
	p.targets = append(p.targets, NewTargetWithAlarms(obj, alarms))
}

func (p *Pool) Drain() {
	for _, target := range p.targets {
		checks, err := target.Generate()
		if err != nil {
			p.Logger.WithError(err).Error("couldn't generate autocheck target")
			continue
		}

		for _, check := range checks {
			err = p.Sink.Send(check)
			if err != nil {
				p.Logger.WithError(err).Error("couldn't send autocheck")
				continue
			}

			p.successCount++
		}
	}
}

func (p *Pool) SuccessCount() int {
	return p.successCount
}
