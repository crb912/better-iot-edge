package config

import (
	"encoding/json"
	"errors"
	"io"
	"os"
)

// Format defines the source file type.
type Format string

type JSONLoader struct{}
type ExcelLoader struct{}

// ValidatorFunc defines a generic validation function.
type ValidatorFunc func() error

// Loader defines how to load device configs.
type Loader interface {
	// Load reads data from io.Reader and returns a list of configs.
	Load(r io.Reader) ([]*DevicesList, error)
}

// Unmarshal loads config into the dest interface.
// It accepts optional validator functions.
func Unmarshal(p Protocol, dest any, validators ...ValidatorFunc) error {
	var data []byte

	// 1. Fetch raw data based on protocol.
	switch p {
	case Modbus:
		data = []byte(`[{"device_id":"1","device_name":"modbus_dev","resources":[{"address":100,"data_type":"int16"}]}]`)
	case SNMP:
		data = []byte(`[{"device_id":"2","device_name":"snmp_dev","resources":[{"address":200,"data_type":"float32"}]}]`)
	default:
		return errors.New("unsupported protocol")
	}

	// 2. Decode JSON into the generic destination.
	if err := json.Unmarshal(data, dest); err != nil {
		return err
	}

	// 3. Run all provided validators.
	for _, validate := range validators {
		if err := validate(); err != nil {
			return err
		}
	}

	return nil
}

// UnmarshalFromfile loads config from a specific file path.
// UnmarshalFromfile loads config from a file.
func UnmarshalFromfile(path string, dest any, validators ...ValidatorFunc) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(data, dest); err != nil {
		return err
	}

	for _, validate := range validators {
		if err := validate(); err != nil {
			return err
		}
	}

	return nil
}

// MarshalToFile saves the config struct to a specific file.
func MarshalToFile(path string, src interface{}) error {
	// 1. Convert struct to JSON format.
	data, err := json.MarshalIndent(src, "", "  ")
	if err != nil {
		return err
	}

	// 2. Write data to file.
	return os.WriteFile(path, data, 0644)
}

// Load implements the Loader interface for JSON.
func (l *JSONLoader) Load(r io.Reader) ([]*DevicesList, error) {
	var configs []*DevicesList
	decoder := json.NewDecoder(r)

	// Decode JSON data into configs array.
	if err := decoder.Decode(&configs); err != nil {
		return nil, err
	}

	return configs, nil
}

// ExcelLoader loads config from Excel format.

// Load implements the Loader interface for Excel.
func (l *ExcelLoader) Load(r io.Reader) ([]*DevicesList, error) {
	var configs []*DevicesList

	// Example using excelize:
	// f, err := excelize.OpenReader(r)
	// if err != nil { return nil, err }
	// rows, _ := f.GetRows("Sheet1")
	// Loop through rows to build your configs array...

	return configs, nil
}
