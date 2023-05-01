package server

import (
	"log"
	"net/http"

	"github.com/KryukovO/metricscollector/internal/server/handlers"
	"github.com/KryukovO/metricscollector/internal/storage"
)

func Run(s storage.Storage) error {
	log.Println("Server is running...")

	mux, err := handlers.NewHandlers(s)
	if err != nil {
		return err
	}

	http.ListenAndServe(":8080", mux)
	return err
}
