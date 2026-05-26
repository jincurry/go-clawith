.PHONY: build run test clean dev deps

build:
	go build -o bin/server ./cmd/server

run: build
	./bin/server

dev:
	go run ./cmd/server

deps:
	go mod tidy
	go mod download

test:
	go test ./... -v

clean:
	rm -rf bin/

docker-up:
	docker compose up -d

docker-down:
	docker compose down

lint:
	golangci-lint run ./...
