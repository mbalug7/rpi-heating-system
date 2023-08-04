package config

import (
	"fmt"
	"io"
	"os"

	jsoniter "github.com/json-iterator/go"
)

// LoadConfigFromFile reads the contents of the file specified by 'filename' and unmarshals the JSON data into the 'confStruct'.
// 'confStruct' is a pointer to the struct into which the JSON data will be unmarshaled.
// If the file cannot be opened or read, or if the JSON data cannot be unmarshaled, an error is returned.
func LoadConfigFromFile(filename string, confStruct any) error {
	data, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}
	defer data.Close()

	byteResult, err := io.ReadAll(data)
	if err != nil {
		return fmt.Errorf("failed to read config file data: %w", err)
	}
	err = jsoniter.Unmarshal(byteResult, confStruct)
	if err != nil {
		return fmt.Errorf("failed to unmarshal JSON data: %w", err)
	}

	return nil
}
