package model

// BinarySensor represents a binary sensor entity in Home Assistant.
type BinarySensor struct {
	Schema            string  `json:"schema"` // e.g., "json"
	UniqueID          string  `json:"unique_id"`
	Name              string  `json:"name"`
	Device            *Device `json:"device,omitempty"`
	StateTopic        string  `json:"state_topic"`
	AvailabilityTopic string  `json:"availability_topic,omitempty"`
}
