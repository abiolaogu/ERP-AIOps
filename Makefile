.PHONY: build dev test docker-build docker-up docker-down migrate gateway api ai-brain web lint smoke

build:
	cd cmd/server && go build -o ../../bin/aiops-gateway .
	cd crates/opstrac-api && cargo build --release

dev:
	cd infra && docker compose up --build

test:
	go test ./...
	cargo test --workspace
	cd services/ai-brain && python -m pytest
	cd web && npm run test

docker-build:
	cd infra && docker compose build

docker-up:
	cd infra && docker compose up -d

docker-down:
	cd infra && docker compose down

migrate:
	@echo "Running migrations against YugabyteDB..."
	ysqlsh -h localhost -p 5433 -U yugabyte -d erp_aiops -f migrations/001_initial_schema.sql

gateway:
	cd cmd/server && go run .

api:
	cd crates/opstrac-api && cargo run

ai-brain:
	cd services/ai-brain && uvicorn app.main:app --host 0.0.0.0 --port 8001 --reload

web:
	cd web && npm run dev

lint:
	cd cmd/server && go vet ./...
	cargo clippy --workspace
	cd services/ai-brain && ruff check .
	cd web && npm run lint

smoke:
	./scripts/smoke.sh

test-all:
	./scripts/test-all.sh
