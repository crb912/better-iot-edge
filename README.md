# Edge Gateway

基于 [EdgeX device-sdk-go v3](https://github.com/edgexfoundry/device-sdk-go) 的物联网边缘网关，
在单个 Device Service 中同时支持 **Modbus TCP 温度传感器** 和 **HTTP REST 湿度传感器**。

## 目录结构

```
edge-gateway/
├── cmd/
│   └── main.go               # 入口，调用 startup.Bootstrap
├── driver/
│   ├── composite.go          # 根据 protocols key 路由到子驱动
│   ├── modbus/
│   │   └── driver.go         # Modbus ProtocolDriver 实现
│   └── http/
│       └── driver.go         # HTTP ProtocolDriver 实现
├── internal/
│   └── transform/
│       ├── transform.go      # Modbus 字节 ↔ 数值转换 / 缩放
│       └── transform_test.go
├── res/
│   ├── profiles/
│   │   ├── temperature.yaml  # 温度传感器 Device Profile
│   │   └── humidity.yaml     # 湿度传感器 Device Profile
│   ├── devices/
│   │   └── device-list.yaml  # 静态设备预置 + AutoEvent 配置
│   └── configuration.yaml    # 服务主配置
├── Makefile
├── Dockerfile
└── docker-compose.yml        # 完整 EdgeX 开发环境
```

## 快速开始

### 前置条件

- Go 1.21+
- Docker & Docker Compose

### 1. 启动 EdgeX 核心服务

```bash
docker-compose up -d consul redis core-metadata core-data core-command
```

### 2. 修改设备地址

编辑 `res/devices/device-list.yaml`，将设备 IP 改为实际地址：

```yaml
# 温度传感器 Modbus TCP
Address: "192.168.1.10:502"

# 湿度传感器 HTTP
BaseURL: "http://192.168.1.20:8080"
```

### 3. 本地运行

```bash
make run
```

### 4. 验证数据采集

```bash
# 查看最新的 Event（需要 EdgeX 核心服务运行）
curl http://localhost:59880/api/v3/event/device/name/temperature-sensor-01?limit=5

# 通过 core-command 下发读取命令
curl http://localhost:59882/api/v3/device/name/temperature-sensor-01/command/readTemperature
```

## 配置说明

### 新增 Modbus 设备

在 `res/devices/device-list.yaml` 追加：

```yaml
- name: "temperature-sensor-02"
  profileName: "temperature-sensor"
  autoEvents:
    - interval: "5s"
      onChange: false
      sourceName: "readTemperature"
  protocols:
    modbus:
      Address: "192.168.1.11:502"
      SlaveID: "2"
```

### 修改采集频率

修改 `device-list.yaml` 中对应设备的 `autoEvents[].interval`，格式为 Go duration（如 `"5s"`、`"1m"`）。重启时加 `--overwriteDevices` 标志使修改生效。

### 添加新的 Device Resource

在对应的 Profile YAML 中追加 `deviceResources` 条目，指定 `modbusAddress` / `httpPath` 等属性。

## 协议驱动说明

### ModbusDriver 协议属性

| 属性          | 说明                              | 示例              |
|-------------|-----------------------------------|-------------------|
| `Address`   | Modbus TCP 服务器地址:端口         | `192.168.1.10:502`|
| `SlaveID`   | 从机地址（1–247）                  | `1`               |

### ModbusDriver 资源属性

| 属性                | 说明                            | 默认值     |
|-------------------|---------------------------------|-----------|
| `modbusFunction`  | `holding` / `input`            | `holding` |
| `modbusAddress`   | 寄存器偏移地址（十进制）          | `0`        |
| `modbusDataType`  | `float32/int16/uint16/int32/uint32` | `float32`|
| `scale`           | 数值缩放因子                     | `1.0`      |

### HttpDriver 协议属性

| 属性      | 说明                        | 示例                        |
|---------|-----------------------------|-----------------------------|
| `BaseURL` | 设备 HTTP 服务根地址        | `http://192.168.1.20:8080` |

### HttpDriver 资源属性

| 属性          | 说明                                    | 示例             |
|-------------|----------------------------------------|-----------------|
| `httpMethod`  | HTTP 方法                              | `GET`           |
| `httpPath`    | 请求路径                               | `/api/humidity` |
| `jsonPath`    | 点路径，从 JSON 响应中提取数值          | `data.humidity` |
| `scale`       | 数值缩放因子                           | `1.0`           |

## 运行测试

```bash
make test
make test-cover   # 生成 HTML 覆盖率报告
```

## 生产部署

```bash
# 构建并推送镜像
make docker
docker tag edge-gateway:dev registry.example.com/edge-gateway:1.0.0
docker push registry.example.com/edge-gateway:1.0.0
```
