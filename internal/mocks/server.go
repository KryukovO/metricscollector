// Package mocks содержит мок-объекты.
package mocks

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
)

// MockServer - мок модуля-сервера.
type MockServer struct {
	*httptest.Server
}

// NewMockServer создаёт новый мок модуля-сервера.
func NewMockServer() *MockServer {
	server := httptest.NewServer(http.HandlerFunc(updatesHandler))

	return &MockServer{
		Server: server,
	}
}

// Close выполняет закрытие мока модуля-сервера.
func (s MockServer) Close() {
	s.Server.Close()
}

// updatesHandler представляет собой "подмену" налогичного обработчика модуля-сервера.
func updatesHandler(w http.ResponseWriter, r *http.Request) {
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) != 1 || pathParts[0] != "updates" {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "Unexected path: %s", r.URL.Path)
	}

	w.WriteHeader(http.StatusOK)
}
