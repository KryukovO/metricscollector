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
		counterVal    int64 = 100
		newCounterVal int64 = 500
		gaugeVal            = 12345.67
		newGaugeVal         = 67.12345
		storage             = []metric.Metrics{
			{
				ID:    "ExistsGaugeMetric",
				MType: metric.GaugeMetric,
				Value: &gaugeVal,
			},
			{
				ID:    "ExistsCounterMetric",
				MType: metric.CounterMetric,
				Delta: &counterVal,
			},
		}
	)

	tests := []struct {
		name      string
		arg       metric.Metrics
		newMetric bool
		wantErr   bool
	}{
		{
			name: "Correct gauge update for new metric",
			arg: metric.Metrics{
				ID:    "RandomValue",
				MType: metric.GaugeMetric,
				Value: &gaugeVal,
			},
			newMetric: true,
			wantErr:   false,
		},
		{
			name: "Correct counter update for new metric",
			arg: metric.Metrics{
				ID:    "PollCount",
				MType: metric.CounterMetric,
				Delta: &counterVal,
			},
			newMetric: true,
			wantErr:   false,
		},
		{
			name: "Correct gauge update for existing metric",
			arg: metric.Metrics{
				ID:    "ExistsGaugeMetric",
				MType: metric.GaugeMetric,
				Value: &newGaugeVal,
			},
			newMetric: false,
			wantErr:   false,
		},
		{
			name: "Correct counter update for existing metric",
			arg: metric.Metrics{
				ID:    "ExistsCounterMetric",
				MType: metric.CounterMetric,
				Delta: &newCounterVal,
			},
			newMetric: false,
			wantErr:   false,
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			storageCopy := make([]metric.Metrics, len(storage))
			copy(storageCopy, storage)

			s := &MemStorage{
				storage: storageCopy,
			}

			err := s.Update(context.Background(), &test.arg)

			if test.wantErr {
				assert.Error(t, err)
				assert.Equal(t, true, len(s.storage) == len(storage), "The update returned an error, but the value was saved")
			} else {
				assert.NoError(t, err)

				if test.newMetric {
					require.Equal(t, true, len(s.storage) == len(storage)+1, "The new metric update was successful, but no value was added.")
				} else {
					require.Equal(t, true, len(s.storage) == len(storage), "The update to an existing metric was successful, but the value was added rather than changed.")
				}

				var v metric.Metrics

				for _, mtrc := range s.storage {
					if mtrc.ID == test.arg.ID {
						v = mtrc
						return
					}
				}

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

func TestUpdateMany(t *testing.T) {
	var (
		counterVal    int64 = 100
		newCounterVal int64 = 500
		gaugeVal            = 12345.67
		newGaugeVal         = 67.12345
		storage             = []metric.Metrics{
			{
				ID:    "ExistsGaugeMetric",
				MType: metric.GaugeMetric,
				Value: &gaugeVal,
			},
			{
				ID:    "ExistsCounterMetric",
				MType: metric.CounterMetric,
				Delta: &counterVal,
			},
		}
	)

	tests := []struct {
		name       string
		arg        []metric.Metrics
		newMetrics bool
		wantErr    bool
	}{
		{
			name: "Correct updates for new metrics",
			arg: []metric.Metrics{
				{
					ID:    "RandomValue",
					MType: metric.GaugeMetric,
					Value: &gaugeVal,
				},
				{
					ID:    "PollCount",
					MType: metric.CounterMetric,
					Delta: &counterVal,
				},
			},
			newMetrics: true,
			wantErr:    false,
		},
		{
			name: "Correct updates for existing metrics",
			arg: []metric.Metrics{
				{
					ID:    "ExistsGaugeMetric",
					MType: metric.GaugeMetric,
					Value: &newGaugeVal,
				},
				{
					ID:    "ExistsCounterMetric",
					MType: metric.CounterMetric,
					Delta: &newCounterVal,
				},
			},
			newMetrics: false,
			wantErr:    false,
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			storageCopy := make([]metric.Metrics, len(storage))
			copy(storageCopy, storage)

			s := &MemStorage{
				storage: storageCopy,
			}

			err := s.UpdateMany(context.Background(), test.arg)

			if test.wantErr {
				assert.Error(t, err)
				assert.Equal(t, true, len(s.storage) == len(storage), "The update returned an error, but the value was saved")
			} else {
				assert.NoError(t, err)

				if test.newMetrics {
					require.Equal(t, true, len(s.storage) == len(storage)+len(test.arg), "Updating new metrics was successful, but no value was added.")
				} else {
					require.Equal(t, true, len(s.storage) == len(storage), "Updating existing metrics was successful, but values was added rather than changed.")
				}

				for _, testMtrc := range test.arg {
					var v metric.Metrics

					for _, mtrc := range s.storage {
						if mtrc.ID == testMtrc.ID {
							v = mtrc
							return
						}
					}

					if testMtrc.Delta != nil {
						assert.EqualValues(
							t, *testMtrc.Delta, *v.Delta,
							"Saved value '%v' is not equal to expected '%v'", *v.Delta, *testMtrc.Delta,
						)
					}
					if testMtrc.Value != nil {
						assert.EqualValues(
							t, *testMtrc.Value, *v.Value,
							"Saved value '%v' is not equal to expected '%v'", *v.Value, *testMtrc.Value,
						)
					}
				}
			}
		})
	}
}

func BenchmarkGet(b *testing.B) {
	ctx := context.Background()
	counterVal := int64(100)
	gaugeVal := 12345.67

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

	b.Run("getValue", func(b *testing.B) {
		s.GetValue(ctx, metric.CounterMetric, "PollCount")
	})

	b.Run("getAll", func(b *testing.B) {
		s.GetAll(ctx)
	})
}

func BenchmarkUpdate(b *testing.B) {
	ctx := context.Background()
	counterVal := int64(100)
	gaugeVal := 12345.67
	mtrc := []metric.Metrics{
		{
			ID:    "RandomValue",
			MType: metric.GaugeMetric,
			Value: &gaugeVal,
		},
		{
			ID:    "PollCount",
			MType: metric.CounterMetric,
			Delta: &counterVal,
		},
	}

	b.Run("update", func(b *testing.B) {
		s := &MemStorage{
			storage: make([]metric.Metrics, 0),
		}

		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			s.Update(ctx, &mtrc[0])
		}
	})

	b.Run("updateMany", func(b *testing.B) {
		s := &MemStorage{
			storage: make([]metric.Metrics, 0),
		}

		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			s.UpdateMany(ctx, mtrc)
		}
	})
}
