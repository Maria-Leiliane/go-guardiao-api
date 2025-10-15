#!/bin/bash

set -e

PORT=8080
API_URL="http://localhost:$PORT/api/v1/auth/login"

echo "Construindo e iniciando a arquitetura Guardião da Saúde..."

# Detecta se precisa de sudo
DOCKER_CMD="docker"
if ! docker info >/dev/null 2>&1; then
  DOCKER_CMD="sudo docker"
fi

echo "Passo 1: Build e subida dos containers..."
$DOCKER_CMD compose up --build -d

echo "----------------------------------------------------"
echo "Arquitetura Go Iniciada:"
echo "API (Síncrona): $API_URL"
echo "Worker (Assíncrono): Monitorando logs para cálculo de Mana"
echo "----------------------------------------------------"
echo "Para ver os logs: $DOCKER_CMD compose logs -f"
echo "Para parar: $DOCKER_CMD compose down"
echo "----------------------------------------------------"
echo "Agora: acesse $API_URL"