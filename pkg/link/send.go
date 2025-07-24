// SPDX-FileCopyrightText: © 2023 OneEyeFPV oneeyefpv@gmail.com
// SPDX-License-Identifier: GPL-3.0-or-later
// SPDX-License-Identifier: FS-0.9-or-later

package link

import (
	"errors"
	"fmt"
	crsf "github.com/kaack/elrs-joystick-control/pkg/crossfire"
	telem "github.com/kaack/elrs-joystick-control/pkg/crossfire/telemetry"
	"github.com/kaack/elrs-joystick-control/pkg/serial"
	"gopkg.in/tomb.v2"
	"time"
)

func (c *Controller) StartSendLoop(port *serial.Port, sendChan chan any, recvChan chan any) error {
	if c.sendLoopTomb != nil && c.sendLoopTomb.Alive() {
		return errors.New("send loop is already active")
	}

	c.sendLoopTomb = &tomb.Tomb{}
	c.sendLoopTomb.Go(func() error {
		return c.SendLoop(port, sendChan, recvChan)
	})

	return nil
}

func (c *Controller) StopSendLoop() error {
	if c.sendLoopTomb == nil || !c.sendLoopTomb.Alive() {
		return nil
	}

	c.sendLoopTomb.Kill(nil)
	if err := c.sendLoopTomb.Wait(); err != nil {
		return err
	}
	return nil
}

func (c *Controller) SendLoop(port *serial.Port, sendChan chan any, recvChan chan any) error {
	currentRefreshRate := crsf.GetRefreshRate(port.BaudRate)
	nextRefreshRate := currentRefreshRate

	fmt.Printf("(send-loop) starting, refresh rate %v\n", currentRefreshRate)

	var err error
	ticker := time.NewTicker(currentRefreshRate)

	c.sentPacketsCount = 0

Loop:
	for {
		select {
		case <-c.sendLoopTomb.Dying():
			break Loop
			
		case chData := <-sendChan:
			switch data := (chData).(type) {
			case ChannelRequest:
				if data == SendModelId {
					fmt.Printf("(send-loop) writing model id frame\n")
					if _, err = port.Write(crsf.CreateModelIDFrame(0)); err != nil {
						c.errorPacketsCount += 1
						fmt.Printf("(send-loop) could not write model id frame on port %s. %s\n", port.Name, err.Error())
					}
				} else if data == PingDevices {
					fmt.Printf("(send-loop) pinging devices\n")
					if _, err = port.Write(crsf.CreatePingDevicesFrame()); err != nil {
						c.errorPacketsCount += 1
						fmt.Printf("(send-loop) could not write ping devices frame on port %s. %s\n", port.Name, err.Error())
					}
				}
			case *telem.TelemSyncType:
				nextRefreshRate = crsf.AdjustSendRate((*data).Rate(), (*data).Offset())
				ticker.Reset(nextRefreshRate)
			default:
				// Ignore other requests
			}

		case <-ticker.C:
			channels := c.GetChannels()
			if _, err = port.Write(crsf.PackChannels(&channels)); err != nil {
				fmt.Printf("(send-loop) could not write channels on port %s. %s\n", port.Name, err.Error())
				break Loop
			}
			c.sentPacketsCount += 1
		}
	}

	fmt.Println("(send-loop): exiting send loop ...")
	return nil
}
