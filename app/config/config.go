package config

import (
	"rpi-heating-system/lib/homeassistant"
	"rpi-heating-system/lib/homeassistant/model"
)

type AppConfig struct {
	Mqtt              *homeassistant.MqttConfig `json:"mqtt"`
	Gpiod             *Gpiod                    `json:"gpiod"`
	HADevice          *model.Device             `json:"home_assistant_device"`
	Pumps             []*PumpConfig             `json:"pumps"`
	TempSensors       []*TempSensorsConfig      `json:"temperature_sensors,omitempty"`
	Buttons           []*ButtonConfig           `json:"buttons,omitempty"`
	OfflineController *OfflineController        `json:"offline_controller,omitempty"`
}

type Gpiod struct {
	Chip     string `json:"chip"`
	Consumer string `json:"consumer"`
}

type PumpConfig struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	GpioStatePin int    `json:"gpio_state_pin"`
}

type TempSensorsConfig struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type ButtonConfig struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	GpioInputPin int    `json:"gpio_input_pin"`
	EnablePullUp bool   `json:"enable_pull_up"`
}

type OfflineController struct {
	PumpIDs            []int   `json:"pump_ids"`
	SensorTriggerID    string  `json:"sensor_trigger_id"`
	TriggerTemperature float64 `json:"trigger_temperature"`
}
