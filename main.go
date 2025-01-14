package main

import (
	"errors"
	"flag"
	"io"
	"rpi-heating-system/app/config"
	"rpi-heating-system/app/controllers"
	"rpi-heating-system/app/services"
	"rpi-heating-system/lib"
	"rpi-heating-system/lib/homeassistant"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/warthog618/gpiod"

	ConfLoader "rpi-heating-system/lib/config"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// 1. Create a channel for connection events
type onConnectEvent struct {
	mqttClient mqtt.Client
}
type onConnectionLostEvent struct {
	mqttClient mqtt.Client
	err        error
}

func main() {

	configPath := flag.String("config", "./app/config/config.json", "Path to the config file")
	flag.Parse()

	// Load the configuration from the specified file into 'conf' struct
	conf := &config.AppConfig{}
	err := ConfLoader.LoadConfigFromFile(*configPath, conf)
	lib.Panic(err)

	// Set the time format for logging using zerolog package
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	c, err := gpiod.NewChip(conf.Gpiod.Chip, gpiod.WithConsumer(conf.Gpiod.Consumer))
	lib.Panic(err)
	ts := services.NewDS18B20Service()

	var btnCtl *controllers.HAButtonsHandler
	var haPumpHandler *controllers.HAHeatingPumpsHandler
	var tempSensorsCtl *controllers.HATemperatureSensorsHandler
	var offlineController io.Closer

	// Create a new instance of the heating pumps handler service
	ps, err := services.NewHeatingPumpsHandler(c, conf.Pumps)
	lib.Panic(err)
	defer func() {
		err := ps.Close()
		if err != nil {
			log.Error().Msgf("failed to close pump service: %s", err)
		}
	}()

	bh, err := services.NewButtonHandler(c, conf.Buttons)
	lib.Panic(err)

	onConnectCh := make(chan onConnectEvent)
	onConnectionLostCh := make(chan onConnectionLostEvent)
	stopChan := make(chan struct{})

	go func(
		ts *services.DS18B20Service,
		btnCtl *controllers.HAButtonsHandler,
		haPumpHandler *controllers.HAHeatingPumpsHandler,
		tempSensorsCtl *controllers.HATemperatureSensorsHandler,
		offlineController io.Closer) {
		for {
			select {
			case evt, ok := <-onConnectCh:
				log.Info().Msg("MQTT connected event received")
				if !ok {
					lib.Panic(errors.New("connect channel closed"))
					return
				}
				// Stop the offline controller if it exists
				if offlineController != nil {
					offlineController.Close()
				}
				log.Info().Msg("MQTT connected!")

				bc, err := controllers.NewHAButtonsHandler(evt.mqttClient, conf, bh)
				lib.Panic(err)
				btnCtl = bc

				// Create a new instance of the Home Assistant heating pumps handler controller
				haPH, err := controllers.NewHAHeatingPumpsHandler(evt.mqttClient, conf, ps)
				lib.Panic(err)
				haPumpHandler = haPH

				tsCtl, err := controllers.NewHATemperatureSensorsHandler(evt.mqttClient, conf, ts)
				lib.Panic(err)
				tempSensorsCtl = tsCtl

			case _, ok := <-onConnectionLostCh:
				log.Info().Msg("MQTT connection lost event received")
				if !ok {
					lib.Panic(errors.New("connection lost channel closed"))
					return
				}

				if btnCtl != nil {
					err := btnCtl.Close()
					if err != nil {
						log.Error().Msgf("failed to close home assistant button controller: %s", err)
					}
				}
				btnCtl = nil
				if haPumpHandler != nil {
					err := haPumpHandler.Close()
					if err != nil {
						log.Error().Msgf("failed to close home assistant pump controller: %s", err)
					}
				}
				haPumpHandler = nil
				if tempSensorsCtl != nil {
					err := tempSensorsCtl.Close()
					if err != nil {
						log.Error().Msgf("failed to close home assistant temperature sensor controller: %s", err)
					}
				}
				tempSensorsCtl = nil

				if offlineController != nil {
					log.Warn().Msg("offline controller already running")
					continue
				}

				offlineController = controllers.NewOfflineController(conf.OfflineController, bh, ts, ps)
			case <-stopChan:
				// do any cleanup if needed
				return
			}
		}
	}(ts, btnCtl, haPumpHandler, tempSensorsCtl, offlineController)

	// 2. Define onConnect and onConnectionLost callbacks
	onConnect := func(client mqtt.Client) {
		log.Info().Msg("Connected to MQTT broker")
		onConnectCh <- onConnectEvent{client}

	}

	onConnectionLost := func(client mqtt.Client, err error) {
		log.Warn().Msgf("MQTT connection lost: %v", err)
		onConnectionLostCh <- onConnectionLostEvent{mqttClient: client, err: err}
	}

	var haMqttClient mqtt.Client

	// Attempt to create MQTT client with retries
	go func() {
		for {
			var err error
			haMqttClient, err = homeassistant.NewHAMqttClient(conf.Mqtt, onConnect, onConnectionLost)
			if err != nil {
				log.Error().Msgf("Failed to create MQTT client: %v", err)
				onConnectionLostCh <- onConnectionLostEvent{mqttClient: nil, err: err}
				time.Sleep(30 * time.Second)
				continue
			}
			log.Info().Msg("MQTT client created")
			break
		}
	}()

	// Wait for the quit signal to terminate the application
	lib.WaitForQuitSignal()
	close(stopChan)

	if haMqttClient != nil {
		haMqttClient.Disconnect(100)
	}
}
