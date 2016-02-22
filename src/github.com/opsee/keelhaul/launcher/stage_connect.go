package launcher

import (
	"fmt"
	"github.com/opsee/basic/com"
	"github.com/opsee/keelhaul/bus"
	"math"
	"time"
)

const (
	connectWaitTimeout = 1 * time.Minute
	connectWaitDecay   = float64(0.1)
	connectAttempts    = 50
)

type waitConnect struct{}

func (s waitConnect) Execute(launch *Launch) {
	for {
		if launch.connectAttempts > connectAttempts {
			launch.error(
				fmt.Errorf("timed out waiting for bastion to connect"),
				&bus.Message{
					Command: commandConnectBastion,
					Message: "timed out waiting for bastion to connect",
				},
			)

			return
		}

		_, err := launch.router.GetServices(launch.Bastion)

		if launch.Bastion.State == com.BastionStateActive && err == nil {
			launch.event(&bus.Message{
				State:   stateComplete,
				Command: commandConnectBastion,
				Message: "bastion active and connected",
			})

			return
		}

		launch.event(&bus.Message{
			State:   stateInProgress,
			Command: commandConnectBastion,
			Message: "waiting for bastion connection",
		})

		time.Sleep(decay(launch.connectAttempts))
		launch.connectAttempts = launch.connectAttempts + float64(1)
	}
}

func decay(attempts float64) time.Duration {
	d := time.Duration(float64(connectWaitTimeout) * math.Pow(1-connectWaitDecay, attempts))

	if d < (5 * time.Second) {
		d = 5 * time.Second
	}

	return d
}
