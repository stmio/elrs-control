// SPDX-FileCopyrightText: © 2023 OneEyeFPV oneeyefpv@gmail.com
// SPDX-License-Identifier: GPL-3.0-or-later
// SPDX-License-Identifier: FS-0.9-or-later

package link

import (
	"errors"
	"fmt"
	"github.com/kaack/elrs-joystick-control/pkg/crossfire"
	telem "github.com/kaack/elrs-joystick-control/pkg/crossfire/telemetry"
	"github.com/kaack/elrs-joystick-control/pkg/serial"
	"gopkg.in/tomb.v2"
	"time"
)

func (c *Controller) StartRecvLoop(port *serial.Port, sendChan chan any, recvChan chan any) error {
	if c.recvLoopTomb != nil && c.recvLoopTomb.Alive() {
		return errors.New("recv loop is already active")
	}

	c.recvLoopTomb = &tomb.Tomb{}
	c.recvLoopTomb.Go(func() error {
		return c.RecvLoop(port, sendChan, recvChan)
	})

	return nil
}

func (c *Controller) StopRecvLoop() error {
	if c.recvLoopTomb == nil || !c.recvLoopTomb.Alive() {
		return nil
	}

	c.recvLoopTomb.Kill(nil)
	if err := c.recvLoopTomb.Wait(); err != nil {
		return err
	}
	return nil
}

func (c *Controller) RecvLoop(port *serial.Port, sendChan chan any, recvChan chan any) error {
	refreshRate := crossfire.GetRefreshRate(port.BaudRate)
	maxInactivityTime := refreshRate * 4
	fmt.Printf("(recv-loop) starting, refresh rate %v, max inactivity: %v\n", refreshRate, maxInactivityTime)
	ticker := time.NewTicker(refreshRate)

	tickCount := uint64(0)
	currentTickTime := time.Now()
	lastRecvTelemTime := time.Now()
	lastSyncReqTime := time.Now()

	reader := telem.NewReader(port)

	var tPacket telem.TelemType
	var err error

	c.recvPacketsCount = 0
	c.errorPacketsCount = 0

Loop:
	for {
		select {
		case <-c.recvLoopTomb.Dying():
			break Loop

		case chData := <-recvChan:
			switch chData.(type) {
			default:
				//no-op
			}

		case <-ticker.C:
			tickCount += 1
			currentTickTime = time.Now()

			timeSinceLastTelem := currentTickTime.Sub(lastRecvTelemTime) / time.Millisecond
			timeSinceLastSyncReq := currentTickTime.Sub(lastSyncReqTime) / time.Millisecond
			if timeSinceLastTelem > maxInactivityTime && timeSinceLastSyncReq > maxInactivityTime {
				fmt.Printf("(recv-loop) requesting TelemSync lt:%d, ls:%d\n", timeSinceLastTelem, timeSinceLastSyncReq)
				lastSyncReqTime = currentTickTime
				sendChan <- SendModelId
			}

			if tPacket, err = reader.Next(c.recvLoopTomb); err != nil {
				if _, ok := err.(*telem.InterruptedError); ok {
					break
				}
				fmt.Printf("(recv-loop) error reading telemetry data. error: %s\n", err.Error())
				c.errorPacketsCount += 1
				break
			}

			c.recvPacketsCount += 1
			lastRecvTelemTime = currentTickTime

			// Process telemetry based on type
			switch tFrame := (tPacket).(type) {
			case telem.TelemSyncType:
				sendChan <- &tFrame

			case telem.TelemLinkStatsType:
				c.sendLinkStats(LinkStats{
					UplinkRSSI1:  tFrame.UplinkRSSI1(),
					UplinkRSSI2:  tFrame.UplinkRSSI2(),
					UplinkLQ:     tFrame.UplinkLinkQuality(),
					UplinkSNR:    tFrame.UplinkSNR(),
					DownlinkRSSI: tFrame.DownlinkRSSI(),
					DownlinkLQ:   tFrame.DownlinkLinkQuality(),
				})

			case telem.TelemBatteryType:
				c.sendBattery(BatteryData{
					Voltage:   tFrame.Voltage(),
					Current:   tFrame.Current(),
					Remaining: tFrame.Remaining(),
				})

			case telem.TelemGPSType:
				c.sendGPS(GPSData{
					Latitude:    tFrame.Latitude(),
					Longitude:   tFrame.Longitude(),
					Altitude:    tFrame.Altitude(),
					Satellites:  tFrame.Satellites(),
					GroundSpeed: tFrame.GroundSpeed(),
				})

			case telem.TelemAttitudeType:
				c.sendAttitude(AttitudeData{
					Pitch: tFrame.Pitch(),
					Roll:  tFrame.Roll(),
					Yaw:   tFrame.Yaw(),
				})

			// Ignore other telemetry types
			case telem.TelemFlightModeType,
				telem.TelemLinkTXType,
				telem.TelemLinkRXType,
				telem.TelemBarometerType,
				telem.TelemVariometerType,
				telem.TelemBarometerVariometerType:
				// Skip these

			default:
				// Unknown telemetry
			}
		}
	}

	fmt.Printf("(recv-loop) exiting recv telemetry loop...\n")
	return nil
}
