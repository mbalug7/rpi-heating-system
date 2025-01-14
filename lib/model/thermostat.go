package model

type ThermostatMode string

const (
	ThermostatModeOff  ThermostatMode = "off"
	ThermostatModeAuto ThermostatMode = "auto"
	ThermostatModeHeat ThermostatMode = "heat"
	ThermostatModeCool ThermostatMode = "cool"
)

type Thermostat struct {
	ID                    string         `json:"id"`
	Name                  string         `json:"name"`
	Mode                  ThermostatMode `json:"mode"`
	SensorID              string         `json:"sensor_id"`
	MinTemperature        *float64       `json:"min_temperature"`
	MaxTemperature        *float64       `json:"max_temperature"`
	TargetTemperatureHeat *float64       `json:"target_temperature_heat"`
	TargetTemperatureCool *float64       `json:"target_temperature_cool"`
	HeatingDeviceID       *string        `json:"heating_device_id"`
	CoolingDeviceID       *string        `json:"cooling_device_id"`
	Priority              *Priority      `json:"priority"`
}
