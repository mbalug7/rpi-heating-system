package model

type TemperatureSensor struct {
	Schema            string  `json:"schema"` // e.g., "json"
	UniqueID          string  `json:"unique_id"`
	Name              string  `json:"name"`
	Device            *Device `json:"device,omitempty"`
	StateTopic        string  `json:"state_topic"`
	AvailabilityTopic string  `json:"availability_topic,omitempty"`
	UnitOfMeasurement string  `json:"unit_of_measurement,omitempty"`
}
