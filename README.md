# 🛡️ Guardião da Saúde - Backend (GoLang & AWS Architecture)

Este repositório contém o **backend completo** do projeto **Guardião da Saúde**, desenvolvido em **Go (GoLang)** e projetado para rodar em uma arquitetura **robusta, resiliente e escalável** baseada em microsserviços na AWS (simulada via Docker Compose).

---

## 🌟 Visão Geral da Arquitetura

A aplicação adota uma arquitetura orientada a serviços para garantir **alta disponibilidade, desacoplamento e escalabilidade**.

| Camada                 | Tecnologia Principal        | Função Chave                                                           |
|------------------------|----------------------------|-----------------------------------------------------------------------|
| **API Síncrona**       | Go (Gorilla Mux)           | Autenticação, Perfis, CRUD de Hábitos e Gestão de Mana.               |
| **Persistência**       | PostgreSQL (RDS) & Redis   | Armazenamento de dados e cache de performance (ElastiCache).          |
| **Worker Assíncrono**  | Go                         | Processamento assíncrono de eventos de Mana e Logs de Hábitos.        |
| **Infraestrutura**     | Docker Compose             | Orquestração local do DB, Cache e Aplicações Go.                      |

---

## 🖥️ Diagrama da Arquitetura

```mermaid
flowchart LR
    subgraph Backend Go
        API["API (Gorilla Mux)"]
        Worker["Worker Assíncrono"]
    end

    Client[Usuário / Frontend] -->|HTTP/REST| API

    API <--> |Postgres| DB[(PostgreSQL)]
    API <--> |Redis| Cache[(Redis)]

    Worker <--> |Postgres| DB
    Worker <--> |Redis| Cache

    API -- Mensagens/Jobs --> Worker

    classDef infra fill:#E7F6F2,stroke:#333,stroke-width:2px;
    DB,Cache class infra;
```

---

## 🚀 Como Rodar o Projeto Localmente

A arquitetura foi projetada para ser iniciada com **um único comando**, simulando perfeitamente o ambiente de produção (PostgreSQL como RDS e Redis como ElastiCache).

### 🔧 Pré-requisitos

- Docker Desktop (ou Docker Engine + Docker Compose)
- Terminal Linux/macOS (ou WSL no Windows)
- Go 1.20+ (opcional, apenas para desenvolvimento local)

---

### ▶️ Inicialização Rápida

```bash
# Garante permissão de execução
chmod +x run.sh

# Constrói as imagens Go e inicia todos os containers (DB, Cache, API, Worker)
./run.sh
```

> **Dica:** Para times, o projeto possui um [Makefile](./Makefile) com vários comandos úteis para build, up, down, logs, testes, etc.

---

### 💻 Principais Comandos do Makefile

```bash
make up         # Sobe todos os serviços em background
make logs       # Logs em tempo real
make down       # Para containers
make build      # Build das imagens Docker
make test       # Executa os testes Go
make prod       # Sobe usando .env.production
```

---

## ⚙️ Estrutura dos Serviços

- **/cmd/api/**: Entrypoint da API HTTP.
- **/cmd/gamification_worker/**: Entrypoint do worker assíncrono.
- **/internal/**: Domínios de regras de negócio, autenticação, banco, cache, etc.
- **/pkg/**: Modelos e utilitários compartilhados.
- **docker-compose.yml**: Orquestração local dos serviços.
- **Dockerfile**: Build multi-stage para API e Worker.
- **.env.example**: Modelo de variáveis de ambiente.

---

## 🔒 Segurança & Boas Práticas

- Segredos nunca versionados (.env.production fora do git!)
- Senhas e JWT gerados aleatoriamente.
- Banco e Redis isolados em rede privada.
- Healthchecks para todos os serviços.
- Imagem Docker mínima (Alpine, usuário não-root).

---

## 🛠️ Contribuições

Pull requests são bem-vindos! Siga as convenções de commit e abra issues para bugs e sugestões.  
Antes de contribuir, leia o [CONTRIBUTING.md](./CONTRIBUTING.md) se disponível.

---

## 📦 Deploy em Produção

- Use `make prod` ou `docker compose --env-file .env.production up -d`
- Configure variáveis reais e seguras em `.env.production`
- Para auto scaling e alta disponibilidade, utilize Docker Swarm ou Kubernetes.

---

## 📚 Licença

MIT © [Guardião da Saúde](https://github.com/Maria-Leiliane)

---