SHELL := /bin/bash
RUNNERS_COMPOSE := docker compose -f docker-compose.runners.yml

.PHONY: help install install-frontend run-backend run-frontend dev build build-backend build-frontend test test-backend test-frontend docker-build docker-up docker-down docker-logs runners-build runners-up-all runners-up-go runners-up-javascript runners-up-python runners-down clean

help:
	@echo "Learning Platform - Make Targets"
	@echo ""
	@echo "Setup:"
	@echo "  make install           Install frontend dependencies"
	@echo ""
	@echo "Run:"
	@echo "  make run-backend       Run Go backend API (port from env or 8080)"
	@echo "  make run-frontend      Run React frontend dev server"
	@echo "  make dev               Run backend and frontend together"
	@echo "  make docker-up         Run full stack in Docker containers"
	@echo ""
	@echo "Build:"
	@echo "  make build             Build backend and frontend"
	@echo "  make build-backend     Compile backend binary"
	@echo "  make build-frontend    Build frontend production bundle"
	@echo "  make docker-build      Build full-stack Docker images"
	@echo ""
	@echo "Test:"
	@echo "  make test              Run backend and frontend tests"
	@echo "  make test-backend      Run Go tests"
	@echo "  make test-frontend     Run frontend tests if configured"
	@echo ""
	@echo "Runners:"
	@echo "  make runners-build     Build all sandbox runner images"
	@echo "  make runners-up-all    Start all sandbox runners"
	@echo "  make runners-up-go     Start only Go runner"
	@echo "  make runners-up-javascript Start only JavaScript runner"
	@echo "  make runners-up-python Start only Python runner"
	@echo "  make runners-down      Stop all sandbox runners"
	@echo "  make docker-down       Stop full stack containers"
	@echo "  make docker-logs       Tail full stack container logs"
	@echo ""
	@echo "Misc:"
	@echo "  make clean             Remove generated backend binary and frontend dist"

install: install-frontend

install-frontend:
	cd frontend && pnpm install

run-backend:
	cd backend && go run ./cmd/server

run-frontend:
	cd frontend && pnpm dev

dev:
	@set -e; \
	trap 'kill 0' INT TERM EXIT; \
	$(MAKE) --no-print-directory run-backend & \
	$(MAKE) --no-print-directory run-frontend & \
	wait

build: build-backend build-frontend

build-backend:
	cd backend && go build -o bin/server ./cmd/server

build-frontend:
	cd frontend && pnpm build

test: test-backend test-frontend

test-backend:
	cd backend && go test ./...

test-frontend:
	cd frontend && if pnpm run | grep -q " test"; then pnpm test; else echo "No frontend test script configured, skipping."; fi

docker-build:
	docker compose build

docker-up:
	docker compose up -d --build

docker-down:
	docker compose down

docker-logs:
	docker compose logs -f --tail=200

runners-build:
	$(RUNNERS_COMPOSE) build

runners-up-all:
	$(RUNNERS_COMPOSE) --profile all up -d

runners-up-go:
	$(RUNNERS_COMPOSE) --profile go up -d runner-go

runners-up-javascript:
	$(RUNNERS_COMPOSE) --profile javascript up -d runner-javascript

runners-up-python:
	$(RUNNERS_COMPOSE) --profile python up -d runner-python

runners-down:
	$(RUNNERS_COMPOSE) down

clean:
	rm -f backend/bin/server
	rm -rf frontend/dist
