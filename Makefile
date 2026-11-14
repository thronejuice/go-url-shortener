.PHONY: run test lint build up down

run:
	go run ./cmd/server

test:
	go test -race -cover ./...

lint:
	golangci-lint run

build:
	CGO_ENABLED=0 go build -o bin/server ./cmd/server

up:
	docker compose up --build -d

down:
	docker compose down -v
