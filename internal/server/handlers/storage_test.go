package handlers

import (
	"math"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/KryukovO/metricscollector/internal/models/metric"
	"github.com/KryukovO/metricscollector/internal/storage/memstorage"
	"github.com/labstack/echo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStorageController_updateHandler(t *testing.T) {
	type args struct {
		url    string
		method string
	}
	tests := []struct {
		name   string
		args   args
		status int
	}{
		{
			name: "Correct gauge test #1",
			args: args{
				url:    "/update/gauge/Mallocs/100.0001",
				method: http.MethodPost,
			},
			status: http.StatusOK,
		},
		{
			name: "Correct gauge test #2",
			args: args{
				url:    "/update/gauge/Mallocs/100",
				method: http.MethodPost,
			},
			status: http.StatusOK,
		},
		{
			name: "Correct counter test",
			args: args{
				url:    "/update/counter/PollCount/1",
				method: http.MethodPost,
			},
			status: http.StatusOK,
		},
		{
			name: "Incorrect gauge value",
			args: args{
				url:    "/update/gauge/Mallocs/value",
				method: http.MethodPost,
			},
			status: http.StatusBadRequest,
		},
		{
			name: "Incorrect counter value #1",
			args: args{
				url:    "/update/counter/PollCount/100.0001",
				method: http.MethodPost,
			},
			status: http.StatusBadRequest,
		},
		{
			name: "Incorrect counter value #2",
			args: args{
				url:    "/update/counter/PollCount/value",
				method: http.MethodPost,
			},
			status: http.StatusBadRequest,
		},
		{
			name: "Incorrect metric type",
			args: args{
				url:    "/update/type/PollCount/1",
				method: http.MethodPost,
			},
			status: http.StatusBadRequest,
		},
		{
			name: "Empty metric type",
			args: args{
				url:    "/update//PollCount/1",
				method: http.MethodPost,
			},
			status: http.StatusBadRequest,
		},
		{
			name: "Empty metric name",
			args: args{
				url:    "/update/counter//1",
				method: http.MethodPost,
			},
			status: http.StatusNotFound,
		},
		{
			name: "Empty metric value",
			args: args{
				url:    "/update/counter/PollCount/",
				method: http.MethodPost,
			},
			status: http.StatusBadRequest,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(test.args.method, test.args.url, nil)
			rec := httptest.NewRecorder()

			c := e.NewContext(req, rec)
			c.SetPath(test.args.url)
			c.SetParamNames("mtype", "mname", "value")
			values := strings.Split(strings.Trim(test.args.url, "/"), "/")
			bound := int(math.Min(float64(len(values)), 1))
			c.SetParamValues(values[bound:]...)

			s := StorageController{
				storage: memstorage.New(),
			}
			s.updateHandler(c)

			res := rec.Result()
			defer res.Body.Close()

			assert.Equal(t, test.status, res.StatusCode)
		})
	}
}

func TestStorageController_getValueHandler(t *testing.T) {
	type args struct {
		url    string
		method string
	}
	type want struct {
		status      int
		contentType string
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "Correct test",
			args: args{
				url:    "/value/counter/PollCount",
				method: http.MethodGet,
			},
			want: want{
				status:      http.StatusOK,
				contentType: "text/plain; charset=UTF-8",
			},
		},
		{
			name: "Metric with name does not exists",
			args: args{
				url:    "/value/counter/Count",
				method: http.MethodGet,
			},
			want: want{
				status:      http.StatusNotFound,
				contentType: "",
			},
		},
		{
			name: "Metric type does not exists",
			args: args{
				url:    "/value/count/PollCount",
				method: http.MethodGet,
			},
			want: want{
				status:      http.StatusNotFound,
				contentType: "",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(test.args.method, test.args.url, nil)
			rec := httptest.NewRecorder()

			c := e.NewContext(req, rec)
			c.SetPath(test.args.url)
			c.SetParamNames("mtype", "mname")
			values := strings.Split(strings.Trim(test.args.url, "/"), "/")
			bound := int(math.Min(float64(len(values)), 1))
			c.SetParamValues(values[bound:]...)

			stor := memstorage.New()
			err := stor.Update(metric.CounterMetric, "PollCount", int64(100))
			require.NoError(t, err)
			s := StorageController{
				storage: stor,
			}
			s.getValueHandler(c)

			res := rec.Result()
			defer res.Body.Close()

			assert.Equal(t, test.want.status, res.StatusCode)
			assert.Equal(t, test.want.contentType, res.Header.Get("Content-Type"))
		})
	}
}

func TestStorageController_getAllHandler(t *testing.T) {
	type args struct {
		url    string
		method string
	}
	type want struct {
		status      int
		contentType string
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "Correct test",
			args: args{
				url:    "/",
				method: http.MethodGet,
			},
			want: want{
				status:      http.StatusOK,
				contentType: "text/html; charset=UTF-8",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(test.args.method, test.args.url, nil)
			rec := httptest.NewRecorder()

			c := e.NewContext(req, rec)
			c.SetPath(test.args.url)
			c.SetParamNames("mtype", "mname")
			values := strings.Split(strings.Trim(test.args.url, "/"), "/")
			bound := int(math.Min(float64(len(values)), 1))
			c.SetParamValues(values[bound:]...)

			stor := memstorage.New()
			err := stor.Update(metric.CounterMetric, "PollCount", int64(100))
			require.NoError(t, err)
			s := StorageController{
				storage: stor,
			}
			s.getAllHandler(c)

			res := rec.Result()
			defer res.Body.Close()

			assert.Equal(t, test.want.status, res.StatusCode)
			assert.Equal(t, test.want.contentType, res.Header.Get("Content-Type"))
		})
	}
}
