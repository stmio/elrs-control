package settings

import (
	"github.com/kaack/elrs-joystick-control/pkg/crossfire/telemetry"
)

type FieldType interface {
	Name() string
	Type() telemetry.CRSFFieldType
	Id() uint32
	ParentId() uint32

	String() string
}
