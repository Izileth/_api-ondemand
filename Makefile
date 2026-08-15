.PHONY: all build run stop clean test logs help

# ─── Variables ────────────────────────────────────────────────────────────────
GO      = go
SERVERS = server1 server2 server3

# ─── Default ──────────────────────────────────────────────────────────────────
all: help

help: ## Show this help
	@echo ""
	@echo "  api-ondemand — Load Balancer Project"
	@echo ""
	@awk 'BEGIN {FS = ":.*##"; printf "  \033[36m%-20s\033[0m %s\n", "Target", "Description"} \
	      /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2 }' $(MAKEFILE_LIST)
	@echo ""

# ─── Build ────────────────────────────────────────────────────────────────────
build: ## Build all 3 Go servers
	@for srv in $(SERVERS); do \
		echo "→ Building $$srv..."; \
		$(GO) build -ldflags="-s -w" -o $$srv/$$srv ./$$srv/; \
	done
	@echo "✓ All servers built."

# ─── Run locally (without Docker) ────────────────────────────────────────────
run-servers: ## Start all 3 Go servers in background
	@echo "→ Starting server1 on :8080..."
	@PORT=8080 SERVER_NAME="Server 1" $(GO) run ./server1/ &
	@echo "→ Starting server2 on :8081..."
	@PORT=8081 SERVER_NAME="Server 2" $(GO) run ./server2/ &
	@echo "→ Starting server3 on :8082..."
	@PORT=8082 SERVER_NAME="Server 3" $(GO) run ./server3/ &
	@echo "✓ All servers running. Use 'make stop' to terminate."

stop: ## Stop all running Go servers
	@pkill -f "go run ./server" 2>/dev/null || true
	@pkill -f "server1\|server2\|server3" 2>/dev/null || true
	@echo "✓ Servers stopped."

# ─── Docker ───────────────────────────────────────────────────────────────────
docker-up: ## Build and start all services with Docker Compose
	docker compose up --build -d
	@echo "✓ Stack up at http://localhost"

docker-down: ## Stop and remove all containers
	docker compose down
	@echo "✓ Stack down."

docker-logs: ## Follow logs from all containers
	docker compose logs -f

docker-ps: ## List running containers and their status
	docker compose ps

# ─── Testing ──────────────────────────────────────────────────────────────────
test: ## Run a quick smoke test against all endpoints
	@echo ""
	@echo "── Smoke Test ─────────────────────────────────────────────────────"
	@for port in 8080 8081 8082; do \
		echo ""; \
		echo "  Server :$$port"; \
		echo "  /ping   → $$(curl -sf http://localhost:$$port/ping)"; \
		echo "  /health → $$(curl -sf http://localhost:$$port/health)"; \
		echo "  /info   → $$(curl -sf http://localhost:$$port/info | head -c 120)..."; \
	done
	@echo ""
	@echo "  Load Balancer :80"; \
	@for i in 1 2 3 4 5 6; do \
		echo "  req $$i → $$(curl -sf http://localhost/ping)"; \
	done
	@echo "── Done ────────────────────────────────────────────────────────────"

load-test: ## Run a quick load test using hey (install: go install github.com/rakyll/hey@latest)
	@command -v hey >/dev/null 2>&1 || { echo "hey not found. Install: go install github.com/rakyll/hey@latest"; exit 1; }
	@echo "→ Load testing http://localhost/ — 500 requests, 50 concurrent..."
	hey -n 500 -c 50 http://localhost/ping

# ─── Utilities ────────────────────────────────────────────────────────────────
clean: ## Remove compiled binaries
	@for srv in $(SERVERS); do \
		rm -f $$srv/$$srv; \
	done
	@echo "✓ Cleaned."

metrics: ## Show live metrics from all servers
	@for port in 8080 8081 8082; do \
		echo "── Server :$$port metrics ──────────────────────────────────────────"; \
		curl -sf http://localhost:$$port/metrics | python3 -m json.tool 2>/dev/null || curl -sf http://localhost:$$port/metrics; \
		echo ""; \
	done
