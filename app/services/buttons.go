package services

import (
	"fmt"
	"rpi-heating-system/app/config"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/warthog618/gpiod"
)

// SubscriptionID represents the unique identifier for a subscription
type SubscriptionID string

// BtnPressSubscription represents a subscription for button press events
type BtnPressSubscription struct {
	SID     SubscriptionID
	GpioPin int
	EventCh chan gpiod.LineEvent
}

// ButtonService is an interface that defines the operations for button press events
type ButtonService interface {
	SubscribeOnButtonPress(gpio int, observerIdentifier string) (*BtnPressSubscription, error)
	Unsubscribe(subscriptionID SubscriptionID) error
}

// ButtonHandler represents a handler for button press events using GPIO lines
type ButtonHandler struct {
	observers map[SubscriptionID]*BtnPressSubscription
}

// NewButtonHandler creates a new ButtonHandler with the given GPIO configuration and button configurations
func NewButtonHandler(gpiodChip *gpiod.Chip, btnsCfg []*config.ButtonConfig) (*ButtonHandler, error) {

	bh := &ButtonHandler{
		observers: make(map[SubscriptionID]*BtnPressSubscription),
	}
	for _, btn := range btnsCfg {
		l, err := gpiodChip.RequestLine(
			btn.GpioInputPin,
			gpiod.WithBothEdges,
			gpiod.WithEventHandler(bh.eventHandler),
			gpiod.WithDebounce(20*time.Millisecond),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to request %s GPIO line %d: %w", btn.Name, btn.GpioInputPin, err)
		}
		if btn.EnablePullUp {
			l.Reconfigure(gpiod.WithPullUp)
		}
	}

	return bh, nil
}

// SubscribeOnButtonPress subscribes to button press events for the specified GPIO pin
func (obj *ButtonHandler) SubscribeOnButtonPress(gpio int, observerIdentifier string) (*BtnPressSubscription, error) {
	log.Info().Msgf("SubscribeOnButtonPress: %v", gpio)

	sid := SubscriptionID(fmt.Sprintf("btn-%s-%d", observerIdentifier, gpio))

	// check if sid already exists in the map
	if _, ok := obj.observers[sid]; ok {
		return nil, fmt.Errorf("observer with id %s already exists", sid)
	}

	ch := make(chan gpiod.LineEvent)
	sub := &BtnPressSubscription{
		SID:     sid,
		GpioPin: gpio,
		EventCh: ch,
	}

	obj.observers[sid] = sub
	return sub, nil
}

// Unsubscribe removes the subscription with the specified ID
func (obj *ButtonHandler) Unsubscribe(subscriptionID SubscriptionID) error {
	log.Info().Msgf("Unsubscribe: %v", subscriptionID)

	// check if sid already exists in the map
	if _, ok := obj.observers[subscriptionID]; !ok {
		return fmt.Errorf("observer with id %s does not exist", subscriptionID)
	}
	// close channel and remove from map
	close(obj.observers[subscriptionID].EventCh)
	delete(obj.observers, subscriptionID)
	return nil
}

// eventHandler handles button press events
func (obj *ButtonHandler) eventHandler(evt gpiod.LineEvent) {
	log.Info().Msgf("button eventHandler: %v", evt)
	for _, sub := range obj.observers {
		if sub.GpioPin == evt.Offset {
			log.Info().Msgf("publishing event: %v to sub: %v", evt, sub)
			sub.EventCh <- evt
		}
	}
}
