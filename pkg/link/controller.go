// SPDX-FileCopyrightText: © 2023 OneEyeFPV oneeyefpv@gmail.com
// SPDX-License-Identifier: GPL-3.0-or-later
// SPDX-License-Identifier: FS-0.9-or-later

package link

import (
	sc "github.com/kaack/elrs-joystick-control/pkg/serial"
	"github.com/kaack/elrs-joystick-control/pkg/util"
	"gopkg.in/tomb.v2"
	"sync"
)

// Simple telemetry types
type LinkStats struct {
	UplinkRSSI1   int32
	UplinkRSSI2   int32
	UplinkLQ      uint32
	UplinkSNR     int32
	DownlinkRSSI  int32
	DownlinkLQ    uint32
}

type BatteryData struct {
	Voltage   float32
	Current   float32
	Remaining float32
}

type GPSData struct {
	Latitude    float32
	Longitude   float32
	Altitude    int32
	Satellites  uint32
	GroundSpeed float32
}

type AttitudeData struct {
	Pitch float32
	Roll  float32
	Yaw   float32
}

type Controller struct {
	serialCtl *sc.Controller

	currentChannels *[16]util.CRSFValue
	channelsMutex   sync.RWMutex

	portState       PortState
	supervisorState SupervisorState

	sentPacketsCount  uint64
	recvPacketsCount  uint64
	errorPacketsCount uint64

	supervisorTomb *tomb.Tomb
	sendLoopTomb   *tomb.Tomb
	recvLoopTomb   *tomb.Tomb
	portLoopTomb   *tomb.Tomb

	// Simple telemetry channels
	LinkStatsChan  chan LinkStats
	BatteryChan    chan BatteryData
	GPSChan        chan GPSData
	AttitudeChan   chan AttitudeData

	sendChan chan any
	recvChan chan any
}

func NewCtl(sc *sc.Controller) *Controller {
	defaultChannels := &[16]util.CRSFValue{
		992, 992, 992, 992, 992, 992, 992, 992,
		992, 992, 992, 992, 992, 992, 992, 992,
	}

	linkCtl := &Controller{
		portState:       PortUnknown,
		supervisorState: SupervisorInactive,
		serialCtl:       sc,
		currentChannels: defaultChannels,
		// Create buffered channels for telemetry
		LinkStatsChan:   make(chan LinkStats, 10),
		BatteryChan:     make(chan BatteryData, 10),
		GPSChan:         make(chan GPSData, 10),
		AttitudeChan:    make(chan AttitudeData, 10),
	}

	return linkCtl
}

func (c *Controller) Init() error {
	return nil
}

func (c *Controller) Quit() {
	// Close telemetry channels
	close(c.LinkStatsChan)
	close(c.BatteryChan)
	close(c.GPSChan)
	close(c.AttitudeChan)
}

func (c *Controller) UpdateChannels(channels [16]util.CRSFValue) {
	c.channelsMutex.Lock()
	defer c.channelsMutex.Unlock()
	c.currentChannels = &channels
}

func (c *Controller) GetChannels() [16]util.CRSFValue {
	c.channelsMutex.RLock()
	defer c.channelsMutex.RUnlock()
	return *c.currentChannels
}

func (c *Controller) IsActive() bool {
	return c.supervisorState == SupervisorActive
}

// Add method to send initialization commands
func (c *Controller) SendModelID() {
	if c.sendChan != nil {
		c.sendChan <- SendModelId
	}
}

func (c *Controller) PingDevices() {
	if c.sendChan != nil {
		c.sendChan <- PingDevices
	}
}

// Send telemetry to channels (non-blocking)
func (c *Controller) sendLinkStats(stats LinkStats) {
	select {
	case c.LinkStatsChan <- stats:
	default:
		// Channel full, discard
	}
}

func (c *Controller) sendBattery(battery BatteryData) {
	select {
	case c.BatteryChan <- battery:
	default:
	}
}

func (c *Controller) sendGPS(gps GPSData) {
	select {
	case c.GPSChan <- gps:
	default:
	}
}

func (c *Controller) sendAttitude(attitude AttitudeData) {
	select {
	case c.AttitudeChan <- attitude:
	default:
	}
}
