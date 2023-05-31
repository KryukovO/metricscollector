package memstorage

import (
	"testing"

	"github.com/KryukovO/metricscollector/internal/metric"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetAll(t *testing.T) {
	var (
		counterVal int64 = 100
		gaugeVal         = 12345.67
	)

	s := &MemStorage{
		storage: []metric.Metrics{
			{
				ID:    "PollCount",
				MType: metric.CounterMetric,
				Delta: &counterVal,
			},
			{
				ID:    "RandomValue",
				MType: metric.GaugeMetric,
				Value: &gaugeVal,
			},
		},
	}

	v := s.GetAll()
	assert.Equal(t, s.storage, v)
}

func TestGetValue(t *testing.T) {
	var (
		counterVal int64 = 100
		gaugeVal         = 12345.67
	)

	type args struct {
		mtype string
		mname string
	}
	type want struct {
		value *metric.Metrics
		ok    bool
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "Existing gauge value",
			args: args{
				mtype: "gauge",
				mname: "RandomValue",
			},
			want: want{
				value: &metric.Metrics{
					ID:    "RandomValue",
					MType: metric.GaugeMetric,
					Value: &gaugeVal,
				},
				ok: true,
			},
		},
		{
			name: "Existing counter value",
			args: args{
				mtype: "counter",
				mname: "PollCount",
			},
			want: want{
				value: &metric.Metrics{
					ID:    "PollCount",
					MType: metric.CounterMetric,
					Delta: &counterVal,
				},
				ok: true,
			},
		},
		{
			name: "Metric with name does not exists",
			args: args{
				mtype: "gauge",
				mname: "Alloc",
			},
			want: want{
				value: nil,
				ok:    false,
			},
		},
		{
			name: "Metric with name exists, but type incorrect",
			args: args{
				mtype: "gauge",
				mname: "PollCount",
			},
			want: want{
				value: nil,
				ok:    false,
			},
		},
		{
			name: "Incorrect metric type",
			args: args{
				mtype: "metric",
				mname: "PollCount",
			},
			want: want{
				value: nil,
				ok:    false,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s := &MemStorage{
				storage: []metric.Metrics{
					{
						ID:    "PollCount",
						MType: metric.CounterMetric,
						Delta: &counterVal,
					},
					{
						ID:    "RandomValue",
						MType: metric.GaugeMetric,
						Value: &gaugeVal,
					},
				},
			}

			v, ok := s.GetValue(test.args.mtype, test.args.mname)
			assert.Equal(t, test.want.value, v)
			assert.Equal(t, test.want.ok, ok)
		})
	}

}

func TestUpdate(t *testing.T) {
	var (
		counterVal int64 = 100
		gaugeVal         = 12345.67
	)

	tests := []struct {
		name    string
		arg     metric.Metrics
		wantErr bool
	}{
		{
			name: "Correct gauge update",
			arg: metric.Metrics{
				ID:    "RandomValue",
				MType: metric.GaugeMetric,
				Value: &gaugeVal,
			},
			wantErr: false,
		},
		{
			name: "Correct counter update",
			arg: metric.Metrics{
				ID:    "PollCount",
				MType: metric.CounterMetric,
				Delta: &counterVal,
			},
			wantErr: false,
		},
		{
			name: "Incorrect gauge value #1",
			arg: metric.Metrics{
				ID:    "RandomValue",
				MType: metric.GaugeMetric,
			},
			wantErr: true,
		},
		{
			name: "Incorrect gauge value #2",
			arg: metric.Metrics{
				ID:    "RandomValue",
				MType: metric.GaugeMetric,
				Delta: &counterVal,
			},
			wantErr: true,
		},
		{
			name: "Incorrect counter value #1",
			arg: metric.Metrics{
				ID:    "PollCount",
				MType: metric.CounterMetric,
			},
			wantErr: true,
		},
		{
			name: "Incorrect counter value #2",
			arg: metric.Metrics{
				ID:    "PollCount",
				MType: metric.CounterMetric,
				Value: &gaugeVal,
			},
			wantErr: true,
		},
		{
			name: "Incorrect metric type #1",
			arg: metric.Metrics{
				ID:    "PollCount",
				MType: "metric",
				Delta: &counterVal,
			},
			wantErr: true,
		},
		{
			name: "Incorrect metric type #2",
			arg: metric.Metrics{
				ID:    "RandomValue",
				MType: "metric",
				Value: &gaugeVal,
			},
			wantErr: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s := &MemStorage{
				storage: make([]metric.Metrics, 0),
			}

			err := s.Update(&test.arg)

			if test.wantErr {
				assert.Error(t, err)
				assert.Equal(t, true, len(s.storage) == 0, "The update returned an error, but the value was saved")
			} else {
				assert.NoError(t, err)
				require.Equal(t, true, len(s.storage) == 1, "The update was successful, but the value was not saved")
				v := s.storage[0]
				if test.arg.Delta != nil {
					assert.EqualValues(t, *test.arg.Delta, *v.Delta, "Saved value '%v' is not equal to expected '%v'", *v.Delta, *test.arg.Delta)
				}
				if test.arg.Value != nil {
					assert.EqualValues(t, *test.arg.Value, *v.Value, "Saved value '%v' is not equal to expected '%v'", *v.Value, *test.arg.Value)
				}
			}
		})
	}
}
