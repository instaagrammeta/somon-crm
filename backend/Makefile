.PHONY: help dev build run migrate-up migrate-down lint test docker-up docker-down clean

APP=somon-crm-backend
GO=go

help:
	@echo "make dev          - run server with live reload (requires air)"
	@echo "make build        - build binaries"
	@echo "make run          - run server"
	@echo "make migrate-up   - run migrations up"
	@echo "make migrate-down - rollback migrations"
	@echo "make lint         - run go vet & go fmt"
	@echo "make test         - run go tests"
	@echo "make docker-up    - start full stack via docker-compose"
	@echo "make docker-down  - stop docker-compose stack"

build:
	$(GO) build -o bin/server ./cmd/server
	$(GO) build -o bin/migrate ./cmd/migrate

run:
	$(GO) run ./cmd/server

migrate-up:
	$(GO) run ./cmd/migrate -direction=up

migrate-down:
	$(GO) run ./cmd/migrate -direction=down

lint:
	$(GO) vet ./...
	$(GO) fmt ./...

test:
	$(GO) test ./... -v -race -count=1

docker-up:
	docker compose up -d --build

docker-down:
	docker compose down

clean:
	rm -rf bin/
