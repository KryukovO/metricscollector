package mocks

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
)

type MockServer struct {
	*httptest.Server
}

func NewMockServer() *MockServer {
	server := httptest.NewServer(http.HandlerFunc(updatesHandler))

	return &MockServer{
		Server: server,
	}
}

func (a MockServer) Close() {
	a.Server.Close()
}

func updatesHandler(w http.ResponseWriter, r *http.Request) {
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) != 1 || pathParts[0] != "updates" {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(fmt.Sprintf("Unexected path: %s", r.URL.Path)))
	}

	w.WriteHeader(http.StatusOK)
}
