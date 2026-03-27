package transform

import (
	"math"
	"testing"
)

func TestDecodeModbusBytes_Float32(t *testing.T) {
	// 25.0 °C 的 IEEE 754 big-endian 字节：0x41C80000
	data := []byte{0x41, 0xC8, 0x00, 0x00}
	got, err := DecodeModbusBytes(data, "float32")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if math.Abs(got-25.0) > 0.001 {
		t.Errorf("expected 25.0, got %f", got)
	}
}

func TestDecodeModbusBytes_Int16(t *testing.T) {
	// -1 的 int16 big-endian
	data := []byte{0xFF, 0xFF}
	got, err := DecodeModbusBytes(data, "int16")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != -1.0 {
		t.Errorf("expected -1.0, got %f", got)
	}
}

func TestDecodeModbusBytes_Uint16(t *testing.T) {
	data := []byte{0x01, 0x2C} // 300
	got, err := DecodeModbusBytes(data, "uint16")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != 300.0 {
		t.Errorf("expected 300.0, got %f", got)
	}
}

func TestDecodeModbusBytes_UnsupportedType(t *testing.T) {
	_, err := DecodeModbusBytes([]byte{0x00}, "double")
	if err == nil {
		t.Fatal("expected error for unsupported type")
	}
}

func TestEncodeDecodeRoundTrip(t *testing.T) {
	cases := []struct {
		name     string
		dataType string
		value    float64
	}{
		{"float32 round trip", "float32", 23.5},
		{"int16 positive", "int16", 100},
		{"int16 negative", "int16", -50},
		{"uint16", "uint16", 1000},
		{"uint32", "uint32", 65536},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			encoded, err := EncodeModbusBytes(tc.value, tc.dataType)
			if err != nil {
				t.Fatalf("encode error: %v", err)
			}
			decoded, err := DecodeModbusBytes(encoded, tc.dataType)
			if err != nil {
				t.Fatalf("decode error: %v", err)
			}
			if math.Abs(decoded-tc.value) > 0.01 {
				t.Errorf("round trip mismatch: expected %f, got %f", tc.value, decoded)
			}
		})
	}
}

func TestParseScale(t *testing.T) {
	if ParseScale("0.1") != 0.1 {
		t.Error("expected 0.1")
	}
	if ParseScale("invalid") != 1.0 {
		t.Error("invalid should return 1.0")
	}
	if ParseScale("0") != 1.0 {
		t.Error("zero should return 1.0")
	}
}
