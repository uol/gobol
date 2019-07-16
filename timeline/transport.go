package timeline

// Transport - the implementation type to send a event
type Transport interface {

	// Send - send a new event using the parent Point interface
	Send(point interface{}) error

	// ConfigureBackend - configures the backend
	ConfigureBackend(backend *Backend) error

	// Close - closes this transport
	Close()
}
