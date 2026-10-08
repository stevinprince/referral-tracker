.PHONY: dev-backend dev-frontend dev \
       build-backend build-frontend build \
       test-backend test-frontend test \
       lint-backend lint-frontend lint \
       hash-password

# --- Development ---

dev-backend:
	cd backend && go run ./cmd/server

dev-frontend:
	cd frontend && npm run dev

dev:
	@echo "Run 'make dev-backend' and 'make dev-frontend' in separate terminals"

# --- Build ---

build-backend:
	cd backend && go build -o bin/server ./cmd/server

build-frontend:
	cd frontend && npm run build

build: build-frontend build-backend

# --- Test ---

test-backend:
	cd backend && go test ./... -v -count=1

test-frontend:
	cd frontend && npm run test

test: test-backend test-frontend

# --- Lint ---

lint-backend:
	cd backend && golangci-lint run

lint-frontend:
	cd frontend && npm run lint

lint: lint-backend lint-frontend

# --- Utilities ---

hash-password:
	@echo "Password hash utility coming in a later task"
