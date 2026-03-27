// internal/transform/transform.go
// transform 包提供 Modbus 寄存器字节与 Go 数值之间的双向转换，
// 以及通用的数值缩放和类型强转工具。
package transform

import (
	"encoding/binary"
	"fmt"
	"math"
	"strconv"
)

// ParseScale 将字符串解析为缩放因子；解析失败时返回 1.0（不缩放）。
func ParseScale(s string) float64 {
	v, err := strconv.ParseFloat(s, 64)
	if err != nil || v == 0 {
		return 1.0
	}
	return v
}

// DecodeModbusBytes 将 Modbus 读寄存器返回的字节序列解码为 float64。
//
// dataType 支持：
//   - "float32"  : 4 字节 IEEE 754 Big-Endian float
//   - "int16"    : 2 字节有符号整数
//   - "uint16"   : 2 字节无符号整数
//   - "int32"    : 4 字节有符号整数
//   - "uint32"   : 4 字节无符号整数
func DecodeModbusBytes(data []byte, dataType string) (float64, error) {
	switch dataType {
	case "float32":
		if len(data) < 4 {
			return 0, fmt.Errorf("float32 requires 4 bytes, got %d", len(data))
		}
		bits := binary.BigEndian.Uint32(data[:4])
		return float64(math.Float32frombits(bits)), nil

	case "int16":
		if len(data) < 2 {
			return 0, fmt.Errorf("int16 requires 2 bytes, got %d", len(data))
		}
		raw := int16(binary.BigEndian.Uint16(data[:2]))
		return float64(raw), nil

	case "uint16":
		if len(data) < 2 {
			return 0, fmt.Errorf("uint16 requires 2 bytes, got %d", len(data))
		}
		return float64(binary.BigEndian.Uint16(data[:2])), nil

	case "int32":
		if len(data) < 4 {
			return 0, fmt.Errorf("int32 requires 4 bytes, got %d", len(data))
		}
		raw := int32(binary.BigEndian.Uint32(data[:4]))
		return float64(raw), nil

	case "uint32":
		if len(data) < 4 {
			return 0, fmt.Errorf("uint32 requires 4 bytes, got %d", len(data))
		}
		return float64(binary.BigEndian.Uint32(data[:4])), nil

	default:
		return 0, fmt.Errorf("unsupported dataType: %s", dataType)
	}
}

// EncodeModbusBytes 将 float64 编码为 Modbus 写寄存器所需的字节序列。
func EncodeModbusBytes(value float64, dataType string) ([]byte, error) {
	switch dataType {
	case "float32":
		bits := math.Float32bits(float32(value))
		buf := make([]byte, 4)
		binary.BigEndian.PutUint32(buf, bits)
		return buf, nil

	case "int16":
		buf := make([]byte, 2)
		binary.BigEndian.PutUint16(buf, uint16(int16(value)))
		return buf, nil

	case "uint16":
		buf := make([]byte, 2)
		binary.BigEndian.PutUint16(buf, uint16(value))
		return buf, nil

	case "int32":
		buf := make([]byte, 4)
		binary.BigEndian.PutUint32(buf, uint32(int32(value)))
		return buf, nil

	case "uint32":
		buf := make([]byte, 4)
		binary.BigEndian.PutUint32(buf, uint32(value))
		return buf, nil

	default:
		return nil, fmt.Errorf("unsupported dataType: %s", dataType)
	}
}

// ToFloat64 将任意数值类型强转为 float64，用于写命令参数解包。
func ToFloat64(v interface{}) (float64, error) {
	switch val := v.(type) {
	case float64:
		return val, nil
	case float32:
		return float64(val), nil
	case int:
		return float64(val), nil
	case int8:
		return float64(val), nil
	case int16:
		return float64(val), nil
	case int32:
		return float64(val), nil
	case int64:
		return float64(val), nil
	case uint:
		return float64(val), nil
	case uint8:
		return float64(val), nil
	case uint16:
		return float64(val), nil
	case uint32:
		return float64(val), nil
	case uint64:
		return float64(val), nil
	case string:
		return strconv.ParseFloat(val, 64)
	default:
		return 0, fmt.Errorf("cannot convert %T to float64", v)
	}
}
