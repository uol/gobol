package rubber

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"path"
	"time"

	"go.uber.org/zap"
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
	log     *zap.Logger
	nodes   []string
	timeout time.Duration

	client *http.Client
}

func (es *singleServerBackend) Request(index, method, urlPath string, body io.Reader) (int, []byte, error) {

	ctxt := es.log.With(
		zap.String("struct", "singleServerBackend"),
		zap.String("func", "request"),
		zap.String("method", method),
	)

	for _, node := range es.nodes {
		url := fmt.Sprintf("http://%s%s", node, path.Join("/", index, urlPath))

		ctxt = ctxt.With(zap.String("url", url))

		req, err := http.NewRequest(method, url, body)
		if err != nil {
			return 0, []byte{}, err
		}

		startTime := time.Now()
		resp, err := es.client.Do(req)
		elapsedTime := time.Since(startTime)

		ctxt = ctxt.With(zap.String("elapsed", elapsedTime.String()))

		if err != nil {
			ctxt.Error("trying next node...", zap.Error(err))
			continue
		}
		defer resp.Body.Close()

		ctxt = ctxt.With(
			zap.String("elapsed", elapsedTime.String()),
			zap.Int("httpCode", resp.StatusCode),
		)

		reqResponse, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return 0, []byte{}, err
		}

		if resp.StatusCode < http.StatusInternalServerError {
			return resp.StatusCode, reqResponse, nil
		}

		ctxt.Error(string(reqResponse))
		ctxt.Error("trying next node...")
	}

	return 0, []byte{}, errors.New("elasticsearch: request failed on all nodes")
}
