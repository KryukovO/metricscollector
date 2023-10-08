.PHONY: build test cover cover-html bench lint

BUILDDATE=$(shell date +'%d-%m-%Y')
BUILDVERSION=v0.0.20

build:
	go build -o cmd/agent/agent -ldflags "-X main.buildVersion=${BUILDVERSION} -X main.buildDate=${BUILDDATE}" cmd/agent/main.go
	go build -o cmd/server/server -ldflags "-X main.buildVersion=${BUILDVERSION} -X main.buildDate=${BUILDDATE}" cmd/server/main.go
	go build -o cmd/analyser/analyser cmd/analyser/main.go

test:
	go test -v -timeout 30s -race ./...

cover:
	go test -timeout 30s -race -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out
	rm coverage.out

cover-html:
	go test -timeout 30s -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	rm coverage.out

bench:
	go test -bench=. -benchmem ./...

lint:
	golangci-lint run ./...