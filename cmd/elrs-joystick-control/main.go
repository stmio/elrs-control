// SPDX-FileCopyrightText: © 2023 OneEyeFPV oneeyefpv@gmail.com
// SPDX-License-Identifier: GPL-3.0-or-later
// SPDX-License-Identifier: FS-0.9-or-later

package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"

	cc "github.com/kaack/elrs-joystick-control/pkg/config"
	dc "github.com/kaack/elrs-joystick-control/pkg/devices"
	lc "github.com/kaack/elrs-joystick-control/pkg/link"
	sc "github.com/kaack/elrs-joystick-control/pkg/serial"
)

func main() {
    txPortName := flag.String("tx-port", "", "TX Serial port name")
    txBaudRate := flag.Int("tx-baud", 921600, "TX Serial port baud rate")
    configFile := flag.String("config", "", "Config JSON file path")
    flag.Parse()

    // Initialize controllers
    devicesCtl := dc.NewCtl()
    defer devicesCtl.Quit()

    configCtl := cc.NewCtl(devicesCtl)
    defer configCtl.Quit()

    serialCtl := sc.NewCtl()
    defer serialCtl.Quit()

    linkCtl := lc.NewCtl(devicesCtl, serialCtl, configCtl)
    defer linkCtl.Quit()

    // Load config if provided
    if *configFile != "" {
        // Load your config file
    }

    // Start the RF link if port specified
    if *txPortName != "" {
        if err := linkCtl.StartSupervisor(*txPortName, int32(*txBaudRate)); err != nil {
            panic(err)
        }
    }

    // Handle Ctrl-C
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, os.Interrupt)
    <-sigChan
    
    fmt.Println("Shutting down...")
}
Creating Your API
F
