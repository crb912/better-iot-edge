package modbus

type DevicesList struct {
	Devices []Device `json:"devices"`
}
type Device struct {
	DeviceID   string     `json:"device_id"`
	DevcieName string     `json:"device_name"`
	Resources  []Resource `json:"resources"`
}

type Resource struct {
	Address  int    `json:"address"`
	DataType string `json:"data_type"` // e.g., "int16", "float32"
}

// 检查字段是否有效
func resCheck(res *Resource) bool {
	if res.Address < 0 || res.Address > 65535 {
		return false
	}
	return true
}
