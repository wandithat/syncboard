.PHONY: run migrate-up migrate-down docker-up docker-down test tidy

run:
	go run ./cmd/api

docker-up:
	docker compose up -d postgres

docker-down:
	docker compose down

migrate-up:
	migrate -path migrations -database "postgres://postgres:postgres123@localhost:5432/syncboard?sslmode=disable" up

migrate-down:
	migrate -path migrations -database "postgres://postgres:postgres123@localhost:5432/syncboard?sslmode=disable" down

test:
	go test ./...

tidy:
	go mod tidy
