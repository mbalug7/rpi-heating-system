package controllers

import (
	"rpi-heating-system/app/config"
	"rpi-heating-system/app/services"
	"time"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/rs/zerolog/log"
)

type OfflineController struct {
	conf                 *config.OfflineController
	buttonSvc            services.ButtonService
	tempSensorReader     services.TempSensorReader
	pumpsSvc             services.PumpsService
	lostConnectionTicker *time.Ticker
	lostConnectionQuit   chan struct{}
}

func NewOfflineController(
	conf *config.OfflineController,
	buttonSvc services.ButtonService,
	tempSensorReader services.TempSensorReader,
	pumpsSvc services.PumpsService,
) *OfflineController {
	if conf == nil {
		log.Error().Msg("offline controller config is nil")
		return nil
	}
	if buttonSvc == nil {
		log.Error().Msg("button service is nil")
		return nil
	}
	if tempSensorReader == nil {
		log.Error().Msg("temperature sensor reader is nil")
		return nil
	}
	if pumpsSvc == nil {
		log.Error().Msg("pumps service is nil")
		return nil
	}
	oc := &OfflineController{
		conf:             conf,
		buttonSvc:        buttonSvc,
		tempSensorReader: tempSensorReader,
		pumpsSvc:         pumpsSvc,
	}
	return oc
}

func (c *OfflineController) OnConnect(client MQTT.Client) error {
	return c.Close()
}

func (c *OfflineController) OnConnectionLost(client MQTT.Client) error {
	log.Warn().Msg("Offline controller - mqtt connection lost, starting offline control loop")
	c.start()
	return nil
}

func (c *OfflineController) Close() error {
	close(c.lostConnectionQuit)
	c.lostConnectionTicker.Stop()
	// Turn off all configured pumps first
	for _, pumpID := range c.conf.PumpIDs {
		c.pumpsSvc.SetPumpState(services.PumpID(pumpID), services.PumpOFF)
	}
	return nil
}

func (c *OfflineController) start() {
	c.lostConnectionTicker = time.NewTicker(10 * time.Minute)
	c.lostConnectionQuit = make(chan struct{})
	pumpIndex := 1 // reset or keep it if you want the cycling to continue from last index

	var tickerHandler = func() error {
		// Do your 30-second logic here
		// 1. Read the temperature
		owenTemp, readErr := c.tempSensorReader.Read(c.conf.SensorTriggerID)
		if readErr != nil {
			log.Error().Msgf("failed to read owen temperature: %s", readErr)
			return readErr
		}

		// 2. Only one pump ON logic
		//    - If temp > trigger_temperature, turn OFF all pumps
		//    - Else, turn ON exactly one pump in round-robin fashion
		if owenTemp < c.conf.TriggerTemperature {
			log.Info().Msgf("Temp=%.2f < %.2f, turning OFF all pumps",
				owenTemp,
				c.conf.TriggerTemperature)
			// Turn off all configured pumps
			for _, pumpID := range c.conf.PumpIDs {
				c.pumpsSvc.SetPumpState(services.PumpID(pumpID), services.PumpOFF)
			}
			return nil
		}

		// Turn off all configured pumps first
		for _, pumpID := range c.conf.PumpIDs {
			c.pumpsSvc.SetPumpState(services.PumpID(pumpID), services.PumpOFF)
		}

		// Turn on only the current pump in rotation
		currentPumpID := c.conf.PumpIDs[pumpIndex]
		log.Info().Msgf("Temp=%.2f >= %.2f, turning ON pump %d",
			owenTemp,
			c.conf.TriggerTemperature,
			currentPumpID)
		c.pumpsSvc.SetPumpState(services.PumpID(currentPumpID), services.PumpON)

		// Round-robin: next time we move to the next pump
		pumpIndex = (pumpIndex + 1) % len(c.conf.PumpIDs)
		return nil
	}
	err := tickerHandler()
	if err != nil {
		log.Error().Msgf("failed to handle lost connection: %s", err)
	}
	go func() {
		// On each tick, do your check
		for {
			select {
			case <-c.lostConnectionQuit:
				// We were told to stop
				log.Info().Msg("Stopping lost-connection temperature check loop.")
				return
			case <-c.lostConnectionTicker.C:
				err := tickerHandler()
				if err != nil {
					log.Error().Msgf("failed to handle lost connection: %s", err)
				}
			}
		}
	}()
}
