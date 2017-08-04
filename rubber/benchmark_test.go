package rubber

import (
	"fmt"
	"sync"
	"testing"

	"github.com/pborman/uuid"
)

const (
	BenchSize      = 100
	BenchParallell = 50
)

func createBenchmarkData() []Document {
	var data []Document
	for i := 0; i < BenchSize; i++ {
		data = append(data, randomDocumet(i))
	}
	return data
}

func benchmarkBackend(b *testing.B, backend Backend) {
	var (
		index   string = uuid.New()
		doctype string = uuid.New()

		data [][BenchParallell][]Document
	)
	es := Create(backend)
	es.CreateIndex(index, nil)

	for i := 0; i < b.N; i++ {
		var instance [BenchParallell][]Document
		for j := 0; j < BenchParallell; j++ {
			instance[j] = createBenchmarkData()
		}

		data = append(data, instance)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var wg sync.WaitGroup
		wg.Add(BenchParallell)
		for j := 0; j < BenchParallell; j++ {
			go func(data []Document, thread int) {
				for _, doc := range data {
					es.Put(index, doctype, fmt.Sprintf("%d-%d", thread, doc.Number), doc)
				}
				wg.Done()
			}(data[i][j], j)
		}
		wg.Wait()
	}
}

func BenchmarkSimpleBackend(b *testing.B) {
	benchmarkBackend(b, testSimpleBackend())
}
func BenchmarkWeigthedBackend(b *testing.B) {
	benchmarkBackend(b, testWeightedBackend(DefaultUpdate))
}
