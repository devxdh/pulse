-include .env
export

DB_SERVICE_NAME=postgres
COMPOSE_FILE=docker-compose.yml

start: check-db
	go run ./cmd/api

# The test command now just swaps the compose file and runs tests.
# Go natively reads TEST_DATABASE_URL because we included the .env file above!
test: COMPOSE_FILE=docker-compose.test.yml
test: check-db
	go test -v ./...
	docker compose -f $(COMPOSE_FILE) down

check-db:
	@echo "Ensuring database is running via $(COMPOSE_FILE)..."
	@docker compose -f $(COMPOSE_FILE) up -d $(DB_SERVICE_NAME)
	@echo "Waiting for database healthcheck..."
	@until [ "$$(docker inspect --format='{{.State.Health.Status}}' $$(docker compose -f $(COMPOSE_FILE) ps -q $(DB_SERVICE_NAME)))" = "healthy" ]; do \
		sleep 0.2; \
	done
	@echo "Database is ready!"

cover:
	go test -cover -v ./...

start-db:
	docker compose -f $(COMPOSE_FILE) up -d

stop-db:
	docker compose -f $(COMPOSE_FILE) down

clean-db:
	docker compose -f $(COMPOSE_FILE) down -v
