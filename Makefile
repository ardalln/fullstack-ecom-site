.PHONY: db-up db-down tidy run build

db-up:
	docker compose up -d

db-down:
	docker compose down

tidy:
	go mod tidy

run:
	go run ./cmd/api

build:
	go build -o bin/shop-api ./cmd/api
