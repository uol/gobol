package rubber

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"path"
)

const (
	// HEAD is an http method constant
	HEAD = "HEAD"
	// GET is an http method constant
	GET = "GET"
	// POST is an http method constant
	POST = "POST"
	// PUT is an http method constant
	PUT = "PUT"
	// DELETE is an http method constant
	DELETE = "DELETE"
)

// Backend defines a way to connect to elasticsearch
type Backend interface {
	Request(index, method, path string, body io.Reader) (int, []byte, error)
}

// Elastic is a wrapper that creates helper functions around the backend
type Elastic struct {
	Backend
}

// Put makes a PUT request to the elasticsearch server
func (es *Elastic) Put(index, esType, id string, obj interface{}) (int, error) {
	body := &bytes.Buffer{}
	if err := json.NewEncoder(body).Encode(obj); err != nil {
		return 0, err
	}

	status, _, err := es.Request(index, PUT, path.Join(esType, id), body)
	return status, err
}

// Post makes a POST request to the ElasticSearch server
func (es *Elastic) Post(index, esType, id string, obj interface{}) (int, error) {
	body := &bytes.Buffer{}
	err := json.NewEncoder(body).Encode(obj)
	if err != nil {
		return 0, err
	}

	status, _, err := es.Request(index, POST, path.Join(esType, id), body)
	return status, err
}

// GetByID gets a document with a given ID
func (es *Elastic) GetByID(index, esType, id string, Response interface{}) (int, error) {
	status, req, err := es.Request(index, GET, path.Join(esType, id), nil)
	if err != nil {
		return status, err
	}
	if status == http.StatusOK {
		err = json.Unmarshal(req, &Response)
	}
	return status, err
}

// Query performs a query in elasticsearch
func (es *Elastic) Query(index, esType string, query, response interface{}) (int, error) {
	var err error
	body := &bytes.Buffer{}
	if query != nil {
		err = json.NewEncoder(body).Encode(query)
	}
	if err != nil {
		return 0, err
	}

	status, req, err := es.Request(index, POST, path.Join(esType, "_search"), body)
	if err != nil {
		return status, err
	}
	if status == http.StatusOK {
		err = json.Unmarshal(req, &response)
	}
	return status, err
}

// GetHead performs a HEAD request against the server
func (es *Elastic) GetHead(index, esType, id string) (int, error) {
	status, _, err := es.Request(index, HEAD, path.Join(esType, id), nil)
	return status, err
}

// PostBulk makes a bulk ElasticSearch post request
func (es *Elastic) PostBulk(body io.Reader) (int, error) {
	status, _, err := es.Request("", POST, "_bulk", body)
	return status, err
}

// Delete makes a DELETE request against the cluster
func (es *Elastic) Delete(index, esType, id string) (int, error) {
	status, _, err := es.Request(index, DELETE, path.Join(esType, id), nil)
	return status, err
}

// CreateIndex creates an index in the cluster
func (es *Elastic) CreateIndex(index string, body io.Reader) (int, error) {
	status, _, err := es.Request(NoIndex, PUT, index, body)
	return status, err
}

// DeleteIndex deletes an index in the cluster
func (es *Elastic) DeleteIndex(index string) (int, error) {
	status, _, err := es.Request(index, DELETE, "", nil)
	return status, err
}
