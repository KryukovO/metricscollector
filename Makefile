.PHONY: build

build:
	go build -o cmd/agent/agent cmd/agent/main.go
	go build -o cmd/server/server cmd/server/main.go

.PHONY: inc1