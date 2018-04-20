package rubber

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/pborman/uuid"
	"github.com/stretchr/testify/assert"
)

type Document struct {
	Name    string `json:"name"`
	Content string `json:"content"`
	Number  int    `json:"number"`
}

const (
	dataSize = 1000
	esIndex  = "test-index"
)

func randomDocumet(i int) Document {
	return Document{Name: uuid.New(), Content: uuid.New(), Number: i}
}

func generateData() []Document {
	var data []Document
	for i := 0; i < dataSize; i++ {
		data = append(data, randomDocumet(i))
	}
	return data
}

type config struct {
	Datacenter        string `json:"datacenter"`
	ReplicationFactor int    `json:"replicationFactor"`
	Contact           string `json:"contact"`
	TTL               int    `json:"ttl"`
	TUUID             bool   `json:"tuuid"`
}

func genericBackendTest(t *testing.T, backend Backend) {
	client := Create(backend)
	data := generateData()
	doctype := uuid.New()

	t.Run("Generic-CreateKeyspace", func(t *testing.T) {
		var content = config{"dc_gt_a1", 2, "john@example.org", 10, false}
		index := fmt.Sprintf("keyspace_%s", strings.Replace(uuid.New(), "-", "_", -1))
		body := bytes.NewBuffer(nil)
		err := json.NewEncoder(body).Encode(&content)

		assert.NoError(t, err)
		client.CreateIndex(index, body)
	})

	t.Run("Generic-PutDocuments", func(t *testing.T) {
		for index, document := range data {
			status, err := client.Put(esIndex, doctype, fmt.Sprintf("%d", index), document)

			assert.NoError(t, err)
			assert.Equal(t, http.StatusCreated, status)
		}
	})

	time.Sleep(time.Minute)
	t.Run("Generic-GetDocuments", func(t *testing.T) {
		for index, document := range data {
			var doc struct {
				Found  bool     `json:"found"`
				Source Document `json:"_source"`
			}
			client.GetByID(esIndex, doctype, fmt.Sprintf("%d", index), &doc)

			assert.True(t, doc.Found)
			assert.Equal(t, document.Name, doc.Source.Name)
			assert.Equal(t, document.Content, doc.Source.Content)
			assert.Equal(t, document.Number, doc.Source.Number)
		}
	})
}
