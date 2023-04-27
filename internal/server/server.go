package server

import (
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/KryukovO/metricscollector/internal/middleware"
	"github.com/KryukovO/metricscollector/internal/storage"
)

type Server struct {
	storage storage.Storage
}

func New(st storage.Storage) *Server {
	return &Server{
		storage: st,
	}
}

func (s *Server) Run() {
	log.Println("Server is running...")
	http.ListenAndServe(":8080", s.setHandlers())
}

func (s *Server) setHandlers() *http.ServeMux {
	mux := http.NewServeMux()
	mux.Handle("/update/", middleware.LoggingMiddleware(http.HandlerFunc(s.updateHandler)))

	return mux
}

func (s *Server) updateHandler(w http.ResponseWriter, r *http.Request) {
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

	err = s.storage.Update(pathParts[1], pathParts[2], value)
	if err == storage.ErrWrongMetricType {
		log.Println(err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if err != nil {
		log.Printf("something went wrong: %s", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
