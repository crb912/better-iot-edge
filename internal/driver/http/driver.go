// internal/driver/http/driver.go
// HttpDriver 实现 EdgeX ProtocolDriver 接口，通过 HTTP REST 与湿度传感器通信。
//
// 协议属性（device-list.yaml protocols.http 节）：
//
//	BaseURL  string  设备 HTTP 服务根地址，如 "http://192.168.1.20:8080"
//
// 设备资源属性（humidity.yaml deviceResources[].attributes 节）：
//
//	httpMethod   string  HTTP 方法：GET（默认）/ POST
//	httpPath     string  请求路径，如 "/api/humidity"
//	jsonPath     string  从响应 JSON 中提取数值的点路径，如 "data.humidity"
//	scale        string  数值缩放因子（浮点），默认 "1.0"
//
// Driver 配置（configuration.yaml [Driver] 节）：
//
//	HttpTimeout  string  请求超时（Go duration），默认 "5s"
//	HttpRetries  string  失败重试次数，默认 "2"

package http

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/edgexfoundry/device-sdk-go/v2/pkg/interfaces"
	sdkModels "github.com/edgexfoundry/device-sdk-go/v2/pkg/models"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/models"

	"better-iot-edge/internal/transform"
)

type Driver struct {
	sdk     interfaces.DeviceServiceSDK
	client  *http.Client
	timeout time.Duration
	retries int
}

func NewDriver() *Driver {
	return &Driver{
		timeout: 5 * time.Second,
		retries: 2,
	}
}

// ---------- 生命周期 ----------

func (d *Driver) Initialize(sdk interfaces.DeviceServiceSDK) error {
	d.sdk = sdk

	driverConfigs := sdk.DriverConfigs()
	if v, ok := driverConfigs["HttpTimeout"]; ok {
		if dur, err := time.ParseDuration(v); err == nil {
			d.timeout = dur
		}
	}
	if v, ok := driverConfigs["HttpRetries"]; ok {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			d.retries = n
		}
	}

	d.client = &http.Client{Timeout: d.timeout}
	sdk.GetLoggingClient().Infof("HttpDriver initialized: timeout=%s retries=%d", d.timeout, d.retries)
	return nil
}

func (d *Driver) Start() error { return nil }

func (d *Driver) Stop(_ bool) error {
	d.sdk.GetLoggingClient().Info("HttpDriver stopped")
	return nil
}

// ---------- 读命令 ----------

func (d *Driver) HandleReadCommands(
	deviceName string,
	protocols map[string]models.ProtocolProperties,
	reqs []sdkModels.CommandRequest,
) ([]*sdkModels.CommandValue, error) {
	baseURL, err := extractBaseURL(protocols)
	if err != nil {
		return nil, fmt.Errorf("device %s: %w", deviceName, err)
	}

	results := make([]*sdkModels.CommandValue, 0, len(reqs))
	for _, req := range reqs {
		cv, err := d.readResource(deviceName, baseURL, req)
		if err != nil {
			return nil, fmt.Errorf("device %s resource %s: %w", deviceName, req.DeviceResourceName, err)
		}
		results = append(results, cv)
	}
	return results, nil
}

func (d *Driver) readResource(deviceName, baseURL string, req sdkModels.CommandRequest) (*sdkModels.CommandValue, error) {
	attrs := req.Attributes
	method := attrStr(attrs, "httpMethod", "GET")
	path := attrStr(attrs, "httpPath", "/")
	jsonPath := attrStr(attrs, "jsonPath", "")
	scale := transform.ParseScale(attrStr(attrs, "scale", "1.0"))

	url := strings.TrimRight(baseURL, "/") + "/" + strings.TrimLeft(path, "/")

	var (
		body []byte
		err  error
	)
	for attempt := 0; attempt <= d.retries; attempt++ {
		body, err = d.doRequest(method, url)
		if err == nil {
			break
		}
		if attempt < d.retries {
			d.sdk.GetLoggingClient().Warnf("HttpDriver: device %s attempt %d/%d failed: %v", deviceName, attempt+1, d.retries, err)
			time.Sleep(500 * time.Millisecond)
		}
	}
	if err != nil {
		return nil, fmt.Errorf("HTTP %s %s: %w", method, url, err)
	}

	floatVal, err := extractJSONValue(body, jsonPath)
	if err != nil {
		return nil, fmt.Errorf("JSON extract %q: %w", jsonPath, err)
	}
	floatVal *= scale

	// v2 用 sdkModels.NewCommandValue 构造 CommandValue
	cv, err := sdkModels.NewCommandValue(req.DeviceResourceName, common.ValueTypeFloat64, floatVal)
	if err != nil {
		return nil, err
	}
	return cv, nil
}

func (d *Driver) doRequest(method, url string) ([]byte, error) {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")

	resp, err := d.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("unexpected HTTP status: %d", resp.StatusCode)
	}
	return io.ReadAll(resp.Body)
}

// ---------- 写命令 ----------

func (d *Driver) HandleWriteCommands(
	deviceName string,
	protocols map[string]models.ProtocolProperties,
	reqs []sdkModels.CommandRequest,
	params []*sdkModels.CommandValue,
) error {
	baseURL, err := extractBaseURL(protocols)
	if err != nil {
		return fmt.Errorf("device %s: %w", deviceName, err)
	}

	for i, req := range reqs {
		if err := d.writeResource(deviceName, baseURL, req, params[i]); err != nil {
			return fmt.Errorf("device %s resource %s: %w", deviceName, req.DeviceResourceName, err)
		}
	}
	return nil
}

func (d *Driver) writeResource(deviceName, baseURL string, req sdkModels.CommandRequest, cv *sdkModels.CommandValue) error {
	path := attrStr(req.Attributes, "httpPath", "/")
	url := strings.TrimRight(baseURL, "/") + "/" + strings.TrimLeft(path, "/")

	// v2 用 cv.Float64Value() 取值
	value, err := cv.Float64Value()
	if err != nil {
		return err
	}

	body := strings.NewReader(fmt.Sprintf(`{"value":%f}`, value))
	httpReq, err := http.NewRequest(http.MethodPost, url, body)
	if err != nil {
		return err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := d.client.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("device %s: HTTP POST %s returned %d", deviceName, url, resp.StatusCode)
	}
	return nil
}

// ---------- 设备回调 ----------

func (d *Driver) AddDevice(deviceName string, _ map[string]models.ProtocolProperties, _ models.AdminState) error {
	d.sdk.GetLoggingClient().Infof("HttpDriver: device added: %s", deviceName)
	return nil
}

func (d *Driver) UpdateDevice(deviceName string, _ map[string]models.ProtocolProperties, _ models.AdminState) error {
	d.sdk.GetLoggingClient().Infof("HttpDriver: device updated: %s", deviceName)
	return nil
}

func (d *Driver) RemoveDevice(deviceName string, _ map[string]models.ProtocolProperties) error {
	d.sdk.GetLoggingClient().Infof("HttpDriver: device removed: %s", deviceName)
	return nil
}

func (d *Driver) Discover() error { return nil }

func (d *Driver) ValidateDevice(device models.Device) error {
	_, err := extractBaseURL(device.Protocols)
	return err
}

// ---------- 工具函数 ----------

func extractBaseURL(protocols map[string]models.ProtocolProperties) (string, error) {
	p, ok := protocols["http"]
	if !ok {
		return "", fmt.Errorf("missing 'http' protocol section")
	}
	url, ok := p["BaseURL"]
	if !ok || url == "" {
		return "", fmt.Errorf("http.BaseURL is required")
	}
	return url, nil
}

func extractJSONValue(data []byte, path string) (float64, error) {
	var raw interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return 0, fmt.Errorf("JSON unmarshal: %w", err)
	}
	if path == "" {
		return toFloat(raw)
	}
	current := raw
	for _, key := range strings.Split(path, ".") {
		m, ok := current.(map[string]interface{})
		if !ok {
			return 0, fmt.Errorf("path %q: expected object at key %q", path, key)
		}
		current, ok = m[key]
		if !ok {
			return 0, fmt.Errorf("path %q: key %q not found", path, key)
		}
	}
	return toFloat(current)
}

func toFloat(v interface{}) (float64, error) {
	switch val := v.(type) {
	case float64:
		return val, nil
	case float32:
		return float64(val), nil
	case int:
		return float64(val), nil
	case json.Number:
		return val.Float64()
	default:
		return 0, fmt.Errorf("cannot convert %T to float64", v)
	}
}

func attrStr(attrs map[string]interface{}, key, def string) string {
	if v, ok := attrs[key]; ok {
		if s, ok := v.(string); ok && s != "" {
			return s
		}
	}
	return def
}
