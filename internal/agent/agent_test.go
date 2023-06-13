package agent

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestScanMetrics(t *testing.T) {
	tests := []struct {
		name string
		rnd  *rand.Rand
		keys []string
	}{
		{
			name: "Correct test",
			rnd:  rand.New(rand.NewSource(1)),
			keys: []string{
				"Alloc", "BuckHashSys", "Frees", "GCCPUFraction", "GCSys",
				"HeapAlloc", "HeapIdle", "HeapInuse", "HeapObjects", "HeapReleased", "HeapSys", "LastGC", "Lookups",
				"MCacheInuse", "MCacheSys", "MSpanInuse", "MSpanSys", "Mallocs", "NextGC", "NumForcedGC", "NumGC",
				"OtherSys", "PauseTotalNs", "StackInuse", "StackSys", "Sys", "TotalAlloc", "RandomValue",
			},
		},
		{
			name: "Nil random generator",
			keys: []string{
				"Alloc", "BuckHashSys", "Frees", "GCCPUFraction", "GCSys",
				"HeapAlloc", "HeapIdle", "HeapInuse", "HeapObjects", "HeapReleased", "HeapSys", "LastGC", "Lookups",
				"MCacheInuse", "MCacheSys", "MSpanInuse", "MSpanSys", "Mallocs", "NextGC", "NumForcedGC", "NumGC",
				"OtherSys", "PauseTotalNs", "StackInuse", "StackSys", "Sys", "TotalAlloc", "RandomValue",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			stor, err := scanMetrics(test.rnd)
			assert.NoError(t, err)

			keys := make([]string, 0, len(test.keys))
			for _, mtrc := range stor {
				keys = append(keys, mtrc.ID)
			}

			assert.ElementsMatch(t, test.keys, keys)
		})
	}
}
