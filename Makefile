APP_NAME          := avito-trainee
BIN_DIR           := bin
BIN               := $(BIN_DIR)/$(APP_NAME)
CMD               := ./cmd/main.go
SWAG              ?= swag
COMPOSE_FILE      ?= docker-compose.yml
INTEGRATION_COMPOSE ?= test/docker-compose.integration.yml
ENV_FILE          ?= .env
ENV_EXAMPLE       ?= .env-example

.PHONY: up build run clean docs docker-build docker-up docker-up-detached docker-down docker-logs \
	unit-test unit-cover integration-test integration-cover ensure-env

# Создание .env при первом запуске (копируем из .env-example или используем дефолт).
ensure-env:
	@if [ -f "$(ENV_FILE)" ]; then \
		echo "$(ENV_FILE) exists"; \
	elif [ -f "$(ENV_EXAMPLE)" ]; then \
		cp $(ENV_EXAMPLE) $(ENV_FILE) && echo "Created $(ENV_FILE) from $(ENV_EXAMPLE)"; \
	else \
		echo "APP_PORT=8080\nDB_HOST=localhost\nDB_PORT=5432\nDB_USER=app\nDB_PASS=app\nDB_NAME=app\nLOG_LEVEL=info" > $(ENV_FILE) && \
		echo "Created $(ENV_FILE) with defaults"; \
	fi

build: $(BIN)

$(BIN):
	mkdir -p $(BIN_DIR)
	go build -o $(BIN) $(CMD)

docs:
	$(SWAG) init -g cmd/main.go -o docs

# Полный запуск: подготовка .env, генерация docs, сборка, рестарт контейнеров.
up: ensure-env docs build docker-down docker-up-detached

run: build
	if [ -f $(ENV_FILE) ]; then export $$(grep -v '^#' $(ENV_FILE) | xargs); fi; $(BIN)

clean:
	rm -rf $(BIN_DIR) cover.out integration-cover.out coverage.html integration-coverage.html

unit-test:
	go test ./internal/service ./internal/repository ./internal/controller/handlers ./internal/mapper ./internal/config ./internal/db ./internal/model ./internal/service/errs ./internal/controller/dto

unit-cover:
	go test ./internal/... -coverprofile=cover.out
	go tool cover -html=cover.out -o coverage.html
	@echo "Coverage report: coverage.html"

integration-test:
	docker compose -f $(INTEGRATION_COMPOSE) down
	docker compose -f $(INTEGRATION_COMPOSE) up -d
	go test ./test
	docker compose -f $(INTEGRATION_COMPOSE) down

integration-cover:
	docker compose -f $(INTEGRATION_COMPOSE) down
	docker compose -f $(INTEGRATION_COMPOSE) up -d
	go test ./test -coverpkg=./internal/service/...,./internal/repository/...,./internal/controller/handlers/... -coverprofile=integration-cover.out
	go tool cover -html=integration-cover.out -o integration-coverage.html
	@echo "Integration coverage report: integration-coverage.html"
	docker compose -f $(INTEGRATION_COMPOSE) down

docker-build:
	docker build -t $(APP_NAME):local .

docker-up:
	docker compose -f $(COMPOSE_FILE) up

docker-up-detached:
	docker compose -f $(COMPOSE_FILE) up -d

docker-down:
	docker compose -f $(COMPOSE_FILE) down

docker-logs:
	docker compose -f $(COMPOSE_FILE) logs -f app
