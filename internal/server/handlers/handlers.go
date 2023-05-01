package handlers

import (
	"net/http"

	"github.com/KryukovO/metricscollector/internal/storage"
)

func NewHandlers(s storage.Storage) (*http.ServeMux, error) {
	mux := http.NewServeMux()

	if err := newStorageHandlers(mux, s); err != nil {
		return nil, err
	}

	return mux, nil
}
