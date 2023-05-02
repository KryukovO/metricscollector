package memstorage

import (
	"testing"

	"github.com/KryukovO/metricscollector/internal/metric"
	"github.com/stretchr/testify/assert"
)

func TestMemStorage_GetAll(t *testing.T) {
	m := map[string]interface{}{
		"RandomValue": float64(12345.67),
		"PollCount":   int64(100),
	}
	s := &MemStorage{
		storage: m,
	}
	v := s.GetAll()
	assert.Equal(t, m, v)
}

func TestMemStorage_GetValue(t *testing.T) {
	m := map[string]interface{}{
		"RandomValue": float64(12345.67),
		"PollCount":   int64(100),
	}

	type args struct {
		mtype string
		mname string
	}
	type want struct {
		value interface{}
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
				value: float64(12345.67),
				ok:    true,
			},
		},
		{
			name: "Existing counter value",
			args: args{
				mtype: "counter",
				mname: "PollCount",
			},
			want: want{
				value: int64(100),
				ok:    true,
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
			s := MemStorage{
				storage: m,
			}
			v, ok := s.GetValue(test.args.mtype, test.args.mname)
			assert.Equal(t, test.want.value, v)
			assert.Equal(t, test.want.ok, ok)
		})
	}

}

func TestMemStorage_Update(t *testing.T) {
	type args struct {
		mtype string
		mname string
		value interface{}
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Correct gauge update #1",
			args: args{
				mtype: metric.GaugeMetric,
				mname: "Mallocs",
				value: float64(100.0001),
			},
			wantErr: false,
		},
		{
			name: "Correct gauge update #2",
			args: args{
				mtype: metric.GaugeMetric,
				mname: "Mallocs",
				value: int64(100),
			},
			wantErr: false,
		},
		{
			name: "Correct counter update #2",
			args: args{
				mtype: metric.CounterMetric,
				mname: "PollCount",
				value: int64(1),
			},
			wantErr: false,
		},
		{
			name: "Incorrect gauge value",
			args: args{
				mtype: metric.GaugeMetric,
				mname: "Mallocs",
				value: "value",
			},
			wantErr: true,
		},
		{
			name: "Incorrect counter value #1",
			args: args{
				mtype: metric.CounterMetric,
				mname: "PollCount",
				value: float64(100.0001),
			},
			wantErr: true,
		},
		{
			name: "Incorrect counter value #2",
			args: args{
				mtype: metric.CounterMetric,
				mname: "PollCount",
				value: "value",
			},
			wantErr: true,
		},
		{
			name: "Incorrect metric type",
			args: args{
				mtype: "metric",
				mname: "PollCount",
				value: int64(1),
			},
			wantErr: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			m := make(map[string]interface{})
			s := &MemStorage{
				storage: m,
			}

			err := s.Update(test.args.mtype, test.args.mname, test.args.value)
			v, ok := m[test.args.mname]

			if test.wantErr {
				assert.Error(t, err)
				assert.Equal(t, false, ok, "The update returned an error, but the value was saved")
			} else {
				assert.NoError(t, err)
				assert.Equal(t, true, ok, "The update was successful, but the value was not saved")
				assert.EqualValues(t, test.args.value, v, "Saved value '%v' is not equal to expected '%v'", test.args.value, v)
			}
		})
	}
}
