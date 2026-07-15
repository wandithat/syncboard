.PHONY: run migrate-up migrate-down docker-up docker-down test

run:
	go run cmd/api/main.go

docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

migrate-up:
	migrate -path migrations -database "postgres://postgres:postgres123@localhost:5432/syncboard?sslmode=disable" up

migrate-down:
	migrate -path migrations -database "postgres://postgres:postgres123@localhost:5432/syncboard?sslmode=disable" down

test:
	go test ./...
