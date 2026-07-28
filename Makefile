.PHONY: run build up down logs fmt tidy seed

run:
	go run ./cmd/server

build:
	go build -o bin/jobhoo ./cmd/server

up:
	docker compose up --build

down:
	docker compose down -v

logs:
	docker compose logs -f app

seed:
	docker compose run --rm app ./jobhoo-seed

fmt:
	gofmt -l -w .

tidy:
	go mod tidy
