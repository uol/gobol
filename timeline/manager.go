package timeline

import "fmt"

// Manager - the parent of all event managers
type Manager struct {
	defaultTags map[string]string
	transport   Transport
}

// NewManager - creates a timeline manager
func NewManager(transport Transport, backend *Backend, defaultTags map[string]string) (*Manager, error) {

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

	if defaultTags == nil {

		defaultTags = map[string]string{}
	}

	return &Manager{
		transport:   transport,
		defaultTags: defaultTags,
	}, nil
}

// Send - sends a event point
func (m *Manager) Send(point interface{}) error {

	fmt.Println("send manager")
	if point == nil {
		return fmt.Errorf("point is null")
	}

	if len(m.defaultTags) > 0 {

		fmt.Println("add tag")

		casted, ok := point.(*Point)
		if !ok {
			return fmt.Errorf("not a point type")
		}

		for t, v := range m.defaultTags {
			casted.Tags[t] = v
		}
	}

	fmt.Println("send transport")

	return m.transport.Send(point)
}

// Shutdown - shuts down the transport
func (m *Manager) Shutdown() {

	m.transport.Close()
}
