package agent

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_scanMetrics(t *testing.T) {
	type want struct {
		keys    []string
		wantErr bool
	}
	tests := []struct {
		name string
		arg  map[string]interface{}
		want want
	}{
		{
			name: "Correct test",
			arg:  make(map[string]interface{}),
			want: want{
				keys: []string{"Alloc", "BuckHashSys", "Frees", "GCCPUFraction", "GCSys",
					"HeapAlloc", "HeapIdle", "HeapInuse", "HeapObjects", "HeapReleased", "HeapSys", "LastGC", "Lookups",
					"MCacheInuse", "MCacheSys", "MSpanInuse", "MSpanSys", "Mallocs", "NextGC", "NumForcedGC", "NumGC",
					"OtherSys", "PauseTotalNs", "StackInuse", "StackSys", "Sys", "TotalAlloc", "PollCount", "RandomValue",
				},
				wantErr: false,
			},
		},
		{
			name: "Nil metrics storage",
			want: want{
				keys:    []string{},
				wantErr: true,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := scanMetrics(test.arg)

			keys := make([]string, 0, len(test.want.keys))
			for key := range test.arg {
				keys = append(keys, key)
			}

			assert.ElementsMatch(t, test.want.keys, keys)
			if !test.want.wantErr {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}
