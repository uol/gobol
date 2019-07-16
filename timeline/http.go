package timeline

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/uol/gobol/util"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// HTTPTransport - implements the HTTP transport
type HTTPTransport struct {
	Transport
	httpClient        *http.Client
	batchSendInterval time.Duration
	pointChannel      chan interface{}
	serviceEndpoint   string
	serviceURL        string
	logger            *zap.Logger
}

// HTTPTransportConfig - has all HTTP event manager configurations
type HTTPTransportConfig struct {
	BatchSendInterval string
	BufferSize        int
	ServiceEndpoint   string
	RequestTimeout    string
	logger            *zap.Logger
}

// NewHTTPTransport - creates a new HTTP event manager
func NewHTTPTransport(config *HTTPTransportConfig, logger *zap.Logger) (*HTTPTransport, error) {

	if config == nil {
		return nil, fmt.Errorf("null configuration found")
	}

	if config.BufferSize <= 0 {
		return nil, fmt.Errorf("invalid buffer size: %d", config.BufferSize)
	}

	sendIntervalDuration, err := time.ParseDuration(config.BatchSendInterval)
	if err != nil {
		return nil, err
	}

	requestTimeoutDuration, err := time.ParseDuration(config.RequestTimeout)
	if err != nil {
		return nil, err
	}

	m := &HTTPTransport{
		batchSendInterval: sendIntervalDuration,
		pointChannel:      make(chan interface{}, config.BufferSize),
		serviceEndpoint:   config.ServiceEndpoint,
		httpClient:        util.CreateHTTPClient(requestTimeoutDuration, true),
		logger:            logger,
	}

	return m, err
}

// Send - send a new event using the parent Point interface
func (t *HTTPTransport) Send(point interface{}) error {

	fmt.Println("send http")
	t.pointChannel <- point

	return nil
}

// ConfigureBackend - configures the backend
func (t *HTTPTransport) ConfigureBackend(backend *Backend) error {

	if backend == nil {
		return fmt.Errorf("no backend was configured")
	}

	t.serviceURL = fmt.Sprintf("http://%s:%d/%s", backend.Host, backend.Port, t.serviceEndpoint)

	go t.transferData()

	return nil
}

// transferData - transfers the data to the backend throught this transport
func (t *HTTPTransport) transferData() {

	lf := []zapcore.Field{
		zap.String("package", "event"),
		zap.String("func", "transferData"),
	}

	t.logger.Info("initializing transfer data loop...", lf...)

	for {
		<-time.After(t.batchSendInterval)

		points := []interface{}{}
		numPoints := 0

		t.logger.Debug("draining points from the buffer...", lf...)

		for point := range t.pointChannel {
			points = append(points, point)
		}

		numPoints = len(points)
		t.logger.Debug(fmt.Sprintf("drained %d points...", numPoints), lf...)

		if numPoints == 0 {
			t.logger.Info("buffer is empty, no data will be send", lf...)
			continue
		}

		t.logger.Info(fmt.Sprintf("sending a batch of %d points...", numPoints), lf...)

		payload, err := json.Marshal(points)
		if err != nil {
			t.logger.Error(fmt.Sprintf("error marshalling point: %s", err.Error()), lf...)
			continue
		}

		req, err := http.NewRequest("PUT", t.serviceURL, bytes.NewBuffer(payload))
		if err != nil {
			t.logger.Error(fmt.Sprintf("error creating request: %s", err.Error()), lf...)
			continue
		}

		req.Header.Set("Content-type", "application/json")

		res, err := t.httpClient.Do(req)
		if err != nil {
			t.logger.Error(fmt.Sprintf("error executing request: %s", err.Error()), lf...)
			continue
		}

		if res.StatusCode != http.StatusNoContent {

			reqResponse, err := ioutil.ReadAll(res.Body)
			if err != nil {
				t.logger.Error(fmt.Sprintf("error reading response body: %s", err.Error()), lf...)
			} else {
				t.logger.Error(fmt.Sprintf("error body: %s", string(reqResponse)), lf...)
			}

		} else {

			t.logger.Info(fmt.Sprintf("batch of %d points was sent!", numPoints), lf...)
		}

		res.Body.Close()
	}
}

// Close - closes the transport
func (t *HTTPTransport) Close() {

	close(t.pointChannel)
}
