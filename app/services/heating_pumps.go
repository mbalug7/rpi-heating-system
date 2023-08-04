package services

import (
	"fmt"
	"io"
	"rpi-heating-system/app/config"

	"github.com/rs/zerolog/log"
	"github.com/warthog618/gpiod"
)

// PumpID represents the unique identifier for a pump
type PumpID int

// PumpState represents the state of a pump (ON or OFF)
type PumpState int

const (
	PumpOFF PumpState = 0 // PumpOFF indicates the pump is turned off.
	PumpON  PumpState = 1 // PumpON indicates the pump is turned on.
)

// Value returns the integer value of the PumpState
func (ps PumpState) Value() int {
	return int(ps)
}

// PumpsService is an interface that defines the operations for controlling pumps
// It provides methods to set and get the state of a pump and also implements the io.Closer interface for cleanup
type PumpsService interface {
	SetPumpState(pump PumpID, state PumpState) error
	GetPumpState(pump PumpID) (PumpState, error)
	io.Closer
}

// HeatingPumpsHandler represents a handler for controlling pumps using GPIO lines
// It implements the PumpsService interface for pump control and cleanup.
type HeatingPumpsHandler struct {
	pumps map[PumpID]*gpiod.Line
}

// NewHAHeatingHeatingPumpsHandler creates a new HeatingPumpsHandler with the given GPIO configuration and pump configurations
// It requests GPIO lines for each pump and initializes the HeatingPumpsHandler with these lines
func NewHeatingPumpsHandler(gpiodCfg *config.Gpiod, pumpsCfg []*config.PumpConfig) (*HeatingPumpsHandler, error) {

	ph := &HeatingPumpsHandler{
		pumps: make(map[PumpID]*gpiod.Line),
	}
	c, err := gpiod.NewChip(gpiodCfg.Chip, gpiod.WithConsumer(gpiodCfg.Consumer))
	if err != nil {
		return nil, fmt.Errorf("failed to create GPIO chip: %w", err)
	}

	for _, pump := range pumpsCfg {
		ph.pumps[PumpID(pump.ID)], err = c.RequestLine(
			pump.GpioStatePin,
			gpiod.AsOutput(0),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to request %s GPIO line %d: %w", pump.Name, pump.GpioStatePin, err)
		}
	}
	return ph, nil
}

// Close closes the GPIO lines for all the pumps
func (obj *HeatingPumpsHandler) Close() error {
	for _, line := range obj.pumps {
		err := line.Close()
		if err != nil {
			return fmt.Errorf("failed to close gpio line: %w", err)
		}
	}
	return nil
}

// SetPumpState sets the state of the pump with the specified ID
func (obj *HeatingPumpsHandler) SetPumpState(pumpID PumpID, state PumpState) error {
	p, ok := obj.pumps[pumpID]
	if !ok {
		return fmt.Errorf("pump %d does not exist", pumpID)
	}
	log.Debug().Msgf("Setting pump state, ID %d, state %d", pumpID, state.Value())

	err := p.SetValue(state.Value())
	if err != nil {
		return fmt.Errorf("failed to set pump state")
	}
	return nil
}

// GetPumpState returns the current state of the pump with the specified ID
func (obj *HeatingPumpsHandler) GetPumpState(pumpID PumpID) (PumpState, error) {
	p, ok := obj.pumps[pumpID]
	if !ok {
		return 0, fmt.Errorf("pump %d does not exist", pumpID)
	}
	val, err := p.Value()
	if err != nil {
		return 0, fmt.Errorf("failed to get pump state")
	}
	return PumpState(val), nil
}
