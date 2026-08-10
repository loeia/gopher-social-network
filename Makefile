.PHONY: run db seed build up down

run:
	@go run cmd/api/*.go

build:
	@go build cmd/api/*.go

db:
	@psql postgres://admin:admin123@localhost:5432/gopher-social-network

seed:
	@go run cmd/migrate/seed/main.go

up: 
	@docker compose up -d 
down: 
	@docker compose down 
