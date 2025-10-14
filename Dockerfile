#1. Estágio de Build (Compilação)
FROM golang:1.21-alpine AS builder

WORKDIR /app

#Copia os arquivos de configuração do módulo
COPY go.mod .
COPY go.sum .

#Baixa as dependências e faz cache delas
RUN go mod download

Copia o código fonte da aplicação
COPY . .

#Compila os binários finais para a API e o Worker
#CGO_ENABLED=0 garante que o binário é estático e independente do sistema operacional
ENV CGO_ENABLED=0
ENV GOOS=linux

#Compila a API
RUN go build -o /api ./cmd/api/main.go

#Compila o Worker
RUN go build -o /worker ./cmd/gamification_worker/main.go

#2. Estágio Final (Imagem de Produção)
#Usamos 'scratch' para a imagem mais leve possível (apenas o binário)
FROM alpine

WORKDIR /app

#Copia os binários compilados do estágio 'builder'
COPY --from=builder /api /app/api
COPY --from=builder /worker /app/worker

#Define o ponto de entrada padrão como a API
ENTRYPOINT ["/app/api"]

#O comando padrão para executar a API (pode ser sobrescrito no docker-compose)
CMD [""]