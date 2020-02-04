package client

import (
	"errors"

	"autonomia.digital/tonio/app/hosting"
	"autonomia.digital/tonio/app/tor"
)

var (
	// ErrNoClient is throwed when no available client has been initialized
	ErrNoClient = errors.New("error: no client to run")
	// ErrNoService is used when the Tor service can't be started
	ErrNoService = errors.New("error: the service can't be started")
)

// Service is a representation of a Mumble service through Tor
type Service interface {
	tor.Service
}

// LaunchClient executes the current Mumble client instance
func LaunchClient(data hosting.MeetingData, onClose func()) (Service, error) {
	c := System()

	if !c.CanBeUsed() {
		return nil, ErrNoClient
	}

	cm := tor.Command{
		Cmd:      c.GetBinaryPath(),
		Args:     []string{hosting.GenerateURL(data)},
		Modifier: c.GetTorCommandModifier(),
	}

	s, err := tor.NewService(cm)
	if err != nil {
		return nil, ErrNoService
	}

	if onClose != nil {
		s.OnClose(onClose)
	}

	return s, nil
}
