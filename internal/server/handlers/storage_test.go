package handlers

import (
	"context"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/KryukovO/metricscollector/internal/metric"
	"github.com/KryukovO/metricscollector/internal/storage"
	"github.com/KryukovO/metricscollector/internal/storage/repository/memstorage"
	"github.com/labstack/echo"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestStorageRepo(clear bool) (repo storage.StorageRepo, stor []metric.Metrics, err error) {
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

func TestUpdateHandler(t *testing.T) {
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

			repo, _, err := newTestStorageRepo(true)
			require.NoError(t, err)
			s := StorageController{
				storage: storage.NewStorage(repo),
				l:       logrus.StandardLogger(),
			}
			s.updateHandler(c)

			res := rec.Result()
			defer res.Body.Close()

			assert.Equal(t, test.status, res.StatusCode)
		})
	}
}

func TestGetValueHandler(t *testing.T) {
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

			repo, _, err := newTestStorageRepo(false)
			require.NoError(t, err)
			s := StorageController{
				storage: storage.NewStorage(repo),
				l:       logrus.StandardLogger(),
			}
			s.getValueHandler(c)

			res := rec.Result()
			defer res.Body.Close()

			assert.Equal(t, test.want.status, res.StatusCode)
			assert.Equal(t, test.want.contentType, res.Header.Get("Content-Type"))
		})
	}
}

func TestGetAllHandler(t *testing.T) {
	type args struct {
		url    string
		method string
	}
	type want struct {
		status      int
		contentType string
		tableFormat string
		dataFormat  string
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
				tableFormat: "<table><tr><th>Metric name</th><th>Metric type</th><th>Value</th></tr>%s</table>",
				dataFormat:  "<tr><td>%s</td><td>%s</td><td>%v</td></tr>",
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

			repo, _, err := newTestStorageRepo(false)
			require.NoError(t, err)
			s := StorageController{
				storage: storage.NewStorage(repo),
				l:       logrus.StandardLogger(),
			}
			err = s.getAllHandler(c)
			require.NoError(t, err)

			res := rec.Result()
			defer res.Body.Close()

			rowRes, err := io.ReadAll(res.Body)
			require.NoError(t, err)

			var rowWant string
			stor := repo.GetAll(context.Background())
			for _, mtrc := range stor {
				if mtrc.Delta != nil {
					rowWant += fmt.Sprintf(test.want.dataFormat, mtrc.ID, mtrc.MType, *mtrc.Delta)
				} else if mtrc.Value != nil {
					rowWant += fmt.Sprintf(test.want.dataFormat, mtrc.ID, mtrc.MType, *mtrc.Value)
				}
			}
			rowWant = fmt.Sprintf(test.want.tableFormat, rowWant)

			assert.Equal(t, test.want.status, res.StatusCode)
			assert.Equal(t, test.want.contentType, res.Header.Get("Content-Type"))
			assert.Equal(t, rowWant, string(rowRes))
		})
	}
}
