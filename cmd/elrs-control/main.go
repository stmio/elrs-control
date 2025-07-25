package main

import (
	"flag"
	"fmt"
	lc "github.com/kaack/elrs-joystick-control/pkg/link"
	sc "github.com/kaack/elrs-joystick-control/pkg/serial"
	"github.com/kaack/elrs-joystick-control/pkg/util"
	"os"
	"os/signal"
	"time"
)

func main() {
	// Command line flags
	txPortName := flag.String("port", "", "Serial port name (e.g., /dev/ttyUSB0, COM3)")
	txBaudRate := flag.Int("baud", 921600, "Serial port baud rate")
	flag.Parse()

	if *txPortName == "" {
		fmt.Println("Error: Serial port is required")
		flag.Usage()
		os.Exit(1)
	}

	// Initialize controllers
	serialCtl := sc.NewCtl()
	defer serialCtl.Quit()

	linkCtl := lc.NewCtl(serialCtl)
	defer linkCtl.Quit()

	// Start the RF link
	fmt.Printf("Starting RF link on %s at %d baud...\n", *txPortName, *txBaudRate)
	if err := linkCtl.StartSupervisor(*txPortName, int32(*txBaudRate)); err != nil {
		fmt.Printf("Failed to start link: %s\n", err.Error())
		os.Exit(1)
	}

	// Wait for link to be active
	fmt.Println("Waiting for link...")
	timeout := time.After(5 * time.Second)
	for !linkCtl.IsActive() {
	    select {
	    case <-timeout:
		fmt.Println("Timeout waiting for link to become active")
		os.Exit(1)
	    default:
		time.Sleep(100 * time.Millisecond)
	    }
	}
	fmt.Println("Link active!")

	// Give the handshake a moment to complete
	time.Sleep(500 * time.Millisecond)

	// Set up telemetry monitoring (optional)
	go monitorTelemetry(linkCtl)

	// Handle Ctrl-C
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)

	// Example control loop
	go controlLoop(linkCtl)

	// Wait for interrupt
	<-sigChan
	fmt.Println("\nShutting down...")

	// Safe shutdown - disarm and zero throttle
	channels := [16]util.CRSFValue{
		992, 992, 0, 992, 0, 992, 992, 992,
		992, 992, 992, 992, 992, 992, 992, 992,
	}
	linkCtl.UpdateChannels(channels)
	time.Sleep(100 * time.Millisecond)

	// Stop the link
	if err := linkCtl.StopSupervisor(); err != nil {
		fmt.Printf("Error stopping link: %s\n", err.Error())
	}
}

func monitorTelemetry(linkCtl *lc.Controller) {
	for {
		select {
		case linkStats := <-linkCtl.LinkStatsChan:
			fmt.Printf("Link: RSSI=%d/%d LQ=%d%% SNR=%d\n", 
				linkStats.UplinkRSSI1, linkStats.UplinkRSSI2, 
				linkStats.UplinkLQ, linkStats.UplinkSNR)

		case battery := <-linkCtl.BatteryChan:
			fmt.Printf("Battery: %.1fV %.1fA %d%%\n", 
				battery.Voltage, battery.Current, int(battery.Remaining))

		case gps := <-linkCtl.GPSChan:
			fmt.Printf("GPS: %.6f,%.6f Alt=%dm Sats=%d Speed=%.1fm/s\n",
				gps.Latitude, gps.Longitude, gps.Altitude, 
				gps.Satellites, gps.GroundSpeed)

		case attitude := <-linkCtl.AttitudeChan:
			fmt.Printf("Attitude: Pitch=%.1f° Roll=%.1f° Yaw=%.1f°\n",
				attitude.Pitch, attitude.Roll, attitude.Yaw)
		}
	}
}

func controlLoop(linkCtl *lc.Controller) {
	// Example: Simple hover pattern
	channels := [16]util.CRSFValue{
		992, 992, 992, 992, 992, 992, 992, 992,
		992, 992, 992, 992, 992, 992, 992, 992,
	}

	// Wait a bit before arming
	time.Sleep(2 * time.Second)

	// Arm (assuming AUX1 is arm)
	fmt.Println("Arming...")
	channels[4] = 1984 // AUX1 high
	linkCtl.UpdateChannels(channels)
	time.Sleep(2 * time.Second)

	// Throttle up slowly
	fmt.Println("Throttle up...")
	for throttle := 992; throttle < 1400; throttle += 10 {
		channels[2] = util.CRSFValue(throttle)
		linkCtl.UpdateChannels(channels)
		time.Sleep(50 * time.Millisecond)
	}

	// Hover
	fmt.Println("Hovering...")
	time.Sleep(5 * time.Second)

	// Throttle down
	fmt.Println("Landing...")
	for throttle := 1400; throttle > 992; throttle -= 10 {
		channels[2] = util.CRSFValue(throttle)
		linkCtl.UpdateChannels(channels)
		time.Sleep(50 * time.Millisecond)
	}

	// Disarm
	fmt.Println("Disarming...")
	channels[4] = 0 // AUX1 low
	channels[2] = 0 // Zero throttle
	linkCtl.UpdateChannels(channels)
}
