package rubber

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"path"
	"time"

	"github.com/Sirupsen/logrus"
)

// Settings define how the single server client works
type Settings struct {
	// Simple settings
	Nodes     []string
	Preferred string

	// Weighted Settings
	Seed string

	// Generic settings
	Timeout time.Duration
	Type    string
}

type singleServerBackend struct {
	log     *logrus.Logger
	nodes   []string
	timeout time.Duration

	client *http.Client
}

func (es *singleServerBackend) Request(index, method, urlPath string, body io.Reader) (int, []byte, error) {
	lf := map[string]interface{}{
		"struct": "singleServerBackend",
		"func":   "request",
		"method": method,
	}

	for _, node := range es.nodes {
		url := fmt.Sprintf("http://%s%s", node, path.Join("/", index, urlPath))

		lf["url"] = url
		lf["httpCode"] = 0

		req, err := http.NewRequest(method, url, body)
		if err != nil {
			return 0, []byte{}, err
		}

		startTime := time.Now()
		resp, err := es.client.Do(req)
		elapsedTime := time.Since(startTime)
		lf["elapsed"] = elapsedTime

		if err != nil {
			es.log.WithFields(lf).Error("trying next node...", err)
			continue
		}
		defer resp.Body.Close()

		lf["httpCode"] = resp.StatusCode
		reqResponse, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return 0, []byte{}, err
		}

		if resp.StatusCode < http.StatusInternalServerError {
			return resp.StatusCode, reqResponse, nil
		}

		es.log.WithFields(lf).Error(reqResponse)
		es.log.WithFields(lf).Error("trying next node...")
	}

	return 0, []byte{}, errors.New("elasticsearch: request failed on all nodes")
}
