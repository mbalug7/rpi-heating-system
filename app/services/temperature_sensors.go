package services

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/rs/zerolog/log"
)

type TempSensorReader interface {
	Read(id string) (float64, error)
}

type DS18B20Service struct {
}

func NewDS18B20Service() *DS18B20Service {
	return &DS18B20Service{}
}

func (obj *DS18B20Service) Read(id string) (float64, error) {
	log.Debug().Msgf("Reading sensor state %s", id)
	path := filepath.Join("/sys/bus/w1/devices", id, "w1_slave")
	data, err := os.Open(path)
	if err != nil {
		return 0.0, fmt.Errorf("failed to read sensor temp on path %s err: %w", path, err)
	}

	defer data.Close()

	bytes, err := io.ReadAll(data)
	if err != nil {
		return 0.0, fmt.Errorf("failed to read loaded file on path %s err: %w", path, err)
	}

	raw := string(bytes)

	if !strings.Contains(raw, " YES") {
		return 0.0, fmt.Errorf("check CRC failed for sensor %s err: %w", id, err)
	}

	i := strings.LastIndex(raw, "t=")
	if i == -1 {
		return 0.0, fmt.Errorf("temperature value not exist for sensor %s err: %w", id, err)
	}

	c, err := strconv.ParseFloat(raw[i+2:len(raw)-1], 64)
	if err != nil {
		return 0.0, fmt.Errorf("failed to parse temperature for sensor %s err: %w", id, err)
	}

	return c / 1000.0, nil
}
