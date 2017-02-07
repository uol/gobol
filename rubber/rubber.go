package rubber

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/Sirupsen/logrus"
)

const (
	HEAD   = "HEAD"
	GET    = "GET"
	POST   = "POST"
	PUT    = "PUT"
	DELETE = "DELETE"
)

type Settings struct {
	Nodes     []string
	Preferred string
	Timeout   time.Duration
}

type Elastic struct {
	log     *logrus.Logger
	nodes   []string
	timeout time.Duration

	client *http.Client
}

func New(log *logrus.Logger, settings Settings) *Elastic {

	nodes := []string{settings.Preferred}

	nodes = append(nodes, settings.Nodes...)

	return &Elastic{
		log:     log,
		nodes:   nodes,
		timeout: settings.Timeout * time.Second,

		client: &http.Client{
			Timeout: settings.Timeout * time.Second,
		},
	}
}

func (es *Elastic) request(sulfix, method string, body io.Reader) (int, []byte, error) {

	lf := map[string]interface{}{
		"struct": "ElasticSearch",
		"func":   "request",
		"method": method,
	}

	timeoutTimes := 0

	for _, node := range es.nodes {
		url := fmt.Sprintf("http://%v/%v", node, sulfix)

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

			timeoutTimes++

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

func (es *Elastic) Put(esIndex, esType, id string, obj interface{}) (respCode int, err error) {

	body := &bytes.Buffer{}

	enc := json.NewEncoder(body)

	err = enc.Encode(obj)
	if err != nil {
		return
	}

	sulfix := fmt.Sprintf("%v/%v/%v", esIndex, esType, id)

	respCode, _, err = es.request(sulfix, PUT, body)

	return
}

func (es *Elastic) Post(esIndex, esType, id string, obj interface{}) (respCode int, err error) {

	body := &bytes.Buffer{}

	enc := json.NewEncoder(body)

	err = enc.Encode(obj)
	if err != nil {
		return
	}

	sulfix := fmt.Sprintf("%v/%v/%v", esIndex, esType, id)
	respCode, _, err = es.request(sulfix, POST, body)
	return
}

func (es *Elastic) GetById(esIndex, esType, id string, Response interface{}) (respCode int, err error) {

	sulfix := fmt.Sprintf("%v/%v/%v", esIndex, esType, id)

	respCode, req, err := es.request(sulfix, GET, nil)
	if err != nil {
		return
	}

	if respCode == http.StatusOK {
		err = json.Unmarshal(req, &Response)
	}

	return
}

func (es *Elastic) Query(esIndex, esType string, query, response interface{}) (respCode int, err error) {

	body := &bytes.Buffer{}

	if query != nil {
		enc := json.NewEncoder(body)
		err = enc.Encode(query)
	}
	if err != nil {
		return
	}

	sulfix := fmt.Sprintf("%v/%v/_search", esIndex, esType)

	respCode, req, err := es.request(sulfix, POST, body)
	if err != nil {
		return
	}

	if respCode == http.StatusOK {
		err = json.Unmarshal(req, &response)
	}

	return
}

func (es *Elastic) GetHead(esIndex, esType, id string) (respCode int, err error) {

	sulfix := fmt.Sprintf("%v/%v/%v", esIndex, esType, id)

	respCode, _, err = es.request(sulfix, HEAD, nil)

	return
}

func (es *Elastic) PostBulk(body io.Reader) (respCode int, err error) {

	respCode, _, err = es.request("_bulk", POST, body)

	return
}

func (es *Elastic) Delete(esIndex, esType, id string) (respCode int, err error) {

	sulfix := fmt.Sprintf("%v/%v/%v", esIndex, esType, id)

	respCode, _, err = es.request(sulfix, DELETE, nil)

	return
}

func (es *Elastic) CreateIndex(esIndex string, body io.Reader) (respCode int, err error) {

	respCode, _, err = es.request(esIndex, POST, body)

	return
}

func (es *Elastic) DeleteIndex(esIndex string) (respCode int, err error) {

	respCode, _, err = es.request(esIndex, DELETE, nil)

	return
}
