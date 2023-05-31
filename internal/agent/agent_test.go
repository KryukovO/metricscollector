package agent

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestScanMetrics(t *testing.T) {
	type args struct {
		stor map[string]interface{}
		rnd  *rand.Rand
	}
	type want struct {
		keys    []string
		wantErr bool
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "Correct test",
			args: args{
				stor: make(map[string]interface{}),
				rnd:  rand.New(rand.NewSource(1)),
			},
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
			name: "Nil random generator",
			args: args{
				stor: make(map[string]interface{}),
			},
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
			err := scanMetrics(test.args.stor, test.args.rnd)

			keys := make([]string, 0, len(test.want.keys))
			for key := range test.args.stor {
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
