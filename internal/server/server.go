package server

import (
	"log"
	"net/http"

	"github.com/KryukovO/metricscollector/internal/server/handlers"
	"github.com/KryukovO/metricscollector/internal/storage"
)

func Run(s storage.Storage) {
	log.Println("Server is running...")
	http.ListenAndServe(":8080", handlers.NewHandlers(s))
}
