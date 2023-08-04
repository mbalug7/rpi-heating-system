package config

import (
	"rpi-heating-system/lib/homeassistant"
	"rpi-heating-system/lib/homeassistant/model"
)

type AppConfig struct {
	Mqtt     *homeassistant.MqttConfig `json:"mqtt"`
	Gpiod    *Gpiod                    `json:"gpiod"`
	HADevice *model.Device             `json:"home_assistant_device"`
	Pumps    []*PumpConfig             `json:"pumps"`
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
