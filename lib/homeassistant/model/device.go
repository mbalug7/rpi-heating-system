package model

// Device represents a physical device or entity in Home Assistant
type Device struct {
	Identifiers   []string `json:"identifiers,omitempty"`
	Manufacturer  string   `json:"manufacturer,omitempty"`
	Model         string   `json:"model,omitempty"`
	Name          string   `json:"name,omitempty"`
	SwVersion     string   `json:"sw_version,omitempty"`
	SuggestedArea string   `json:"suggested_area,omitempty"`
}
