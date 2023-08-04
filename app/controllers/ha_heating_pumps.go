package controllers

import (
	"fmt"
	"rpi-heating-system/app/config"
	"rpi-heating-system/app/services"
	"rpi-heating-system/lib"
	"rpi-heating-system/lib/homeassistant/model"
	"time"

	"github.com/rs/zerolog/log"

	MQTT "github.com/eclipse/paho.mqtt.golang"

	jsoniter "github.com/json-iterator/go"
)

// HAHeatingPumpsHandler is the implementation of HAController interface
type HAHeatingPumpsHandler struct {
	client   MQTT.Client
	pumpsSvc services.PumpsService
	haDevice *model.Device
	pumpCfgs map[services.PumpID]*model.Switch
}

// NewHAHeatingPumpsHandler creates a new instance of HAHeatingPumpsHandler
func NewHAHeatingPumpsHandler(
	mqttClient MQTT.Client,
	conf *config.AppConfig,
	pumpSvc services.PumpsService,
) (*HAHeatingPumpsHandler, error) {

	h := &HAHeatingPumpsHandler{
		client:   mqttClient,
		pumpsSvc: pumpSvc,
		haDevice: conf.HADevice,
		pumpCfgs: make(map[services.PumpID]*model.Switch),
	}

	// build configs
	for _, pump := range conf.Pumps {
		pumpConf := h.getPumpConfig(pump)
		h.pumpCfgs[services.PumpID(pump.ID)] = pumpConf
	}

	// send configs to HA
	for _, pump := range h.pumpCfgs {
		err := h.sendConfig(pump)
		if err != nil {
			return nil, fmt.Errorf("failed to send config for pump %s, err: %w", pump.Name, err)
		}
		// Home Assistant is slow sometimes while processing new configs... wait a bit
		time.Sleep(100 * time.Millisecond)
	}

	// set all the pumps as available
	for _, pump := range h.pumpCfgs {
		token := h.client.Publish(pump.AvailabilityTopic, 0, true, "online")
		if !token.WaitTimeout(2 * time.Second) {
			return nil, fmt.Errorf("failed to update pump availability, %w", token.Error())
		}
	}

	// report pumps states
	for id, pump := range h.pumpCfgs {
		err := h.reportPumpState(services.PumpID(id), pump)
		if err != nil {
			return nil, fmt.Errorf("failed to report pump %s state, err: %w", pump.Name, err)
		}
	}

	// subscribe to HA commands
	for _, pump := range h.pumpCfgs {
		if token := h.client.Subscribe(pump.CommandTopic, 1, h.onHACommand); token.Wait() && token.Error() != nil {
			return nil, fmt.Errorf("failed to subscribe to pump command topic, %w", token.Error())
		}
	}

	return h, nil
}

// Close closes the HAHeatingPumpsHandler and performs necessary cleanup
func (obj *HAHeatingPumpsHandler) Close() error {
	err := obj.unsubscribeTopics()
	if err != nil {
		return fmt.Errorf("failed to disconnect: %w", err)
	}
	obj.client.Disconnect(100)
	return nil
}

// unsubscribeTopics unsubscribes client from MQTT topics
func (obj *HAHeatingPumpsHandler) unsubscribeTopics() error {
	for _, pump := range obj.pumpCfgs {
		if token := obj.client.Unsubscribe(pump.CommandTopic); token.Wait() && token.Error() != nil {
			return fmt.Errorf("failed to unsubscribe from command topic, %w", token.Error())
		}
	}
	return nil
}

// onHACommand is a callback function for processing Home Assistant commands
func (obj *HAHeatingPumpsHandler) onHACommand(client MQTT.Client, msg MQTT.Message) {
	log.Debug().Msgf("Rx HA: Topic: [%s], Payload: [%s]", msg.Topic(), msg.Payload())
	for id, pump := range obj.pumpCfgs {
		if pump.CommandTopic == msg.Topic() {
			state := services.PumpOFF
			if string(msg.Payload()) == "ON" {
				state = services.PumpON
			}
			err := obj.pumpsSvc.SetPumpState(services.PumpID(id), state)
			if err != nil {
				lib.Panic(fmt.Errorf("failed to set pump state: %s", err))

			}
			err = obj.reportPumpState(services.PumpID(id), pump)
			if err != nil {
				lib.Panic(fmt.Errorf("failed to report pump state: %s", err))
			}
		}
	}
}

// reportPumpState reports the current state of a pump to Home Assistant
func (obj *HAHeatingPumpsHandler) reportPumpState(pumpID services.PumpID, pump *model.Switch) error {
	state, err := obj.pumpsSvc.GetPumpState(pumpID)
	if err != nil {
		return fmt.Errorf("failed to update pump ON/OFF state for pump %s, err: %w", pump.Name, err)
	}
	msg := "OFF"
	if state == services.PumpON {
		msg = "ON"
	}
	return obj.sendFeedbackMessage(msg, pump.StateTopic)
}

// getPumpConfig creates a configuration for a pump
func (obj *HAHeatingPumpsHandler) getPumpConfig(pumpCfg *config.PumpConfig) *model.Switch {
	uid := fmt.Sprintf("heating_pump_%d", pumpCfg.ID)
	return &model.Switch{
		Schema:            "json",
		UniqueID:          uid,
		Name:              pumpCfg.Name,
		Device:            obj.haDevice,
		CommandTopic:      fmt.Sprintf("homeassistant/switch/%s/set", uid),
		StateTopic:        fmt.Sprintf("homeassistant/switch/%s/state", uid),
		AvailabilityTopic: fmt.Sprintf("homeassistant/switch/%s/status", uid),
	}
}

// sendFeedbackMessage sends feedback message to Home Assistant.
func (obj *HAHeatingPumpsHandler) sendFeedbackMessage(msg string, stateTopic string) error {
	token := obj.client.Publish(stateTopic, 0, true, msg)
	if !token.WaitTimeout(2 * time.Second) {
		return fmt.Errorf("failed to send feedback message to topic %s: %w", stateTopic, token.Error())
	}
	return nil
}

// sendConfig sends configuration to Home Assistant for a pump
func (obj *HAHeatingPumpsHandler) sendConfig(sw *model.Switch) error {
	conf, err := jsoniter.MarshalToString(sw)
	if err != nil {
		return err
	}
	token := obj.client.Publish(fmt.Sprintf("homeassistant/switch/%s/config", sw.UniqueID), 0, true, conf)
	if !token.WaitTimeout(10 * time.Second) {
		return fmt.Errorf("failed to send config %v: %w", sw, token.Error())
	}
	return nil
}
