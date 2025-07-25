package main

// #include "callbacks.h"
import "C"
import (
	"fmt"
	lc "github.com/kaack/elrs-joystick-control/pkg/link"
	sc "github.com/kaack/elrs-joystick-control/pkg/serial"
	"github.com/kaack/elrs-joystick-control/pkg/util"
	"sync"
	"time"
	"unsafe"
)

// Global controller instance
var (
	controller *lc.Controller
	serialCtl  *sc.Controller
	mu         sync.Mutex
)

// Telemetry callback function types
type LinkStatsCallback func(rssi1, rssi2 int32, lq uint32, snr int32)
type BatteryCallback func(voltage, current, remaining float32)
type GPSCallback func(lat, lon float32, alt int32, sats uint32, speed float32)
type AttitudeCallback func(pitch, roll, yaw float32)

var (
	linkStatsCallback LinkStatsCallback
	batteryCallback   BatteryCallback
	gpsCallback       GPSCallback
	attitudeCallback  AttitudeCallback
)

//export elrs_init
func elrs_init(port *C.char, baudRate C.int) C.int {
	mu.Lock()
	defer mu.Unlock()

	if controller != nil {
		return -1 // Already initialized
	}

	portStr := C.GoString(port)

	// Initialize controllers
	serialCtl = sc.NewCtl()
	controller = lc.NewCtl(serialCtl)

	// Start the RF link
	if err := controller.StartSupervisor(portStr, int32(baudRate)); err != nil {
		fmt.Printf("Failed to start link: %s\n", err.Error())
		return -2
	}

	// Wait for link to be active
	timeout := time.After(5 * time.Second)
	for !controller.IsActive() {
		select {
		case <-timeout:
			fmt.Println("Timeout waiting for link")
			return -3
		default:
			time.Sleep(100 * time.Millisecond)
		}
	}

	// Start telemetry monitoring
	go telemetryMonitor()

	return 0 // Success
}

//export elrs_close
func elrs_close() {
	mu.Lock()
	defer mu.Unlock()

	if controller == nil {
		return
	}

	// Safe shutdown
	channels := [16]util.CRSFValue{
		992, 992, 0, 992, 0, 992, 992, 992,
		992, 992, 992, 992, 992, 992, 992, 992,
	}
	controller.UpdateChannels(channels)
	time.Sleep(100 * time.Millisecond)

	controller.StopSupervisor()
	controller.Quit()
	serialCtl.Quit()

	controller = nil
	serialCtl = nil
}

//export elrs_set_channels
func elrs_set_channels(channelData *C.ushort) {
	if controller == nil {
		return
	}

	// Convert C array to Go array
	channels := [16]util.CRSFValue{}
	cArray := (*[16]C.ushort)(unsafe.Pointer(channelData))
	for i := 0; i < 16; i++ {
		channels[i] = util.CRSFValue(cArray[i])
	}

	controller.UpdateChannels(channels)
}

//export elrs_set_channel
func elrs_set_channel(channel C.int, value C.ushort) {
	if controller == nil || channel < 0 || channel > 15 {
		return
	}

	channels := controller.GetChannels()
	channels[channel] = util.CRSFValue(value)
	controller.UpdateChannels(channels)
}

//export elrs_arm
func elrs_arm() {
	if controller == nil {
		return
	}

	channels := controller.GetChannels()
	channels[4] = 1984 // AUX1 high
	controller.UpdateChannels(channels)
}

//export elrs_disarm
func elrs_disarm() {
	if controller == nil {
		return
	}

	channels := controller.GetChannels()
	channels[4] = 0 // AUX1 low
	controller.UpdateChannels(channels)
}

//export elrs_is_active
func elrs_is_active() C.int {
	if controller == nil {
		return 0
	}
	if controller.IsActive() {
		return 1
	}
	return 0
}

// Telemetry callbacks - these will be called from Python
//export elrs_set_linkstats_callback
func elrs_set_linkstats_callback(fn unsafe.Pointer) {
	if fn == nil {
		linkStatsCallback = nil
		return
	}
	linkStatsCallback = func(rssi1, rssi2 int32, lq uint32, snr int32) {
		C.call_linkstats_callback(fn, C.int(rssi1), C.int(rssi2), C.uint(lq), C.int(snr))
	}
}

//export elrs_set_battery_callback
func elrs_set_battery_callback(fn unsafe.Pointer) {
	if fn == nil {
		batteryCallback = nil
		return
	}
	batteryCallback = func(voltage, current, remaining float32) {
		C.call_battery_callback(fn, C.float(voltage), C.float(current), C.float(remaining))
	}
}

//export elrs_set_gps_callback
func elrs_set_gps_callback(fn unsafe.Pointer) {
	if fn == nil {
		gpsCallback = nil
		return
	}
	gpsCallback = func(lat, lon float32, alt int32, sats uint32, speed float32) {
		C.call_gps_callback(fn, C.float(lat), C.float(lon), C.int(alt), C.uint(sats), C.float(speed))
	}
}

//export elrs_set_attitude_callback
func elrs_set_attitude_callback(fn unsafe.Pointer) {
	if fn == nil {
		attitudeCallback = nil
		return
	}
	attitudeCallback = func(pitch, roll, yaw float32) {
		C.call_attitude_callback(fn, C.float(pitch), C.float(roll), C.float(yaw))
	}
}

func telemetryMonitor() {
	for controller != nil {
		select {
		case linkStats := <-controller.LinkStatsChan:
			if linkStatsCallback != nil {
				linkStatsCallback(linkStats.UplinkRSSI1, linkStats.UplinkRSSI2,
					linkStats.UplinkLQ, linkStats.UplinkSNR)
			}

		case battery := <-controller.BatteryChan:
			if batteryCallback != nil {
				batteryCallback(battery.Voltage, battery.Current, battery.Remaining)
			}

		case gps := <-controller.GPSChan:
			if gpsCallback != nil {
				gpsCallback(gps.Latitude, gps.Longitude, gps.Altitude,
					gps.Satellites, gps.GroundSpeed)
			}

		case attitude := <-controller.AttitudeChan:
			if attitudeCallback != nil {
				attitudeCallback(attitude.Pitch, attitude.Roll, attitude.Yaw)
			}

		case <-time.After(1 * time.Second):
			// Timeout to check if controller still exists
			if controller == nil {
				return
			}
		}
	}
}

func main() {}
