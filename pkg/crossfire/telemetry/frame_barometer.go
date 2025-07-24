// SPDX-FileCopyrightText: © 2023 OneEyeFPV oneeyefpv@gmail.com
// SPDX-License-Identifier: GPL-3.0-or-later
// SPDX-License-Identifier: FS-0.9-or-later

package telemetry

import (
	"fmt"
	"github.com/kaack/elrs-joystick-control/pkg/crossfire"
)

const BarometerFrameSize int = 7

type TelemBarometerType interface {
	TelemType
	Altitude() float32
}

type BarometerFrame struct {
	RawData []uint8
}

func (t *BarometerFrame) Addr() crossfire.Endpoint {
	return crossfire.Endpoint(t.RawData[0])
}
func (t *BarometerFrame) Type() crossfire.FrameType {
	return crossfire.FrameType(t.RawData[2])
}

func (t *BarometerFrame) Data() []uint8 {
	return t.RawData[2:]
}

func (t *BarometerFrame) Altitude() float32 {
	return BarometerAltitude(t.RawData[3:5])
}

func (t *BarometerFrame) String() string {
	return fmt.Sprintf("(barometer-frame) alt: %v", t.Altitude())
}
