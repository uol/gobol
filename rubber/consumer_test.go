package rubber

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"testing"

	"github.com/Sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

var integration = (os.Getenv("INTEGRATION") == "true")
var master = os.Getenv("MASTER")

func TestMain(m *testing.M) {
	if integration {
		os.Exit(m.Run())
	}
}

func createTestConsumer() *consumer {
	logger := logrus.New()
	logger.Out = ioutil.Discard

	return &consumer{
		server: fmt.Sprintf("%s:9200", master),
		index:  "index",

		client: http.DefaultClient,
		logger: logger,

		input:    make(chan *esRequest),
		shutdown: make(chan bool),
	}
}

func TestConsumer(t *testing.T) {
	c := createTestConsumer()
	status, message, err := c.Request(c.server, "PUT", path.Join(c.index), nil)
	assert.NoError(t, err)
	assert.Equal(t, status, http.StatusOK)
	assert.Equal(t, "{\"acknowledged\":true}", string(message))
}
