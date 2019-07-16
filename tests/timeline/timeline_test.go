package timeline_test

import (
	"encoding/json"
	"net/http"
	"reflect"
	"testing"
	"time"

	"go.uber.org/zap"

	"github.com/stretchr/testify/assert"

	"github.com/uol/gobol/tester/httpserver"
	"github.com/uol/gobol/timeline"
)

var defaultTags = map[string]string{
	"host": "unit-test-host",
	"ttl":  "1",
}

// createTimeseriesBackend - creates a new test server simulating a timeseries backend
func createTimeseriesBackend() *httpserver.HTTPServer {

	headers := http.Header{}
	headers.Add("Content-type", "application/json")

	responses := httpserver.ResponseData{
		RequestData: httpserver.RequestData{
			URI:     "/api/put",
			Method:  "PUT",
			Headers: headers,
		},
		Status: 201,
	}

	return httpserver.CreateNewTestHTTPServer([]httpserver.ResponseData{responses})
}

// createTimelineManager - creates a new timeline manager
func createTimelineManager() *timeline.Manager {

	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}

	transportConf := timeline.HTTPTransportConfig{
		ServiceEndpoint:   "/api/put",
		RequestTimeout:    "1s",
		BatchSendInterval: "1s",
		BufferSize:        5,
	}

	transport, err := timeline.NewHTTPTransport(&transportConf, logger)
	if err != nil {
		panic(err)
	}

	backend := timeline.Backend{
		Host: httpserver.TestServerHost,
		Port: httpserver.TestServerPort,
	}

	manager, err := timeline.NewManager(transport, &backend, defaultTags)
	if err != nil {
		panic(err)
	}

	return manager
}

// newNumberPoint - creates a new number point
func newNumberPoint(value float64) *timeline.NumberPoint {

	return &timeline.NumberPoint{
		Point: timeline.Point{
			Metric:    "number-metric",
			Timestamp: time.Now().Unix(),
			Tags: map[string]string{
				"type":      "number",
				"customTag": "number-test",
			},
		},
		Value: value,
	}
}

// newTextPoint - creates a new text point
func newTextPoint(text string) *timeline.TextPoint {

	return &timeline.TextPoint{
		Point: timeline.Point{
			Metric:    "text-metric",
			Timestamp: time.Now().Unix(),
			Tags: map[string]string{
				"type":      "text",
				"customTag": "text-test",
			},
		},
		Text: text,
	}
}

// testRequestData - tests the request data
func testRequestData(t *testing.T, requestData *httpserver.RequestData, expected interface{}) bool {

	result := true

	result = result && assert.NotNil(t, requestData, "request data cannot be null")
	result = result && assert.Equal(t, "/api/put", requestData.URI, "expected /api/put as endpoint")
	result = result && assert.Equal(t, "PUT", requestData.Method, "expected PUT as method")
	result = result && assert.Equal(t, "application/json", requestData.Headers.Get("Content-type"), "expected aplication/json as content-type header")

	if result {

		if expectedCastedPoint, ok := castToPoint(expected); ok {

			var number *timeline.NumberPoint
			err := json.Unmarshal([]byte(requestData.Body), number)
			if err == nil {
				unmarshalledPoint, _ := castToPoint(number)
				expectedTypedCastedPoint := castToNumberPoint(t, expectedCastedPoint)

				result = result && assert.Equal(t, expectedTypedCastedPoint.Value, number.Value)

				return result && testPoint(t, expectedCastedPoint, unmarshalledPoint)
			}

			var text *timeline.TextPoint
			err = json.Unmarshal([]byte(requestData.Body), text)
			if err == nil {
				unmarshalledPoint, _ := castToPoint(text)
				expectedTypedCastedPoint := castToTextPoint(t, expectedCastedPoint)

				result = result && assert.Equal(t, expectedTypedCastedPoint.Text, text.Text)

				return result && testPoint(t, expectedCastedPoint, unmarshalledPoint)
			}

		} else {

			result = result && assert.Fail(t, "error casting points")
		}
	}

	return result
}

// castToPoint - cast the instance to Point type
func castToPoint(point interface{}) (*timeline.Point, bool) {

	c, ok := point.(*timeline.Point)

	return c, ok
}

// castToNumberPoint - cast the instance to NumberPoint type
func castToNumberPoint(t *testing.T, point interface{}) *timeline.NumberPoint {

	c, ok := point.(*timeline.NumberPoint)
	assert.True(t, ok, "expected a number type point")

	return c
}

// castToTextPoint - cast the instance to TextPoint type
func castToTextPoint(t *testing.T, point interface{}) *timeline.TextPoint {

	c, ok := point.(*timeline.TextPoint)
	assert.True(t, ok, "expected a text type point")

	return c
}

// testPoint - compares two points
func testPoint(t *testing.T, expected *timeline.Point, actualInterface interface{}) bool {

	result := true

	actual := actualInterface.(*timeline.Point)

	result = result && assert.NotNil(t, expected, "point cannot be null")
	result = result && assert.NotNil(t, actual, "point cannot be null")
	result = result && assert.Equal(t, expected.Metric, actual.Metric, "point's metric differs")
	result = result && assert.Equal(t, expected.Timestamp, actual.Timestamp, "point's timestamp differs")
	result = result && assert.True(t, reflect.DeepEqual(expected.Tags, actual.Tags), "point's tags differs")

	return result
}

// TestSendNumber - tests when the lib fires a event
func TestSendNumber(t *testing.T) {

	s := createTimeseriesBackend()
	defer s.Close()

	m := createTimelineManager()
	defer m.Shutdown()

	number := newNumberPoint(1)

	err := m.Send(number)
	assert.NoError(t, err, "no error expected when sending number")

	requestData := httpserver.WaitForHTTPServerRequest(s)
	testRequestData(t, requestData, number)
}
