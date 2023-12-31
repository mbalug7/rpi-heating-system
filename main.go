package main

import (
	"flag"
	"rpi-heating-system/app/config"
	"rpi-heating-system/app/controllers"
	"rpi-heating-system/app/services"
	"rpi-heating-system/lib"
	"rpi-heating-system/lib/homeassistant"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/warthog618/gpiod"

	ConfLoader "rpi-heating-system/lib/config"
)

func main() {

	configPath := flag.String("config", "/home/pi/config.json", "Path to the config file")
	flag.Parse()

	// Load the configuration from the specified file into 'conf' struct
	conf := &config.AppConfig{}
	err := ConfLoader.LoadConfigFromFile(*configPath, conf)
	lib.Panic(err)

	// Set the time format for logging using zerolog package
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	// Create a new instance of the Home Assistant MQTT client
	haMqttClient, err := homeassistant.NewHAMqttClient(conf.Mqtt)
	lib.Panic(err)
	defer haMqttClient.Disconnect(100)

	c, err := gpiod.NewChip(conf.Gpiod.Chip, gpiod.WithConsumer(conf.Gpiod.Consumer))
	lib.Panic(err)

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

	btnSvc, err := controllers.NewHAButtonsHandler(haMqttClient, conf, bh)
	lib.Panic(err)
	defer func() {
		err := btnSvc.Close()
		if err != nil {
			log.Error().Msgf("failed to close home assistant button controller: %s", err)
		}
	}()

	// Create a new instance of the Home Assistant heating pumps handler controller
	haPumpHandler, err := controllers.NewHAHeatingPumpsHandler(haMqttClient, conf, ps)
	lib.Panic(err)

	defer func() {
		err := haPumpHandler.Close()
		if err != nil {
			log.Error().Msgf("failed to close home assistant pump controller: %s", err)
		}
	}()

	ts := services.NewDS18B20Service()

	tempSensorsCtl, err := controllers.NewHATemperatureSensorsHandler(haMqttClient, conf, ts)
	lib.Panic(err)

	defer func() {
		err := tempSensorsCtl.Close()
		if err != nil {
			log.Error().Msgf("failed to close home assistant temperature sensor controller: %s", err)
		}
	}()

	// Wait for the quit signal to terminate the application
	lib.WaitForQuitSignal()
}
