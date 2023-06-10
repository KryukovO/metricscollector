.PHONY: build test lint

build:
	go build -o cmd/agent/agent cmd/agent/main.go
	go build -o cmd/server/server cmd/server/main.go

test:
	go test -v -timeout 30s -race ./...

lint:
	golangci-lint run ./...