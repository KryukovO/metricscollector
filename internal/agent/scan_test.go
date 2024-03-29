package agent

import (
	"context"
	"math/rand"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestScanMetrics(t *testing.T) {
	keys := []string{
		"Alloc", "BuckHashSys", "Frees", "GCCPUFraction", "GCSys",
		"HeapAlloc", "HeapIdle", "HeapInuse", "HeapObjects", "HeapReleased", "HeapSys", "LastGC", "Lookups",
		"MCacheInuse", "MCacheSys", "MSpanInuse", "MSpanSys", "Mallocs", "NextGC", "NumForcedGC", "NumGC",
		"OtherSys", "PauseTotalNs", "StackInuse", "StackSys", "Sys", "TotalAlloc", "RandomValue",
		"TotalMemory", "FreeMemory",
	}

	metricCh := ScanMetrics(context.Background())

	result := make([]string, 0, len(keys))

	for data := range metricCh {
		assert.NoError(t, data.err)

		result = append(result, data.mtrc.ID)
	}

	assert.Subset(t, result, keys)
	// NumCPU используется для проверки того, что в результате находится N метрик CPUutilization
	assert.Len(t, result, len(keys)+runtime.NumCPU())
}

func TestScanRuntimeMetrics(t *testing.T) {
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
			metricCh := scanRuntimeMetrics(context.Background(), test.rnd)

			keys := make([]string, 0, len(test.keys))
			for data := range metricCh {
				assert.NoError(t, data.err)

				keys = append(keys, data.mtrc.ID)
			}

			assert.ElementsMatch(t, keys, test.keys)
		})
	}
}

func TestScanPSUtilMetrics(t *testing.T) {
	tests := []struct {
		name string
		keys []string
	}{
		{
			name: "Correct test",
			keys: []string{"TotalMemory", "FreeMemory"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			metricCh := scanPSUtilMetrics(context.Background())

			keys := make([]string, 0, len(test.keys))
			for data := range metricCh {
				assert.NoError(t, data.err)

				keys = append(keys, data.mtrc.ID)
			}

			assert.Subset(t, keys, test.keys)
			// NumCPU используется для проверки того, что в результате находится N метрик CPUutilization
			assert.Len(t, keys, len(test.keys)+runtime.NumCPU())
		})
	}
}

func BenchmarkScan(b *testing.B) {
	ctx := context.Background()
	rnd := rand.New(rand.NewSource(100))

	b.Run("full", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			ScanMetrics(ctx)
		}
	})

	b.Run("runtime", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			scanRuntimeMetrics(ctx, rnd)
		}
	})

	b.Run("psutil", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			ScanMetrics(ctx)
		}
	})
}
