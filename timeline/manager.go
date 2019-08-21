package timeline

import (
	"fmt"

	json "github.com/uol/serializer/json"
)

/**
* Manages the transport and backend configuration.
* @author rnojiri
**/

// Manager - the parent of all event managers
type Manager struct {
	transport Transport
}

// NewManager - creates a timeline manager
func NewManager(transport Transport, backend *Backend) (*Manager, error) {

	if transport == nil {
		return nil, fmt.Errorf("transport implementation is required")
	}

	if backend == nil {
		return nil, fmt.Errorf("no backend configuration was found")
	}

	err := transport.ConfigureBackend(backend)
	if err != nil {
		return nil, err
	}

	return &Manager{
		transport: transport,
	}, nil
}

// SendNumberPoint - sends a number point
func (m *Manager) SendNumberPoint(parameters ...interface{}) error {

	m.transport.PointChannel() <- json.Parameters{
		Name:       numberPointJSON,
		Parameters: parameters,
	}

	return nil
}

// SendTextPoint - sends a text point
func (m *Manager) SendTextPoint(parameters ...interface{}) error {

	m.transport.PointChannel() <- json.Parameters{
		Name:       textPointJSON,
		Parameters: parameters,
	}

	return nil
}

func (m *Manager) AggregateNumberPoint() {

}

// Start - starts th transport
func (m *Manager) Start() error {

	return m.transport.Start()
}

// Shutdown - shuts down the transport
func (m *Manager) Shutdown() {

	m.transport.Close()
}

// GetTransport - returns the configured transport
func (m *Manager) GetTransport() Transport {

	return m.transport
}
