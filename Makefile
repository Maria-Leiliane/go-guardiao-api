# Makefile simples e silencioso com targets E2E

PROJECT_NAME ?= guardiao
DEV_ENV_FILE ?= docker/.env.development
PRD_ENV_FILE ?= docker/.env.production
E2E_ENV_FILE ?= docker/.env.e2e

COMPOSE_BASE := -f docker/docker-compose.yml
COMPOSE_DEV  := $(COMPOSE_BASE) -f docker/docker-compose.dev.yml
COMPOSE_PRD  := $(COMPOSE_BASE) -f docker/docker-compose.prd.yml
COMPOSE_E2E  := $(COMPOSE_BASE) -f docker/docker-compose.e2e.yml

DC_DEV := docker compose --env-file $(DEV_ENV_FILE) -p $(PROJECT_NAME)-dev $(COMPOSE_DEV)
DC_PRD := docker compose --env-file $(PRD_ENV_FILE) -p $(PROJECT_NAME)-prd $(COMPOSE_PRD)
DC_E2E := docker compose --env-file $(E2E_ENV_FILE) -p $(PROJECT_NAME)-e2e $(COMPOSE_E2E)

DOCKERFILES_DIR ?= docker
DOCKERFILE_API    := $(firstword $(wildcard $(DOCKERFILES_DIR)/Dockerfile-api $(DOCKERFILES_DIR)/Dockerfile.api))
DOCKERFILE_WORKER := $(firstword $(wildcard $(DOCKERFILES_DIR)/Dockerfile-worker $(DOCKERFILES_DIR)/Dockerfile.worker))

.SILENT:

.PHONY: up down logs ps up-dev up-prd down-dev down-prd logs-dev logs-prd ps-dev ps-prd clean-dev clean-prd build build-api build-worker test doctor help \
        e2e e2e-up e2e-test e2e-down logs-e2e ps-e2e

# Aliases curtos (dev por padrão)
up: up-dev
down: down-dev
logs: logs-dev
ps: ps-dev

# Dev
up-dev: doctor
	@echo "🚀 Subindo DEV..."
	$(DC_DEV) up --build -d
	@echo "✅ DEV no ar. Health: http://localhost:8080/health"

down-dev:
	@echo "🛑 Derrubando DEV..."
	$(DC_DEV) down

logs-dev:
	$(DC_DEV) logs -f

ps-dev:
	$(DC_DEV) ps

clean-dev:
	@echo "🧹 Limpando DEV (containers + volumes)..."
	$(DC_DEV) down -v

# Prod (overlay local)
up-prd: doctor
	@echo "🚀 Subindo PRD (local overlay)..."
	$(DC_PRD) up --build -d

down-prd:
	@echo "🛑 Derrubando PRD..."
	$(DC_PRD) down

logs-prd:
	$(DC_PRD) logs -f

ps-prd:
	$(DC_PRD) ps

clean-prd:
	@echo "🧹 Limpando PRD (containers + volumes)..."
	$(DC_PRD) down -v

# E2E
e2e: e2e-up e2e-test e2e-down

e2e-up: doctor
	@echo "🚀 Subindo stack E2E (porta 18080)..."
	$(DC_E2E) up --build -d

e2e-test:
	@echo "⏳ Aguardando API (http://localhost:18080/health) ficar OK..."
	@sh -c 'i=0; until curl -fsS http://localhost:18080/health >/dev/null 2>&1; do \
	  i=$$((i+1)); [ $$i -gt 60 ] && echo "API não respondeu em tempo hábil" && exit 1; sleep 2; done'
	@echo "🧪 Rodando testes E2E..."
	E2E_BASE_URL=http://localhost:18080/api/v1 go test ./tests/e2e -v -count=1 -timeout=10m

e2e-down:
	@echo "🧹 Derrubando stack E2E (containers + volumes)..."
	$(DC_E2E) down -v

logs-e2e:
	$(DC_E2E) logs -f

ps-e2e:
	$(DC_E2E) ps

# Builds diretos (sem compose)
build: build-api build-worker

build-api:
	@if [ -z "$(DOCKERFILE_API)" ]; then echo "ERRO: Dockerfile da API não encontrado em docker/Dockerfile-api ou docker/Dockerfile.api"; exit 1; fi
	@echo "🔨 Build API -> $(DOCKERFILE_API)"
	docker build -f $(DOCKERFILE_API) -t $(PROJECT_NAME)-api:local .

build-worker:
	@if [ -z "$(DOCKERFILE_WORKER)" ]; then echo "ERRO: Dockerfile do Worker não encontrado em docker/Dockerfile-worker ou docker/Dockerfile.worker"; exit 1; fi
	@echo "🔨 Build Worker -> $(DOCKERFILE_WORKER)"
	docker build -f $(DOCKERFILE_WORKER) -t $(PROJECT_NAME)-worker:local .

# Testes Go (no host)
test:
	go test ./...

# Verificações e diagnóstico
doctor:
	@echo "🔎 Verificando arquivos essenciais..."
	@for f in docker/docker-compose.yml docker/docker-compose.dev.yml docker/docker-compose.prd.yml docker/docker-compose.e2e.yml; do \
		[ -f $$f ] || { echo "ERRO: Arquivo ausente: $$f"; exit 1; }; \
	done
	@[ -f "$(E2E_ENV_FILE)" ] || { echo "ERRO: Arquivo ausente: $(E2E_ENV_FILE)"; exit 1; }
	@if [ -z "$(DOCKERFILE_API)" ]; then \
		echo "ERRO: Dockerfile da API não encontrado (docker/Dockerfile-api ou docker/Dockerfile.api)"; exit 1; \
	fi
	@if [ -z "$(DOCKERFILE_WORKER)" ]; then \
		echo "ERRO: Dockerfile do Worker não encontrado (docker/Dockerfile-worker ou docker/Dockerfile.worker)"; exit 1; \
	fi
	@echo "OK: Compose e Dockerfiles encontrados."
	@echo "  API Dockerfile:    $(DOCKERFILE_API)"
	@echo "  Worker Dockerfile: $(DOCKERFILE_WORKER)"
	@echo "  Dev env-file:      $(DEV_ENV_FILE)"
	@echo "  Prd env-file:      $(PRD_ENV_FILE)"
	@echo "  E2E env-file:      $(E2E_ENV_FILE)"
	@echo "Dica: use 'make up', 'make e2e', 'make logs'."

help:
	@echo "Comandos:"
	@echo "  make up | down | logs | ps           # DEV rápido"
	@echo "  make up-prd | down-prd | logs-prd    # PRD local"
	@echo "  make e2e | e2e-up | e2e-test | e2e-down | logs-e2e | ps-e2e"
	@echo "  make clean-dev | clean-prd           # remove volumes"
	@echo "  make build                           # build imagens direto"
	@echo "  make test                            # go test ./..."
	@echo "  make doctor                          # valida arquivos"