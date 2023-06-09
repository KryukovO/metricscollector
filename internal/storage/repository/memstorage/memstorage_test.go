package memstorage

import (
	"context"
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

	v, err := s.GetAll(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, s.storage, v)
}

func TestGetValue(t *testing.T) {
	var (
		counterVal int64 = 100
		gaugeVal         = 12345.67
	)

	type args struct {
		mType string
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
				mType: "gauge",
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
				mType: "counter",
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
				mType: "gauge",
				mname: "Alloc",
			},
			want: want{
				value: &metric.Metrics{},
				ok:    false,
			},
		},
		{
			name: "Metric with name exists, but type incorrect",
			args: args{
				mType: "gauge",
				mname: "PollCount",
			},
			want: want{
				value: &metric.Metrics{},
				ok:    false,
			},
		},
		{
			name: "Incorrect metric type",
			args: args{
				mType: "metric",
				mname: "PollCount",
			},
			want: want{
				value: &metric.Metrics{},
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

			v, err := s.GetValue(context.Background(), test.args.mType, test.args.mname)
			assert.NoError(t, err)
			assert.Equal(t, test.want.value, v)
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
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s := &MemStorage{
				storage: make([]metric.Metrics, 0),
			}

			err := s.Update(context.Background(), &test.arg)

			if test.wantErr {
				assert.Error(t, err)
				assert.Equal(t, true, len(s.storage) == 0, "The update returned an error, but the value was saved")
			} else {
				assert.NoError(t, err)
				require.Equal(t, true, len(s.storage) == 1, "The update was successful, but the value was not saved")
				v := s.storage[0]
				if test.arg.Delta != nil {
					assert.EqualValues(
						t, *test.arg.Delta, *v.Delta,
						"Saved value '%v' is not equal to expected '%v'", *v.Delta, *test.arg.Delta,
					)
				}
				if test.arg.Value != nil {
					assert.EqualValues(
						t, *test.arg.Value, *v.Value,
						"Saved value '%v' is not equal to expected '%v'", *v.Value, *test.arg.Value,
					)
				}
			}
		})
	}
}
