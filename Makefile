.PHONY: run db seed build

run:
	@go run cmd/api/*.go

build:
	@go build cmd/api/*.go

db:
	@psql postgres://admin:admin123@localhost:5432/gopher-social-network

seed:
	@go run cmd/migrate/seed/main.go
