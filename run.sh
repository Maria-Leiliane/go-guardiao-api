#!/bin/bash

echo "Construindo e iniciando a arquitetura Guardião da Saúde..."

# 1. Constroi a imagem do backend Go
echo "Passo 1: Construindo a imagem do backend Go..."
# Usamos 'sudo' para garantir as permissões necessárias
sudo docker build -t go-guardiao-api .

# 2. Inicia os serviços (DB, Cache, API e Worker)
echo "Passo 2: Subindo os containers (PostgreSQL e Redis estão sendo verificados)..."
# Usamos 'sudo' e a sintaxe moderna de 'docker compose'
sudo docker compose up --build -d

echo "----------------------------------------------------"
echo "Arquitetura Go Iniciada:"
echo "API (Síncrona): http://localhost:8080/api/v1/auth/login"
echo "Worker (Assíncrono): Monitorando logs para cálculo de Mana"
echo "----------------------------------------------------"
echo "Para ver os logs: sudo docker compose logs -f"
echo "Para parar: sudo docker compose down"
echo "----------------------------------------------------"
echo "Agora: acesse http://localhost:8080/api/v1/auth/login"

