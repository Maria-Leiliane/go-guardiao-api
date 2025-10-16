# Makefile simples, com comandos curtos e checks inteligentes

# Nome do projeto (isola stacks dev/prd)
PROJECT_NAME ?= guardiao

# Arquivos de env padrão (sem precisar passar ENV_FILE)
DEV_ENV_FILE ?= docker/.env.development
PRD_ENV_FILE ?= docker/.env.production

# Compose files (base + overlays)
COMPOSE_BASE := -f docker/docker-compose.yml
COMPOSE_DEV  := $(COMPOSE_BASE) -f docker/docker-compose.dev.yml
COMPOSE_PRD  := $(COMPOSE_BASE) -f docker/docker-compose.prd.yml

# Comandos docker compose prontos
DC_DEV := docker compose --env-file $(DEV_ENV_FILE) -p $(PROJECT_NAME)-dev $(COMPOSE_DEV)
DC_PRD := docker compose --env-file $(PRD_ENV_FILE) -p $(PROJECT_NAME)-prd $(COMPOSE_PRD)

# Auto-detecção dos Dockerfiles (suporta hífen ou ponto)
DOCKERFILES_DIR ?= docker
DOCKERFILE_API    := $(firstword $(wildcard $(DOCKERFILES_DIR)/Dockerfile-api $(DOCKERFILES_DIR)/Dockerfile.api))
DOCKERFILE_WORKER := $(firstword $(wildcard $(DOCKERFILES_DIR)/Dockerfile-worker $(DOCKERFILES_DIR)/Dockerfile.worker))

# ========= Alvos principais =========
.PHONY: up down logs ps up-dev up-prd down-dev down-prd logs-dev logs-prd ps-dev ps-prd clean-dev clean-prd build build-api build-worker test doctor

# Aliases simples (dev por padrão)
up: up-dev
down: down-dev
logs: logs-dev
ps: ps-dev

# Dev
up-dev: doctor
	$(DC_DEV) up --build -d
down-dev:
	$(DC_DEV) down
logs-dev:
	$(DC_DEV) logs -f
ps-dev:
	$(DC_DEV) ps
clean-dev:
	$(DC_DEV) down -v

# Prod (overlay local)
up-prd: doctor
	$(DC_PRD) up --build -d
down-prd:
	$(DC_PRD) down
logs-prd:
	$(DC_PRD) logs -f
ps-prd:
	$(DC_PRD) ps
clean-prd:
	$(DC_PRD) down -v

# Builds diretos (sem compose)
build: build-api build-worker
build-api:
	@if [ -z "$(DOCKERFILE_API)" ]; then echo "ERRO: Dockerfile da API não encontrado em $(DOCKERFILES_DIR)/Dockerfile-api ou Dockerfile.api"; exit 1; fi
	docker build -f $(DOCKERFILE_API) -t $(PROJECT_NAME)-api:local .
build-worker:
	@if [ -z "$(DOCKERFILE_WORKER)" ]; then echo "ERRO: Dockerfile do Worker não encontrado em $(DOCKERFILES_DIR)/Dockerfile-worker ou Dockerfile.worker"; exit 1; fi
	docker build -f $(DOCKERFILE_WORKER) -t $(PROJECT_NAME)-worker:local .

# Testes Go (no host)
test:
	go test ./...

# Verificações e diagnóstico
doctor:
	@echo "==> Verificando arquivos essenciais..."
	@for f in docker/docker-compose.yml docker/docker-compose.dev.yml docker/docker-compose.prd.yml; do \
		[ -f $$f ] || { echo "ERRO: Arquivo ausente: $$f"; exit 1; }; \
	done
	@if [ -z "$(DOCKERFILE_API)" ]; then \
		echo "ERRO: Dockerfile da API não encontrado."; \
		echo "Procure um destes: $(DOCKERFILES_DIR)/Dockerfile-api OU $(DOCKERFILES_DIR)/Dockerfile.api"; \
		exit 1; \
	fi
	@if [ -z "$(DOCKERFILE_WORKER)" ]; then \
		echo "ERRO: Dockerfile do Worker não encontrado."; \
		echo "Procure um destes: $(DOCKERFILES_DIR)/Dockerfile-worker OU $(DOCKERFILES_DIR)/Dockerfile.worker"; \
		exit 1; \
	fi
	@echo "OK: Compose e Dockerfiles encontrados."
	@echo "  API Dockerfile:    $(DOCKERFILE_API)"
	@echo "  Worker Dockerfile: $(DOCKERFILE_WORKER)"
	@echo "  Dev env-file:      $(DEV_ENV_FILE)"
	@echo "  Prd env-file:      $(PRD_ENV_FILE)"
	@echo "Dica: use 'make up' (dev), 'make down', 'make logs'."