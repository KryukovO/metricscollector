package storage

import (
	"context"
	"testing"

	"github.com/KryukovO/metricscollector/internal/metric"
	"github.com/KryukovO/metricscollector/internal/storage/repository/memstorage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestStorageRepo(clear bool) (repo StorageRepo, stor []metric.Metrics, err error) {
	var (
		counterVal int64 = 100
		gaugeVal         = 12345.67
	)

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

	repo, err = memstorage.NewMemStorage("", false, 0)
	if err != nil {
		return nil, nil, err
	}
	if !clear {
		err = repo.Update(context.Background(), &stor[0])
		if err != nil {
			return nil, nil, err
		}
		err = repo.Update(context.Background(), &stor[1])
		if err != nil {
			return nil, nil, err
		}
	}

	return
}

func TestGetAll(t *testing.T) {
	repo, stor, err := newTestStorageRepo(false)
	require.NoError(t, err)
	s := storage{repo: repo}

	v, err := s.GetAll(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, stor, v)
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
				mtype: "gauge",
				mname: "RandomValue",
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
				mtype: "counter",
				mname: "PollCount",
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
				mtype: "gauge",
				mname: "Alloc",
			},
			want: want{
				expected: nil,
				wantErr:  false,
			},
		},
		{
			name: "Metric with name exists, but type incorrect",
			args: args{
				mtype: "gauge",
				mname: "PollCount",
			},
			want: want{
				expected: nil,
				wantErr:  false,
			},
		},
		{
			name: "Invalid metric type",
			args: args{
				mtype: "metric",
				mname: "PollCount",
			},
			want: want{
				expected: nil,
				wantErr:  true,
			},
		},
	}

	repo, _, err := newTestStorageRepo(false)
	require.NoError(t, err)
	s := NewStorage(repo)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			v, err := s.GetValue(context.Background(), test.args.mtype, test.args.mname)
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
		repo, _, err := newTestStorageRepo(true)
		require.NoError(t, err)
		s := NewStorage(repo)

		t.Run(test.name, func(t *testing.T) {
			err := s.Update(context.Background(), &test.arg)

			if test.wantErr {
				assert.Error(t, err)
				stor, err := repo.GetAll(context.Background())
				require.NoError(t, err)
				assert.Equal(t, true, len(stor) == 0, "The update returned an error, but the value was saved")
			} else {
				assert.NoError(t, err)
				stor, err := repo.GetAll(context.Background())
				require.NoError(t, err)
				require.Equal(t, true, len(stor) == 1, "The update was successful, but the value was not saved")
				v := stor[0]
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