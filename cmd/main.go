// cmd/main.go
// 程序入口，将两个协议驱动组合后交给 EdgeX device-sdk-go 的 Bootstrap 启动。
package main

import (
	"better-iot-edge/internal/driver"
	httpdriver "better-iot-edge/internal/driver/http"
	modbusdriver "better-iot-edge/internal/driver/modbus"

	"github.com/edgexfoundry/device-sdk-go/v2/pkg/startup"
)

const (
	serviceName    = "edge-gateway"
	serviceVersion = "1.0.0"
)

func main() {
	modbusDrv := modbusdriver.NewDriver()
	httpDrv := httpdriver.NewDriver()

	composite := driver.NewCompositeDriver(modbusDrv, httpDrv)

	startup.Bootstrap(serviceName, serviceVersion, composite)
}
