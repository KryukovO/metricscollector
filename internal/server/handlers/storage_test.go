package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/KryukovO/metricscollector/internal/storage/memstorage"
	"github.com/stretchr/testify/assert"
)

func TestStorageController_updateHandler(t *testing.T) {
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
			name: "Correct gauge test #1",
			args: args{
				url:    "/update/gauge/Mallocs/100.0001",
				method: http.MethodPost,
			},
			want: want{
				status:      http.StatusOK,
				contentType: "text/plain",
			},
		},
		{
			name: "Correct gauge test #2",
			args: args{
				url:    "/update/gauge/Mallocs/100",
				method: http.MethodPost,
			},
			want: want{
				status:      http.StatusOK,
				contentType: "text/plain",
			},
		},
		{
			name: "Correct counter test",
			args: args{
				url:    "/update/counter/PollCount/1",
				method: http.MethodPost,
			},
			want: want{
				status:      http.StatusOK,
				contentType: "text/plain",
			},
		},
		{
			name: "Incorrect gauge value",
			args: args{
				url:    "/update/gauge/Mallocs/value",
				method: http.MethodPost,
			},
			want: want{
				status:      http.StatusBadRequest,
				contentType: "",
			},
		},
		{
			name: "Incorrect counter value #1",
			args: args{
				url:    "/update/counter/PollCount/100.0001",
				method: http.MethodPost,
			},
			want: want{
				status:      http.StatusBadRequest,
				contentType: "",
			},
		},
		{
			name: "Incorrect counter value #2",
			args: args{
				url:    "/update/counter/PollCount/value",
				method: http.MethodPost,
			},
			want: want{
				status:      http.StatusBadRequest,
				contentType: "",
			},
		},
		{
			name: "Incorrect metric type",
			args: args{
				url:    "/update/type/PollCount/1",
				method: http.MethodPost,
			},
			want: want{
				status:      http.StatusBadRequest,
				contentType: "",
			},
		},
		{
			name: "Empty metric type",
			args: args{
				url:    "/update/PollCount/1",
				method: http.MethodPost,
			},
			want: want{
				status:      http.StatusNotFound,
				contentType: "",
			},
		},
		{
			name: "Empty metric name",
			args: args{
				url:    "/update/counter/1",
				method: http.MethodPost,
			},
			want: want{
				status:      http.StatusNotFound,
				contentType: "",
			},
		},
		{
			name: "Empty metric value",
			args: args{
				url:    "/update/counter/PollCount/",
				method: http.MethodPost,
			},
			want: want{
				status:      http.StatusNotFound,
				contentType: "",
			},
		},
		{
			name: "Wrong HTTP method",
			args: args{
				url:    "/update/counter/PollCount/1",
				method: http.MethodGet,
			},
			want: want{
				status:      http.StatusMethodNotAllowed,
				contentType: "",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(test.args.method, test.args.url, nil)
			w := httptest.NewRecorder()

			c := StorageController{
				storage: memstorage.New(),
			}
			c.updateHandler(w, request)

			res := w.Result()
			defer res.Body.Close()

			assert.Equal(t, res.StatusCode, test.want.status)
			assert.Equal(t, test.want.contentType, res.Header.Get("Content-Type"))
		})
	}
}
