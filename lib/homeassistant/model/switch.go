package model

// SwitchType represents the type of the switch device in Home Assistant
type SwitchType string

// Constants representing the possible switch types
const (
	OutletSwitch  SwitchType = "outlet" // Represents an outlet switch type
	GenericSwitch SwitchType = "switch" // Represents a generic switch type
)

// Switch represents a switch entity in Home Assistant.
type Switch struct {
	Schema            string     `json:"schema"`                       // Schema type for the switch entity
	Device            *Device    `json:"device,omitempty"`             // Associated device information
	Name              string     `json:"name"`                         // Name of the switch entity
	StateTopic        string     `json:"state_topic"`                  // MQTT topic to publish the switch state
	CommandTopic      string     `json:"command_topic"`                // MQTT topic to receive switch commands
	UniqueID          string     `json:"unique_id,omitempty"`          // Unique ID for the switch entity
	DeviceClass       SwitchType `json:"device_class,omitempty"`       // Type of the switch device
	AvailabilityTopic string     `json:"availability_topic,omitempty"` // MQTT topic to publish availability status
}

// HASwitchMessage represents the state message for a switch entity in Home Assistant.
type HASwitchMessage struct {
	State string `json:"state,omitempty"` // State of the switch entity ("ON" or "OFF")
}
