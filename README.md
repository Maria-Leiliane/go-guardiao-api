# 🛡️ Guardião da Saúde - Backend (GoLang & AWS Architecture)

Este repositório contém o **backend completo** do projeto **Guardião da Saúde**, desenvolvido em **Go (GoLang)** e projetado para rodar em uma arquitetura **robusta e escalável baseada em microsserviços** na AWS (simulada via Docker Compose).

---

## 🌟 Visão Geral da Arquitetura

A aplicação adota uma **arquitetura orientada a serviços** para garantir **alta disponibilidade, desacoplamento e escalabilidade**.

| Camada              | Tecnologia Principal        | Função Chave                                                       |
|---------------------|-----------------------------|--------------------------------------------------------------------|
| **API Síncrona**    | Go (Gorilla Mux)            | Autenticação, Perfis, CRUD de Hábitos e Gestão de Mana.            |
| **Persistência**    | PostgreSQL (RDS) & Redis    | Armazenamento de dados e cache de performance (ElastiCache).       |
| **Worker Assíncrono** | Go                         | Processamento assíncrono de eventos de Mana e Logs de Hábitos.     |
| **Infraestrutura**  | Docker Compose              | Orquestração local do DB, Cache e Aplicações Go.                   |

---

## 🚀 Como Rodar o Projeto Localmente (Docker)

A arquitetura foi projetada para ser iniciada com **um único comando**, simulando perfeitamente o ambiente de produção (PostgreSQL como RDS e Redis como ElastiCache).

### 🔧 Pré-requisitos
- Docker Desktop (ou Docker Engine + Docker Compose)
- Terminal Linux/macOS (ou WSL no Windows)

### ▶️ Passos de Inicialização

```bash
# Garante permissão de execução
chmod +x run.sh

# Constrói as imagens Go e inicia todos os containers (DB, Cache, API, Worker)
./run.sh
