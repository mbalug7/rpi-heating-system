package homeassistant

import "io"

// HAController is an interface for the Home Assistant controller
type HAController interface {
	io.Closer
}
