package embassy

import (
	"errors"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/hashicorp/consul/api"
)

// ErrConsulInvalidPath happens when a given KV path is not present
var ErrConsulInvalidPath = errors.New("Consul invalid path")

// KVSettings is the configuration for the KV module
type KVSettings struct {
	Path string
}

// NewKV returns a MKV object
func NewKV(log *logrus.Logger, settings KVSettings) (*MKV, error) {
	if consulClient == nil {
		return nil, errors.New("No agent connection found")
	}

	kv := &MKV{
		log:      log,
		settings: settings,
		client:   consulClient,
		skac:     make(chan struct{}),
		swsc:     make(chan struct{}),
	}
	return kv, nil
}

// MKV is a wrapper around consul's KV API
type MKV struct {
	log      *logrus.Logger
	settings KVSettings
	client   *api.Client
	watching bool
	poolInt  time.Duration
	sw       ServiceWatcher
	skac     chan struct{}
	swsc     chan struct{}
}

// Put adds a value to the KV store
func (mkv *MKV) Put(path string, value []byte) error {
	lf := map[string]interface{}{
		"struct": "MKV",
		"func":   "Put",
	}
	kv := mkv.client.KV()

	p := &api.KVPair{Key: path, Value: value}

	m, err := kv.Put(p, nil)
	if err != nil {
		mkv.log.WithFields(lf).Error(err)
		return err
	}
	mkv.log.WithFields(lf).Info("requestTime=", m.RequestTime)

	return nil
}

// Get retrieves a value from consul
func (mkv *MKV) Get(path string) ([]byte, error) {
	lf := map[string]interface{}{
		"struct": "MKV",
		"func":   "Get",
	}

	kv := mkv.client.KV()

	pair, m, err := kv.Get(path, nil)
	if err != nil {
		mkv.log.WithFields(lf).Error(err)
		return nil, err
	}
	mkv.log.WithFields(lf).Info("requestTime=", m.RequestTime)

	if pair == nil {
		return nil, ErrConsulInvalidPath
	}

	return pair.Value, nil
}
