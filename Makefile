GO_FILES := $(shell find backend -type f -name '*.go')

.PHONY: lint format test build migrate-up seed-admin

lint:
	@test -z "$$(gofmt -l $(GO_FILES))" || (gofmt -l $(GO_FILES); echo "Go files need formatting; run 'make format'."; exit 1)
	cd backend && go vet ./...

format:
	gofmt -w $(GO_FILES)

test:
	cd backend && go test -race ./...

build:
	mkdir -p backend/bin
	cd backend && go build -o bin/api ./cmd/api

migrate-up:
	cd backend && go run ./cmd/migrate up

seed-admin:
	cd backend && go run ./cmd/seed-admin
