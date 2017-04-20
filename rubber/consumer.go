package rubber

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"path"
	"time"

	"github.com/Sirupsen/logrus"
)

type esResponse struct {
	status  int
	content []byte
	err     error
}

type esRequest struct {
	index  string
	method string
	path   string
	body   io.Reader

	answ chan esResponse
}

type consumer struct {
	server string
	index  string
	client *http.Client
	logger *logrus.Logger

	input    chan *esRequest
	shutdown chan bool
}

func (c *consumer) loop() error {
	for {
		select {
		case request := <-c.input:
			c.logger.WithFields(logrus.Fields{
				"function":   "loop",
				"sctructure": "consumer",
				"index":      c.index,
				"rindex":     request.index,
			}).Debugf("Request executed")
			status, content, err := c.Request(c.server, request.method,
				path.Join("/", request.index, request.path), request.body)
			request.answ <- esResponse{
				status:  status,
				content: content,
				err:     err,
			}
		case <-c.shutdown:
			c.shutdown <- true
			return nil
		}
	}
}

func (c *consumer) close() error {
	c.shutdown <- true
	<-c.shutdown
	close(c.shutdown)
	return nil
}

func (c *consumer) Request(server, method, urlPath string, body io.Reader) (int, []byte, error) {
	lf := map[string]interface{}{
		"struct": "weightedBackend",
		"func":   "request",
		"method": method,
	}

	url := fmt.Sprintf("http://%s%s", server, path.Join("/", urlPath))
	lf["url"] = url
	lf["httpCode"] = 0

	request, err := http.NewRequest(method, url, body)
	if err != nil {
		return 0, nil, err
	}
	start := time.Now()
	resp, err := c.client.Do(request)
	end := time.Now()

	lf["elapsed"] = end.Sub(start)

	if err != nil {
		return 0, nil, err
	}
	defer resp.Body.Close()
	lf["httpCode"] = resp.StatusCode

	content, err := ioutil.ReadAll(resp.Body)
	return resp.StatusCode, content, err
}
