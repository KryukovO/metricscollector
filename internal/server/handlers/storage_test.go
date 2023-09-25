package handlers

import (
	"bytes"
	"context"
	"errors"
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

var (
	ErrRecorderIsNil = errors.New("empty response recorder")
	ErrURLIsEmpty    = errors.New("empty URL")
)

func newTestRepo(clear bool) (*memstorage.MemStorage, error) {
	var (
		retries          = []int{0}
		counterVal int64 = 100
		gaugeVal         = 12345.67
	)

	repo, err := memstorage.NewMemStorage(context.Background(), "", false, 0, retries, nil)
	if err != nil {
		return nil, err
	}

	if !clear {
		err = repo.Update(
			context.Background(),
			&metric.Metrics{
				ID:    "PollCount",
				MType: metric.CounterMetric,
				Delta: &counterVal,
			},
		)
		if err != nil {
			return nil, err
		}

		err = repo.Update(
			context.Background(),
			&metric.Metrics{
				ID:    "RandomValue",
				MType: metric.GaugeMetric,
				Value: &gaugeVal,
			},
		)
		if err != nil {
			return nil, err
		}
	}

	return repo, nil
}

func newEchoContext(
	rec *httptest.ResponseRecorder, method, url string,
	body io.Reader, paramNames []string,
) (echo.Context, error) {
	if rec == nil {
		return nil, ErrRecorderIsNil
	}

	if url == "" {
		return nil, ErrURLIsEmpty
	}

	e := echo.New()
	req := httptest.NewRequest(method, url, body)

	ctx := e.NewContext(req, rec)
	ctx.SetPath(url)

	if len(paramNames) > 0 {
		ctx.SetParamNames(paramNames...)

		values := strings.Split(strings.Trim(url, "/"), "/")
		bound := int(math.Min(float64(len(values)), 1))
		ctx.SetParamValues(values[bound:]...)
	}

	return ctx, nil
}

func TestUpdateHandler(t *testing.T) {
	params := []string{"mtype", "mname", "value"}
	timeout := 10

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
			rec := httptest.NewRecorder()
			ctx, err := newEchoContext(rec, test.args.method, test.args.url, nil, params)
			require.NoError(t, err)

			repo, err := newTestRepo(true)
			require.NoError(t, err)
			s := StorageController{
				storage: storage.NewMetricsStorage(repo, uint(timeout)),
				l:       logrus.StandardLogger(),
			}
			err = s.updateHandler(ctx)
			require.NoError(t, err)

			res := rec.Result()
			defer res.Body.Close()

			assert.Equal(t, test.status, res.StatusCode)
		})
	}
}

func TestUpdateJSONHandler(t *testing.T) {
	url := "/update/"
	timeout := 10

	type want struct {
		status      int
		contentType string
	}

	tests := []struct {
		name string
		body []byte
		want want
	}{
		{
			name: "Correct gauge test #1",
			body: []byte(`{"id":"Mallocs", "type":"gauge", "value":100.0001}`),
			want: want{
				status:      http.StatusOK,
				contentType: "application/json; charset=UTF-8",
			},
		},
		{
			name: "Correct gauge test #2",
			body: []byte(`{"id":"Mallocs", "type":"gauge", "value":100}`),
			want: want{
				status:      http.StatusOK,
				contentType: "application/json; charset=UTF-8",
			},
		},
		{
			name: "Correct counter test",
			body: []byte(`{"id":"PollCount", "type":"counter", "delta":1}`),
			want: want{
				status:      http.StatusOK,
				contentType: "application/json; charset=UTF-8",
			},
		},
		{
			name: "Incorrect gauge value",
			body: []byte(`{"id":"Mallocs", "type":"gauge", "value":"value"}`),
			want: want{
				status:      http.StatusBadRequest,
				contentType: "",
			},
		},
		{
			name: "Incorrect counter value #1",
			body: []byte(`{"id":"PollCount", "type":"counter", "delta":100.0001}`),
			want: want{
				status:      http.StatusBadRequest,
				contentType: "",
			},
		},
		{
			name: "Incorrect counter value #2",
			body: []byte(`{"id":"PollCount", "type":"counter", "delta":"value"}`),
			want: want{
				status:      http.StatusBadRequest,
				contentType: "",
			},
		},
		{
			name: "Incorrect metric type",
			body: []byte(`{"id":"PollCount", "type":"type", "delta":1}`),
			want: want{
				status:      http.StatusBadRequest,
				contentType: "",
			},
		},
		{
			name: "Empty metric type",
			body: []byte(`{"id":"PollCount", "type":"", "delta":1}`),
			want: want{
				status:      http.StatusBadRequest,
				contentType: "",
			},
		},
		{
			name: "Empty metric name",
			body: []byte(`{"id":"", "type":"counter", "delta":1}`),
			want: want{
				status:      http.StatusNotFound,
				contentType: "",
			},
		},
		{
			name: "Empty metric value",
			body: []byte(`{"id":"PollCount", "type":"counter"}`),
			want: want{
				status:      http.StatusBadRequest,
				contentType: "",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			ctx, err := newEchoContext(rec, http.MethodPost, url, bytes.NewReader(test.body), nil)
			require.NoError(t, err)

			repo, err := newTestRepo(true)
			require.NoError(t, err)
			s := StorageController{
				storage: storage.NewMetricsStorage(repo, uint(timeout)),
				l:       logrus.StandardLogger(),
			}
			err = s.updateJSONHandler(ctx)
			require.NoError(t, err)

			res := rec.Result()
			defer res.Body.Close()

			assert.Equal(t, test.want.status, res.StatusCode)
			assert.Equal(t, test.want.contentType, res.Header.Get("Content-Type"))
		})
	}
}

func TestUpdatesHandler(t *testing.T) {
	url := "/updates/"
	timeout := 10

	tests := []struct {
		name   string
		body   []byte
		status int
	}{
		{
			name: "Correct body",
			body: []byte(
				`[{"id":"Mallocs", "type":"gauge", "value":100.0001}, 
				{"id":"PollCount", "type":"counter", "delta":1}]`,
			),
			status: http.StatusOK,
		},
		{
			name:   "Incorrect gauge value",
			body:   []byte(`[{"id":"Mallocs", "type":"gauge", "value":"value"}]`),
			status: http.StatusBadRequest,
		},
		{
			name:   "Incorrect counter value #1",
			body:   []byte(`[{"id":"PollCount", "type":"counter", "delta":100.0001}]`),
			status: http.StatusBadRequest,
		},
		{
			name:   "Incorrect counter value #2",
			body:   []byte(`[{"id":"PollCount", "type":"counter", "delta":"value"}]`),
			status: http.StatusBadRequest,
		},
		{
			name:   "Incorrect metric type",
			body:   []byte(`[{"id":"PollCount", "type":"type", "delta":1}]`),
			status: http.StatusBadRequest,
		},
		{
			name:   "Empty metric type",
			body:   []byte(`[{"id":"PollCount", "type":"", "delta":1}]`),
			status: http.StatusBadRequest,
		},
		{
			name:   "Empty metric name",
			body:   []byte(`[{"id":"", "type":"counter", "delta":1}]`),
			status: http.StatusNotFound,
		},
		{
			name:   "Empty metric value",
			body:   []byte(`[{"id":"PollCount", "type":"counter"}]`),
			status: http.StatusBadRequest,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			ctx, err := newEchoContext(rec, http.MethodPost, url, bytes.NewReader(test.body), nil)
			require.NoError(t, err)

			repo, err := newTestRepo(true)
			require.NoError(t, err)
			s := StorageController{
				storage: storage.NewMetricsStorage(repo, uint(timeout)),
				l:       logrus.StandardLogger(),
			}
			err = s.updatesHandler(ctx)
			require.NoError(t, err)

			res := rec.Result()
			defer res.Body.Close()

			assert.Equal(t, test.status, res.StatusCode)
		})
	}
}

func TestGetValueHandler(t *testing.T) {
	params := []string{"mtype", "mname"}
	timeout := 10

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
			rec := httptest.NewRecorder()
			ctx, err := newEchoContext(rec, test.args.method, test.args.url, nil, params)
			require.NoError(t, err)

			repo, err := newTestRepo(false)
			require.NoError(t, err)
			s := StorageController{
				storage: storage.NewMetricsStorage(repo, uint(timeout)),
				l:       logrus.StandardLogger(),
			}
			err = s.getValueHandler(ctx)
			require.NoError(t, err)

			res := rec.Result()
			defer res.Body.Close()

			assert.Equal(t, test.want.status, res.StatusCode)
			assert.Equal(t, test.want.contentType, res.Header.Get("Content-Type"))
		})
	}
}

func TestGetValueJSONHandler(t *testing.T) {
	url := "/value/"
	timeout := 10

	type want struct {
		status      int
		contentType string
	}

	tests := []struct {
		name string
		body []byte
		want want
	}{
		{
			name: "Correct body",
			body: []byte(`{"id":"PollCount", "type":"counter"}`),
			want: want{
				status:      http.StatusOK,
				contentType: "application/json; charset=UTF-8",
			},
		},
		{
			name: "Metric with name does not exists",
			body: []byte(`{"id":"Count", "type":"counter"}`),
			want: want{
				status:      http.StatusNotFound,
				contentType: "",
			},
		},
		{
			name: "Metric type does not exists",
			body: []byte(`{"id":"PollCount", "type":"count"}`),
			want: want{
				status:      http.StatusNotFound,
				contentType: "",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			ctx, err := newEchoContext(rec, http.MethodPost, url, bytes.NewReader(test.body), nil)
			require.NoError(t, err)

			repo, err := newTestRepo(false)
			require.NoError(t, err)
			s := StorageController{
				storage: storage.NewMetricsStorage(repo, uint(timeout)),
				l:       logrus.StandardLogger(),
			}
			err = s.getValueJSONHandler(ctx)
			require.NoError(t, err)

			res := rec.Result()
			defer res.Body.Close()

			assert.Equal(t, test.want.status, res.StatusCode)
			assert.Equal(t, test.want.contentType, res.Header.Get("Content-Type"))
		})
	}
}

func TestGetAllHandler(t *testing.T) {
	timeout := 10

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
			rec := httptest.NewRecorder()
			ctx, err := newEchoContext(rec, test.args.method, test.args.url, nil, nil)
			require.NoError(t, err)

			repo, err := newTestRepo(false)
			require.NoError(t, err)
			s := StorageController{
				storage: storage.NewMetricsStorage(repo, uint(timeout)),
				l:       logrus.StandardLogger(),
			}
			err = s.getAllHandler(ctx)
			require.NoError(t, err)

			res := rec.Result()
			defer res.Body.Close()

			rowRes, err := io.ReadAll(res.Body)
			require.NoError(t, err)

			var rowWant string
			stor, err := repo.GetAll(context.Background())
			require.NoError(t, err)

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

func BenchmarkUpdateHandler(b *testing.B) {
	timeout := 10
	params := []string{"mtype", "mname", "value"}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		b.StopTimer()

		rec := httptest.NewRecorder()
		ctx, _ := newEchoContext(rec, http.MethodPost, "/update/gauge/Mallocs/100.0001", nil, params)
		repo, _ := newTestRepo(true)

		s := StorageController{
			storage: storage.NewMetricsStorage(repo, uint(timeout)),
			l:       logrus.StandardLogger(),
		}

		b.StartTimer()

		err := s.updateHandler(ctx)
		if err != nil {
			b.Fatal(err)
		}

		b.StopTimer()

		res := rec.Result()
		res.Body.Close()
	}
}

func BenchmarkUpdateJSONHandler(b *testing.B) {
	timeout := 10

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		b.StopTimer()

		body := bytes.NewReader([]byte(`{"id":"Mallocs", "type":"gauge", "value":100.0001}`))
		rec := httptest.NewRecorder()
		ctx, _ := newEchoContext(rec, http.MethodPost, "/update/", body, nil)
		repo, _ := newTestRepo(true)

		s := StorageController{
			storage: storage.NewMetricsStorage(repo, uint(timeout)),
			l:       logrus.StandardLogger(),
		}

		b.StartTimer()

		err := s.updateJSONHandler(ctx)
		if err != nil {
			b.Fatal(err)
		}

		b.StopTimer()

		res := rec.Result()
		res.Body.Close()
	}
}

func BenchmarkUpdatesHandler(b *testing.B) {
	timeout := 10

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		b.StopTimer()

		rec := httptest.NewRecorder()
		body := bytes.NewReader([]byte(
			`[{"id":"Mallocs", "type":"gauge", "value":100.0001}, 
		{"id":"PollCount", "type":"counter", "delta":1}]`,
		))
		ctx, _ := newEchoContext(rec, http.MethodPost, "/updates/", body, nil)
		repo, _ := newTestRepo(true)

		s := StorageController{
			storage: storage.NewMetricsStorage(repo, uint(timeout)),
			l:       logrus.StandardLogger(),
		}

		b.StartTimer()

		err := s.updatesHandler(ctx)
		if err != nil {
			b.Fatal(err)
		}

		b.StopTimer()

		res := rec.Result()
		res.Body.Close()
	}
}

func BenchmarkGetValueHandler(b *testing.B) {
	timeout := 10
	params := []string{"mtype", "mname"}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		b.StopTimer()

		rec := httptest.NewRecorder()
		ctx, _ := newEchoContext(rec, http.MethodGet, "/value/counter/PollCount", nil, params)
		repo, _ := newTestRepo(false)

		s := StorageController{
			storage: storage.NewMetricsStorage(repo, uint(timeout)),
			l:       logrus.StandardLogger(),
		}

		b.StartTimer()

		err := s.getValueHandler(ctx)
		if err != nil {
			b.Fatal(err)
		}

		b.StopTimer()

		res := rec.Result()
		res.Body.Close()
	}
}

func BenchmarkGetValueJSONHandler(b *testing.B) {
	timeout := 10

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		b.StopTimer()

		rec := httptest.NewRecorder()
		body := bytes.NewReader([]byte(`{"id":"PollCount", "type":"counter"}`))
		ctx, _ := newEchoContext(rec, http.MethodGet, "/value/", body, nil)
		repo, _ := newTestRepo(false)

		s := StorageController{
			storage: storage.NewMetricsStorage(repo, uint(timeout)),
			l:       logrus.StandardLogger(),
		}

		b.StartTimer()

		err := s.getValueJSONHandler(ctx)
		if err != nil {
			b.Fatal(err)
		}

		b.StopTimer()

		res := rec.Result()
		res.Body.Close()
	}
}

func BenchmarkGetAllHandler(b *testing.B) {
	timeout := 10

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		b.StopTimer()

		rec := httptest.NewRecorder()
		ctx, _ := newEchoContext(rec, http.MethodGet, "/", nil, nil)
		repo, _ := newTestRepo(false)

		s := StorageController{
			storage: storage.NewMetricsStorage(repo, uint(timeout)),
			l:       logrus.StandardLogger(),
		}

		b.StartTimer()

		err := s.getAllHandler(ctx)
		if err != nil {
			b.Fatal(err)
		}

		b.StopTimer()

		res := rec.Result()
		res.Body.Close()
	}
}
