package services

import (
	"fmt"
	"rpi-heating-system/lib/model"
	"sync"
)

type DeviceController interface {
	SetState(deviceID string, state bool) error
}

type ThermostatStorageRegistrar interface {
	Save(thermostat *model.Thermostat) error
}

type ThermostatController struct {
	thermostat         model.Thermostat
	currentTemperature float64
	deviceController   DeviceController
	storage            ThermostatStorageRegistrar
	isRunning          bool
	sync.RWMutex
}

func (t *ThermostatController) ChangeMode(mode model.ThermostatMode) {
	t.Lock()
	t.thermostat.Mode = mode
	t.storage.Save(&t.thermostat)
	t.Unlock()

	t.UpdateCurrentTemperature(t.currentTemperature)
}

func (t *ThermostatController) UpdateCurrentTemperature(currentTemp float64) {
	t.RLock()
	defer t.RUnlock()

	t.currentTemperature = currentTemp
	switch t.thermostat.Mode {
	case model.ThermostatModeOff:
		t.handleHeatModeOff()
	case model.ThermostatModeHeat:
		t.handleHeatMode(currentTemp)
	case model.ThermostatModeCool:
		t.handleCoolMode(currentTemp)
	}
}

func (t *ThermostatController) handleHeatModeOff() error {
	var errs error
	t.isRunning = false
	if t.thermostat.HeatingDeviceID != nil {
		err := t.deviceController.SetState(*t.thermostat.HeatingDeviceID, false)
		if err != nil {
			errs = err
		}
	}
	if t.thermostat.CoolingDeviceID != nil {
		err := t.deviceController.SetState(*t.thermostat.CoolingDeviceID, false)
		if err != nil {
			errs = fmt.Errorf("%v, %v", errs, err)
		}
	}
	return errs
}

func (t *ThermostatController) handleHeatMode(currentTemp float64) error {
	if t.thermostat.TargetTemperatureHeat == nil {
		return fmt.Errorf("TargetTemperatureHeat is not set")
	}
	if currentTemp < *t.thermostat.TargetTemperatureHeat {
		return t.deviceController.SetState(*t.thermostat.HeatingDeviceID, true)
	}
	return t.deviceController.SetState(*t.thermostat.HeatingDeviceID, false)
}

func (t *ThermostatController) handleCoolMode(currentTemp float64) error {
	if t.thermostat.TargetTemperatureCool == nil {
		return fmt.Errorf("TargetTemperatureCool is not set")
	}
	if currentTemp > *t.thermostat.TargetTemperatureCool {
		return t.deviceController.SetState(*t.thermostat.CoolingDeviceID, true)
	}
	return t.deviceController.SetState(*t.thermostat.CoolingDeviceID, false)
}

func (t *ThermostatController) IsRunning() bool {
	t.RLock()
	defer t.RUnlock()

	switch t.thermostat.Mode {
	case model.ThermostatModeOff:
		return false
	case model.ThermostatModeHeat:
		return t.thermostat.TargetTemperatureHeat != nil && t.currentTemperature < *t.thermostat.TargetTemperatureHeat
	case model.ThermostatModeCool:
		return t.thermostat.TargetTemperatureCool != nil && t.currentTemperature > *t.thermostat.TargetTemperatureCool
	}
	return false
}
