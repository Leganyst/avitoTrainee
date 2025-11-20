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

test:
	go test ./...

clean:
	rm -rf $(BIN_DIR)

docs:
	$(SWAG) init -g cmd/main.go -o docs

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
