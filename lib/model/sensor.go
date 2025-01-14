package model

type SensorType string

func (s SensorType) String() string {
	return string(s)
}

const (
	SensorTypeRemote SensorType = "remote_sensor"
	SensorTypeLocal  SensorType = "local_sensor"
)

type Sensor struct {
	ID         string            `json:"id"`
	Name       string            `json:"name"`
	Type       SensorType        `json:"type"`
	Attributes map[string]string `json:"attributes"`
}
