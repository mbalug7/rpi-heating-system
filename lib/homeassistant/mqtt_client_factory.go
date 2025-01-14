package homeassistant

import (
	"fmt"
	"time"

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
func NewHAMqttClient(
	conf *MqttConfig,
	onConnect func(MQTT.Client),
	onConnectionLost func(MQTT.Client, error),
) (MQTT.Client, error) {

	opts := MQTT.NewClientOptions().AddBroker(conf.Addr)
	opts.SetUsername(conf.Username)
	opts.SetPassword(conf.Password)
	opts.SetClientID("rpi-heating-controller")
	opts.SetAutoReconnect(true)
	opts.SetConnectRetry(true)
	opts.SetConnectRetryInterval(30 * time.Second)

	// The library calls OnConnectionLost automatically when
	// there's a disconnection (and auto reconnect is enabled).
	// We just forward that event through our callback.
	opts.OnConnectionLost = onConnectionLost

	// The library also calls OnConnect after each (re)connection.
	// We forward it as well.
	opts.OnConnect = onConnect

	client := MQTT.NewClient(opts)
	token := client.Connect()
	if !token.WaitTimeout(10 * time.Second) {
		defer client.Disconnect(250)
		if token.Error() != nil {

			return nil, fmt.Errorf("failed to connect to MQTT broker: %w", token.Error())
		}
		return nil, fmt.Errorf("failed to connect to MQTT broker, timeout")
	}

	return client, nil
}
