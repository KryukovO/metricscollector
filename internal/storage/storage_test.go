package storage

import (
	"context"
	"testing"
	"time"

	"github.com/KryukovO/metricscollector/internal/metric"
	"github.com/KryukovO/metricscollector/internal/storage/repository/memstorage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestRepo(ctx context.Context, clear bool) (*memstorage.MemStorage, []metric.Metrics, error) {
	var (
		retries          = []int{0}
		counterVal int64 = 100
		gaugeVal         = 12345.67
		stor       []metric.Metrics
	)

	repo, err := memstorage.NewMemStorage(ctx, "", false, 0, retries, nil)
	if err != nil {
		return nil, nil, err
	}

	if !clear {
		stor = []metric.Metrics{
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
		}

		err = repo.Update(ctx, &stor[0])
		if err != nil {
			return nil, nil, err
		}

		err = repo.Update(ctx, &stor[1])
		if err != nil {
			return nil, nil, err
		}
	}

	return repo, stor, nil
}

func TestGetAll(t *testing.T) {
	repo, stor, err := newTestRepo(context.Background(), false)
	require.NoError(t, err)

	s := MetricsStorage{repo: repo}

	v, err := s.GetAll(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, stor, v)
}

func TestGetValue(t *testing.T) {
	var (
		timeout          = 10 * time.Second
		counterVal int64 = 100
		gaugeVal         = 12345.67
	)

	type args struct {
		mType string
		mName string
	}

	type want struct {
		expected *metric.Metrics
		wantErr  bool
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
				mName: "RandomValue",
			},
			want: want{
				expected: &metric.Metrics{
					ID:    "RandomValue",
					MType: metric.GaugeMetric,
					Value: &gaugeVal,
				},
				wantErr: false,
			},
		},
		{
			name: "Existing counter value",
			args: args{
				mType: "counter",
				mName: "PollCount",
			},
			want: want{
				expected: &metric.Metrics{
					ID:    "PollCount",
					MType: metric.CounterMetric,
					Delta: &counterVal,
				},
				wantErr: false,
			},
		},
		{
			name: "Metric with name does not exists",
			args: args{
				mType: "gauge",
				mName: "Alloc",
			},
			want: want{
				expected: &metric.Metrics{},
				wantErr:  false,
			},
		},
		{
			name: "Metric with name exists, but type incorrect",
			args: args{
				mType: "gauge",
				mName: "PollCount",
			},
			want: want{
				expected: &metric.Metrics{},
				wantErr:  false,
			},
		},
		{
			name: "Invalid metric type",
			args: args{
				mType: "metric",
				mName: "PollCount",
			},
			want: want{
				expected: nil,
				wantErr:  true,
			},
		},
	}

	repo, _, err := newTestRepo(context.Background(), false)
	require.NoError(t, err)

	s := NewMetricsStorage(repo, timeout)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			v, err := s.GetValue(context.Background(), test.args.mType, test.args.mName)
			assert.Equal(t, test.want.expected, v)
			if test.want.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUpdate(t *testing.T) {
	var (
		timeout          = 10 * time.Second
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
		test := test
		repo, _, err := newTestRepo(context.Background(), true)
		require.NoError(t, err)

		s := NewMetricsStorage(repo, timeout)

		t.Run(test.name, func(t *testing.T) {
			err := s.Update(context.Background(), &test.arg)

			if test.wantErr {
				assert.Error(t, err)
				stor, getErr := repo.GetAll(context.Background())
				require.NoError(t, getErr)
				assert.Empty(t, stor, "The update returned an error, but the value was saved")
			} else {
				assert.NoError(t, err)
				stor, getErr := repo.GetAll(context.Background())
				require.NoError(t, getErr)
				require.Len(t, stor, 1, "The update was successful, but the value was not saved")
				v := stor[0]
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
		timeout          = 10 * time.Second
		counterVal int64 = 100
		gaugeVal         = 12345.67
	)

	tests := []struct {
		name    string
		arg     []metric.Metrics
		wantErr bool
	}{
		{
			name: "Correct update",
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
			wantErr: false,
		},
		{
			name: "Incorrect gauge value #1",
			arg: []metric.Metrics{
				{
					ID:    "RandomValue",
					MType: metric.GaugeMetric,
				},
			},
			wantErr: true,
		},
		{
			name: "Incorrect gauge value #2",
			arg: []metric.Metrics{
				{
					ID:    "RandomValue",
					MType: metric.GaugeMetric,
					Delta: &counterVal,
				},
			},
			wantErr: true,
		},
		{
			name: "Incorrect counter value #1",
			arg: []metric.Metrics{
				{
					ID:    "PollCount",
					MType: metric.CounterMetric,
				},
			},
			wantErr: true,
		},
		{
			name: "Incorrect counter value #2",
			arg: []metric.Metrics{
				{
					ID:    "PollCount",
					MType: metric.CounterMetric,
					Value: &gaugeVal,
				},
			},
			wantErr: true,
		},
		{
			name: "Incorrect metric type #1",
			arg: []metric.Metrics{
				{
					ID:    "PollCount",
					MType: "metric",
					Delta: &counterVal,
				},
			},
			wantErr: true,
		},
		{
			name: "Incorrect metric type #2",
			arg: []metric.Metrics{
				{
					ID:    "RandomValue",
					MType: "metric",
					Value: &gaugeVal,
				},
			},
			wantErr: true,
		},
	}

	for _, test := range tests {
		repo, _, err := newTestRepo(context.Background(), true)
		require.NoError(t, err)

		s := NewMetricsStorage(repo, timeout)

		t.Run(test.name, func(t *testing.T) {
			err := s.UpdateMany(context.Background(), test.arg)

			if test.wantErr {
				assert.Error(t, err)
				stor, getErr := repo.GetAll(context.Background())
				require.NoError(t, getErr)
				assert.Empty(t, stor, "The update returned an error, but the value was saved")
			} else {
				assert.NoError(t, err)
				stor, getErr := repo.GetAll(context.Background())
				require.NoError(t, getErr)
				require.Len(t, stor, len(test.arg), "The update was successful, but the value was not saved")
				assert.EqualValues(t, test.arg, stor, "Saved value '%+v' is not equal to expected '%+v'", stor, test.arg)
			}
		})
	}
}

func BenchmarkGet(b *testing.B) {
	ctx := context.Background()
	repo, mtrc, _ := newTestRepo(context.Background(), false)

	s := MetricsStorage{repo: repo}

	b.Run("getValue", func(b *testing.B) {
		_, err := s.GetValue(ctx, mtrc[0].MType, mtrc[0].ID)
		if err != nil {
			b.Fatal(err)
		}
	})

	b.Run("getAll", func(b *testing.B) {
		_, err := s.GetAll(ctx)
		if err != nil {
			b.Fatal(err)
		}
	})
}

func BenchmarkUpdate(b *testing.B) {
	ctx := context.Background()
	timeout := 10 * time.Second
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
		repo, _, _ := newTestRepo(ctx, true)
		s := NewMetricsStorage(repo, timeout)

		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			err := s.Update(ctx, &mtrc[0])
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("updateMany", func(b *testing.B) {
		repo, _, _ := newTestRepo(ctx, true)
		s := NewMetricsStorage(repo, timeout)

		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			err := s.UpdateMany(ctx, mtrc)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}
