// bertter-iot-edge/config/config.go

package config

import "fmt"

// Protocol constants
type Protocol int

const (
	Modbus Protocol = iota
	SNMP
	OPC
	EPS
	MTQQ
)

const (
	FormatJSON  Format = "json"
	FormatExcel Format = "excel"
)

const (
	modbusFilepath = "config.toml"
)

// DevicesList holds the device information.
type DevicesList struct {
	Devices []interface{} `json:"devices"`
}

// Devices holds data for a single register point.

// NewLoader creates a new loader based on the format.
func NewLoader(f Format) (Loader, error) {
	switch f {
	case FormatJSON:
		return &JSONLoader{}, nil
	case FormatExcel:
		return &ExcelLoader{}, nil
	default:
		return nil, fmt.Errorf("unsupported format: %s", f)
	}
}
