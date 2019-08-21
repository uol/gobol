package timeline

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/uol/gobol/util"
	json "github.com/uol/serializer/json"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

/**
* The HTTP JSON transport implementation.
* @author rnojiri
**/

// HTTPTransport - implements the HTTP transport
type HTTPTransport struct {
	Transport
	httpClient           *http.Client
	batchSendInterval    time.Duration
	pointChannel         chan interface{}
	serviceURL           string
	configuration        *HTTPTransportConfig
	jsonSerializer       *json.Serializer
	useCustomJSONMapping bool
	logger               *zap.Logger
}

// HTTPTransportConfig - has all HTTP event manager configurations
type HTTPTransportConfig struct {
	BufferSize             int
	BatchSendInterval      string
	ServiceEndpoint        string
	Method                 string
	RequestTimeout         string
	ExpectedResponseStatus int
	SerializerBufferSize   int
}

const (
	numberPointJSON = "numberPointJSON"
	textPointJSON   = "textPointJSON"
)

// NewHTTPTransport - creates a new HTTP event manager
func NewHTTPTransport(configuration *HTTPTransportConfig, logger *zap.Logger) (*HTTPTransport, error) {

	if configuration == nil {
		return nil, fmt.Errorf("null configuration found")
	}

	if configuration.BufferSize <= 0 {
		return nil, fmt.Errorf("invalid buffer size: %d", configuration.BufferSize)
	}

	sendIntervalDuration, err := time.ParseDuration(configuration.BatchSendInterval)
	if err != nil {
		return nil, err
	}

	requestTimeoutDuration, err := time.ParseDuration(configuration.RequestTimeout)
	if err != nil {
		return nil, err
	}

	s := json.New(configuration.SerializerBufferSize)

	m := &HTTPTransport{
		batchSendInterval:    sendIntervalDuration,
		pointChannel:         make(chan interface{}, configuration.BufferSize),
		configuration:        configuration,
		httpClient:           util.CreateHTTPClient(requestTimeoutDuration, true),
		jsonSerializer:       s,
		useCustomJSONMapping: false,
		logger:               logger,
	}

	return m, err
}

// OverrideNumberPointJSONMapping - overrides the default generic property mappings
func (t *HTTPTransport) OverrideNumberPointJSONMapping(p *NumberPoint, variables ...string) error {

	t.useCustomJSONMapping = true

	return t.jsonSerializer.Add(numberPointJSON, *p, variables...)
}

// OverrideTextPointJSONMapping - overrides the default generic property mappings
func (t *HTTPTransport) OverrideTextPointJSONMapping(p *TextPoint, variables ...string) error {

	t.useCustomJSONMapping = true

	return t.jsonSerializer.Add(textPointJSON, *p, variables...)
}

// PointChannel - send a new point
func (t *HTTPTransport) PointChannel() chan<- interface{} {

	return t.pointChannel
}

// ConfigureBackend - configures the backend
func (t *HTTPTransport) ConfigureBackend(backend *Backend) error {

	if backend == nil {
		return fmt.Errorf("no backend was configured")
	}

	t.serviceURL = fmt.Sprintf("http://%s:%d/%s", backend.Host, backend.Port, t.configuration.ServiceEndpoint)

	lf := []zapcore.Field{
		zap.String("package", "timeline"),
		zap.String("func", "ConfigureBackend"),
	}

	t.logger.Info(fmt.Sprintf("backend was configured to use service: %s", t.serviceURL), lf...)

	return nil
}

// Start - starts the transport
func (t *HTTPTransport) Start() error {

	if !t.useCustomJSONMapping {

		err := t.jsonSerializer.Add(numberPointJSON, NumberPoint{}, "metric", "tags", "timestamp", "value")
		if err != nil {
			return err
		}

		err = t.jsonSerializer.Add(textPointJSON, TextPoint{}, "metric", "tags", "timestamp", "text")
		if err != nil {
			return err
		}
	}

	lf := []zapcore.Field{
		zap.String("package", "timeline"),
		zap.String("func", "Start"),
	}

	t.logger.Info("starting http transport...", lf...)

	go t.transferData()

	return nil
}

// transferData - transfers the data to the backend throught this transport
func (t *HTTPTransport) transferData() {

	lf := []zapcore.Field{
		zap.String("package", "timeline"),
		zap.String("func", "transferData"),
	}

	t.logger.Info("initializing transfer data loop...", lf...)

outterFor:
	for {
		<-time.After(t.batchSendInterval)

		points := []json.Parameters{}
		numPoints := 0

		t.logger.Debug("draining points from the buffer...", lf...)

	innerLoop:
		for {
			select {
			case point, ok := <-t.pointChannel:
				if !ok {
					t.logger.Info("breaking data transfer loop", lf...)
					break outterFor
				}

				if casted, ok := point.(json.Parameters); ok {
					points = append(points, casted)
				} else {
					t.logger.Error("error converting channel item to json.Parameter", lf...)
					continue
				}

			default:
				break innerLoop
			}
		}

		numPoints = len(points)
		t.logger.Debug(fmt.Sprintf("drained %d points...", numPoints), lf...)

		if numPoints == 0 {
			t.logger.Info("buffer is empty, no data will be send", lf...)
			continue
		}

		t.logger.Info(fmt.Sprintf("sending a batch of %d points...", numPoints), lf...)

		payload, err := t.jsonSerializer.SerializeArray(points...)
		if err != nil {
			t.logger.Error(fmt.Sprintf("error marshalling point: %s", err.Error()), lf...)
			continue
		}

		req, err := http.NewRequest(t.configuration.Method, t.serviceURL, bytes.NewBuffer([]byte(payload)))
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

		if res.StatusCode != t.configuration.ExpectedResponseStatus {

			reqResponse, err := ioutil.ReadAll(res.Body)
			if err != nil {
				t.logger.Error(fmt.Sprintf("error reading body: %s", err.Error()), lf...)
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
