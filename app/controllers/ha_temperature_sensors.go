package controllers

import (
	"fmt"
	"rpi-heating-system/app/config"
	"rpi-heating-system/app/services"
	"rpi-heating-system/lib/homeassistant/model"
	"strconv"
	"time"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	jsoniter "github.com/json-iterator/go"
	"github.com/rs/zerolog/log"
)

type HATemperatureSensorsHandler struct {
	client           MQTT.Client
	haDevice         *model.Device
	tempSensorReader services.TempSensorReader
	sensorCfgs       map[string]*model.TemperatureSensor
	ticker           *time.Ticker
}

func NewHATemperatureSensorsHandler(
	mqttClient MQTT.Client,
	conf *config.AppConfig,
	tempSensorReader services.TempSensorReader,
) (*HATemperatureSensorsHandler, error) {
	log.Debug().Msg("Creating Temp sensor HA handler")
	h := &HATemperatureSensorsHandler{
		client:           mqttClient,
		tempSensorReader: tempSensorReader,
		haDevice:         conf.HADevice,
		sensorCfgs:       make(map[string]*model.TemperatureSensor),
	}

	// build configs
	for _, sensor := range conf.TempSensors {
		sensorConf := h.getSensorConfig(sensor)
		h.sensorCfgs[sensor.ID] = sensorConf
	}
	log.Debug().Msgf("Sending sensor configs %+v", h.sensorCfgs)
	// send configs
	for _, sensor := range h.sensorCfgs {
		err := h.sendConfig(sensor)
		if err != nil {
			return nil, fmt.Errorf("failed to send config for sensor %s, err: %w", sensor.UniqueID, err)
		}
		// Home Assistant is slow sometimes while processing new configs... wait a bit
		time.Sleep(100 * time.Millisecond)
	}

	// set all the sensors as available
	for _, sensor := range h.sensorCfgs {
		token := h.client.Publish(sensor.AvailabilityTopic, 0, true, "online")
		if !token.WaitTimeout(2 * time.Second) {
			return nil, fmt.Errorf("failed to update sensor %s availability, %w", sensor.UniqueID, token.Error())
		}
	}

	for id, sensor := range h.sensorCfgs {
		err := h.reportSensorTemperature(id, sensor)
		if err != nil {
			return nil, fmt.Errorf("failed to report sensor %s temperature, err: %w", sensor.UniqueID, err)
		}
	}

	ticker := time.NewTicker(30 * time.Second)
	// Create a goroutine to read sensor values periodically
	go func() {
		for range ticker.C {
			// Report sensor state
			for id, sensor := range h.sensorCfgs {
				err := h.reportSensorTemperature(id, sensor)
				if err != nil {
					log.Error().Msgf("failed to report sensor %s temperature: %s", sensor.UniqueID, err)
				}
			}
		}
	}()

	// Store the ticker in the struct to be able to stop it later
	h.ticker = ticker

	return h, nil
}

// Close closes the HATemperatureSensorsHandler and performs necessary cleanup
func (obj *HATemperatureSensorsHandler) Close() error {
	obj.ticker.Stop()
	return nil
}

// getPumpConfig creates a configuration for a pump
func (obj *HATemperatureSensorsHandler) getSensorConfig(cfg *config.TempSensorsConfig) *model.TemperatureSensor {
	uid := fmt.Sprintf("temp_%s", cfg.ID)
	return &model.TemperatureSensor{
		Schema:            "json",
		UniqueID:          uid,
		Name:              cfg.Name,
		Device:            obj.haDevice,
		StateTopic:        fmt.Sprintf("homeassistant/sensor/%s/state", uid),
		AvailabilityTopic: fmt.Sprintf("homeassistant/sensor/%s/status", uid),
		UnitOfMeasurement: "°C",
	}
}

func (obj *HATemperatureSensorsHandler) reportSensorTemperature(ID string, sensor *model.TemperatureSensor) error {
	temp, err := obj.tempSensorReader.Read(ID)
	if err != nil {
		return fmt.Errorf("failed to read temperature for sensor %s, err: %w", ID, err)
	}
	log.Debug().Msgf("Reporting temperature for sensor %s, temp %f[]", ID, temp)
	return obj.sendFeedbackMessage(strconv.FormatFloat(temp, 'f', 2, 64), sensor.StateTopic)
}

// sendFeedbackMessage sends feedback message to Home Assistant.
func (obj *HATemperatureSensorsHandler) sendFeedbackMessage(msg string, stateTopic string) error {
	token := obj.client.Publish(stateTopic, 0, true, msg)
	if !token.WaitTimeout(2 * time.Second) {
		return fmt.Errorf("failed to send feedback message to topic %s: %w", stateTopic, token.Error())
	}
	return nil
}

// sendConfig sends configuration to Home Assistant for a sensor
func (obj *HATemperatureSensorsHandler) sendConfig(sensor *model.TemperatureSensor) error {
	conf, err := jsoniter.MarshalToString(sensor)
	if err != nil {
		return err
	}
	token := obj.client.Publish(fmt.Sprintf("homeassistant/sensor/%s/config", sensor.UniqueID), 0, true, conf)
	if !token.WaitTimeout(10 * time.Second) {
		return fmt.Errorf("failed to send config %v: %w", sensor, token.Error())
	}
	return nil
}
