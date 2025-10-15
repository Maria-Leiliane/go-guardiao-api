# Variáveis de ambiente
ENV_FILE ?= .env.development

# Build da imagem Docker
.PHONY: build
build:
	docker build -t go-guardiao-api .

# Sobe todos os serviços em modo detach
.PHONY: up
up:
	docker compose --env-file $(ENV_FILE) up --build -d

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

# Executa testes unitários do Go (dentro do host)
.PHONY: test
test:
	go test ./...

# Limpa volumes/dados (cuidado: apaga tudo!)
.PHONY: clean
clean:
	docker compose down -v

# Acesso ao shell do container API
.PHONY: shell-api
shell-api:
	docker exec -it guardian_api /bin/sh

# Acesso ao shell do container DB
.PHONY: shell-db
shell-db:
	docker exec -it guardian_db sh

# Build e sobe tudo (atalho)
.PHONY: run
run: build up

# Troca para ambiente de produção
.PHONY: prod
prod:
	$(MAKE) up ENV_FILE=.env.production

# Ajuda
.PHONY: help
help:
	@echo "Comandos principais:"
	@echo "  make build     # Builda a imagem Docker"
	@echo "  make up        # Sobe containers (default=dev)"
	@echo "  make down      # Para containers"
	@echo "  make logs      # Logs dos containers"
	@echo "  make ps        # Status dos containers"
	@echo "  make test      # Roda os testes Go"
	@echo "  make clean     # Remove containers e volumes"
	@echo "  make shell-api # Acessa o shell da API"
	@echo "  make shell-db  # Acessa o shell do DB"
	@echo "  make prod      # Sobe usando .env.production"