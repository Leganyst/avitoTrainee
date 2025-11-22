APP_NAME := avito-trainee
BIN_DIR  := bin
BIN      := $(BIN_DIR)/$(APP_NAME)
CMD      := ./cmd/main.go
SWAG     ?= swag

.PHONY: build run test clean docs docker-build docker-up docker-up-detached docker-down docker-logs

build: $(BIN)

$(BIN):
	mkdir -p $(BIN_DIR)
	go build -o $(BIN) $(CMD)

run: build
	if [ -f .env ]; then export $$(grep -v '^#' .env | xargs); fi; $(BIN)

clean:
	rm -rf $(BIN_DIR)

docs:
	$(SWAG) init -g cmd/main.go -o docs

unit-test:
	go test ./internal/service -coverprofile=cover.out
	go tool cover -html=cover.out

test-integration:
	docker compose -f test/docker-compose.integration.yml down
	docker compose -f test/docker-compose.integration.yml up -d 
	go test ./test -coverpkg=./internal/service/...,./internal/repository/... -coverprofile=integration-cover.out
	go tool cover -html=integration-cover.out
	docker compose -f test/docker-compose.integration.yml down

docker-build:
	docker build -t $(APP_NAME):local .

docker-up:
	docker compose up

docker-up-detached:
	docker compose up -d

docker-down:
	docker compose down

docker-logs:
	docker compose logs -f app
