#!/bin/bash

set -e

PORT=8080
API_URL="http://localhost:$PORT/api/v1/auth/login"

echo "Construindo e iniciando a arquitetura Guardião da Saúde..."

# Detecta se precisa de sudo para docker
DOCKER_CMD="docker"
if ! docker info >/dev/null 2>&1; then
  DOCKER_CMD="sudo docker"
fi

# Detecta ambiente (dev por padrão)
if [[ "$1" == "prod" ]]; then
  COMPOSE_FILES="-f docker-compose.yml -f docker-compose.prd.yml"
  ENV_LABEL="Produção"
else
  COMPOSE_FILES="-f docker-compose.yml -f docker-compose.dev.yml"
  ENV_LABEL="Desenvolvimento"
fi

echo "Ambiente selecionado: $ENV_LABEL"
echo "Passo 1: Build e subida dos containers..."
$DOCKER_CMD compose $COMPOSE_FILES up --build -d

echo "----------------------------------------------------"
echo "Arquitetura Go Iniciada ($ENV_LABEL):"
echo "API (Síncrona): $API_URL"
echo "Worker (Assíncrono): Monitorando logs para cálculo de Mana"
echo "----------------------------------------------------"
echo "Para ver os logs: $DOCKER_CMD compose $COMPOSE_FILES logs -f"
echo "Para parar:      $DOCKER_CMD compose $COMPOSE_FILES down"
echo "----------------------------------------------------"
echo "Agora: acesse $API_URL"