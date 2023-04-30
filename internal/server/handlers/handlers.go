package handlers

import (
	"net/http"

	"github.com/KryukovO/metricscollector/internal/storage"
)

func NewHandlers(s storage.Storage) *http.ServeMux {
	mux := http.NewServeMux()

	newStorageHandlers(mux, s)

	return mux
}
