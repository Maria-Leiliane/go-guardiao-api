# Makefile para Guardião da Saúde - Ambiente Dev e Prod

# Variável de ambiente padrão para dev (pode ser sobrescrita)
ENV_FILE ?= .env.development

# Build da imagem Docker
.PHONY: build
build:
	docker build -t go-guardiao-api .

# Sobe todos os serviços em modo detach para desenvolvimento
.PHONY: up-dev
up-dev:
	docker compose -f docker-compose.yml -f docker-compose.dev.yml up --build -d

# Sobe todos os serviços em modo detach para produção
.PHONY: up-prd
up-prd:
	docker compose -f docker-compose.yml -f docker-compose.prd.yml up --build -d

# Para todos os serviços
.PHONY: down
down:
	docker compose down

# Logs dos containers em tempo real
.PHONY: logs
logs:
	docker compose logs -f

# Status dos containers
.PHONY: ps
ps:
	docker compose ps

# Executa testes unitários do Go (no host)
.PHONY: test
test:
	go test ./...

# Limpa volumes/dados (cuidado: apaga tudo!)
.PHONY: clean
clean:
	docker compose down -v

# Acesso ao shell do container API (ajuste para o nome real do container)
.PHONY: shell-api
shell-api:
	docker exec -it $$(docker compose ps -q api) /bin/sh

# Acesso ao shell do container DB (ajuste para o nome real do container)
.PHONY: shell-db
shell-db:
	docker exec -it $$(docker compose ps -q db) sh

# Ajuda
.PHONY: help
help:
	@echo "Comandos principais:"
	@echo "  make build       # Builda a imagem Docker"
	@echo "  make up-dev      # Sobe containers para desenvolvimento"
	@echo "  make up-prd      # Sobe containers para produção"
	@echo "  make down        # Para containers"
	@echo "  make logs        # Logs dos containers"
	@echo "  make ps          # Status dos containers"
	@echo "  make test        # Roda os testes Go"
	@echo "  make clean       # Remove containers e volumes"
	@echo "  make shell-api   # Acessa o shell da API"
	@echo "  make shell-db    # Acessa o shell do DB"

