package homeassistant

import (
	"fmt"
	"rpi-heating-system/lib"

	MQTT "github.com/eclipse/paho.mqtt.golang"
)

// MqttConfig holds the configuration settings for the MQTT client used in Home Assistant communication
type MqttConfig struct {
	Addr     string `json:"addr"`     // Address of the MQTT broker, e.g., "tcp://192.168.0.100:1883"
	Username string `json:"username"` // Username for MQTT authentication
	Password string `json:"password"` // Password for MQTT authentication
}

// NewHAMqttClient creates a new MQTT client for communication with Home Assistant
// It takes the MqttConfig as input and returns an MQTT.Client instance or an error if connection fails
func NewHAMqttClient(conf *MqttConfig) (MQTT.Client, error) {
	// Create MQTT client options and set the provided configuration settings
	opts := MQTT.NewClientOptions().AddBroker(conf.Addr)
	opts.SetUsername(conf.Username)
	opts.SetPassword(conf.Password)
	opts.SetClientID("rpi-heating-controller")
	opts.OnConnectionLost = func(client MQTT.Client, err error) {
		// If the MQTT connection is lost, panic with the error using the Panic utility function from the lib package
		lib.Panic(fmt.Errorf("mqtt-event connection lost %w", err))
	}
	client := MQTT.NewClient(opts)

	// Connect to the MQTT broker
	token := client.Connect()
	if token.Wait(); token.Error() != nil {
		return nil, fmt.Errorf("failed to connect to MQTT broker, %w", token.Error())
	}

	// Return the MQTT client instance if the connection is successful
	return client, nil
}
