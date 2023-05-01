package handlers

import (
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/KryukovO/metricscollector/internal/server/middleware"
	"github.com/KryukovO/metricscollector/internal/storage"
)

type StorageController struct {
	storage storage.Storage
}

func newStorageHandlers(mux *http.ServeMux, s storage.Storage) {
	c := &StorageController{storage: s}
	mux.Handle("/update/", middleware.LoggingMiddleware(http.HandlerFunc(c.updateHandler)))
}

func (c *StorageController) updateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		log.Println("method not allowed")
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) != 4 {
		log.Println("empty metric name")
		w.WriteHeader(http.StatusNotFound)
		return
	}

	var value interface{}
	value, err := strconv.ParseInt(pathParts[3], 10, 64)
	if err != nil {
		value, err = strconv.ParseFloat(pathParts[3], 64)
		if err != nil {
			log.Println(storage.ErrWrongMetricValue.Error())
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}

	err = c.storage.Update(pathParts[1], pathParts[2], value)
	if err == storage.ErrWrongMetricType || err == storage.ErrWrongMetricValue {
		log.Println(err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if err != nil {
		log.Printf("something went wrong: %s", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
}
