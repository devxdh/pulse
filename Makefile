DB_SERVICE_NAME=postgres

start: check-db
	go run cmd/api/main.go

check-db:
	@echo "Checking database status..."
	@if [ -z "$$(docker compose ps -q $(DB_SERVICE_NAME))" ] || [ -z "$$(docker compose ps --filter "status=running" -q $(DB_SERVICE_NAME))" ]; then \
		echo "Database is not running. Starting database..."; \
		docker compose up -d; \
		echo "Waiting for database to be ready..."; \
		sleep 5; \
	else \
		echo "Database is already running."; \
	fi

test:
	go test -v ./...

cover:
	go cover -v ./...

start-db:
	docker compose up -d

stop-db:
	docker compose down

clean-db:
	docker compose down -v
