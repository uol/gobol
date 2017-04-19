package rubber

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
)

const (
	// ConfigWeightedBackend is the constant that should appear in the configuration to create
	// a weighted backend
	ConfigWeightedBackend = "weighted"

	// ConfigSimpleBackend is the constant that should appear in the configuration to create
	// a simple backend
	ConfigSimpleBackend = "simple"
)

func compare(name, constant string) bool {
	return strings.ToLower(name) == strings.ToLower(constant)
}

func makeBackend(log *logrus.Logger, settings Settings) (Backend, error) {
	if compare(settings.Type, ConfigWeightedBackend) {
		log.Debugln("Using weighted backend")
		return CreateWeightedBackend(settings, log)
	}
	if compare(settings.Type, ConfigSimpleBackend) || compare(settings.Type, "") {
		log.Debugln("Using simple backend")
		nodes := []string{settings.Preferred}
		nodes = append(nodes, settings.Nodes...)
		return &singleServerBackend{
			log:     log,
			nodes:   nodes,
			timeout: settings.Timeout * time.Second,

			client: &http.Client{
				Timeout: settings.Timeout * time.Second,
			},
		}, nil
	}
	return nil, errors.New("Unknown backend")
}

// Create is a function that creates an elasticsearch client
func Create(backend Backend) *Elastic {
	return &Elastic{Backend: backend}
}

// New creates an elasticsearch client
func New(log *logrus.Logger, settings Settings) (*Elastic, error) {
	backend, err := makeBackend(log, settings)
	if err != nil {
		return nil, err
	}
	return Create(backend), nil
}
