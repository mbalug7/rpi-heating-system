package controllers

import (
	"fmt"
	"rpi-heating-system/app/config"
	"rpi-heating-system/app/services"
	"rpi-heating-system/lib/homeassistant/model"
	"time"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	jsoniter "github.com/json-iterator/go"
	"github.com/rs/zerolog/log"
	"github.com/warthog618/gpiod"
)

// HAButtonsHandler is the implementation of HAController interface
type HAButtonsHandler struct {
	client           MQTT.Client
	buttonSvc        services.ButtonService
	haDevice         *model.Device
	buttonsCfgs      map[int]*model.BinarySensor
	svcSubscriptions []*services.BtnPressSubscription
}

// NewHAButtonsHandler creates a new instance of HAButtonsHandler
func NewHAButtonsHandler(
	mqttClient MQTT.Client,
	conf *config.AppConfig,
	buttonSvc services.ButtonService,
) (*HAButtonsHandler, error) {

	h := &HAButtonsHandler{
		client:      mqttClient,
		buttonSvc:   buttonSvc,
		haDevice:    conf.HADevice,
		buttonsCfgs: make(map[int]*model.BinarySensor),
	}

	for _, button := range conf.Buttons {
		subs, err := buttonSvc.SubscribeOnButtonPress(button.GpioInputPin, fmt.Sprintf("%d", button.ID))
		if err != nil {
			return nil, fmt.Errorf("failed to subscribe to button %s, err: %w", button.Name, err)
		}
		h.svcSubscriptions = append(h.svcSubscriptions, subs)
		log.Info().Msgf("Listening for button %s events", button.Name)

		// start a goroutine to listen for button events
		go func() {
			for {
				event, ok := <-subs.EventCh
				if !ok {
					return
				}
				msg := "OFF"
				if event.Type == gpiod.LineEventFallingEdge {
					msg = "ON"
				}
				btnCfg, ok := h.buttonsCfgs[event.Offset]
				if !ok {
					log.Error().Msgf("Button %d not found in config", event.Offset)
					continue
				}
				h.sendFeedbackMessage(msg, btnCfg.StateTopic)
			}
		}()
	}

	// build configs
	for _, button := range conf.Buttons {
		buttonConf := h.getButtonConfig(button)
		h.buttonsCfgs[button.GpioInputPin] = buttonConf
	}

	// send configs
	for _, button := range h.buttonsCfgs {
		err := h.sendConfig(button)
		if err != nil {
			return nil, fmt.Errorf("failed to send config for button %s, err: %w", button.UniqueID, err)
		}
		// Home Assistant is slow sometimes while processing new configs... wait a bit
		time.Sleep(100 * time.Millisecond)
	}

	// set all the buttons as available
	for _, button := range h.buttonsCfgs {
		token := h.client.Publish(button.AvailabilityTopic, 0, true, "online")
		if !token.WaitTimeout(2 * time.Second) {
			return nil, fmt.Errorf("failed to update button %s availability, %w", button.UniqueID, token.Error())
		}
	}

	return h, nil
}

// Close closes the HAButtonsHandler and performs necessary cleanup
func (h *HAButtonsHandler) Close() error {
	for _, sub := range h.svcSubscriptions {
		err := h.buttonSvc.Unsubscribe(sub.SID)
		if err != nil {
			return fmt.Errorf("failed to unsubscribe from button %s, err: %w", sub.SID, err)
		}
	}
	return nil
}

// getButtonConfig creates a configuration for a button
func (h *HAButtonsHandler) getButtonConfig(button *config.ButtonConfig) *model.BinarySensor {
	uid := fmt.Sprintf("button_%d", button.ID)
	return &model.BinarySensor{
		Name:              button.Name,
		UniqueID:          uid,
		Device:            h.haDevice,
		StateTopic:        fmt.Sprintf("homeassistant/binary_sensor/%s/state", uid),
		AvailabilityTopic: fmt.Sprintf("homeassistant/binary_sensor/%s/availability", uid),
	}
}

// sendFeedbackMessage sends feedback message to Home Assistant.
func (obj *HAButtonsHandler) sendFeedbackMessage(msg string, stateTopic string) error {
	token := obj.client.Publish(stateTopic, 0, true, msg)
	if !token.WaitTimeout(2 * time.Second) {
		return fmt.Errorf("failed to send feedback message to topic %s: %w", stateTopic, token.Error())
	}
	return nil
}

// sendConfig sends configuration to Home Assistant for a pump
func (obj *HAButtonsHandler) sendConfig(sw *model.BinarySensor) error {
	conf, err := jsoniter.MarshalToString(sw)
	if err != nil {
		return err
	}
	token := obj.client.Publish(fmt.Sprintf("homeassistant/binary_sensor/%s/config", sw.UniqueID), 0, true, conf)
	if !token.WaitTimeout(10 * time.Second) {
		return fmt.Errorf("failed to send config %v: %w", sw, token.Error())
	}
	return nil
}
