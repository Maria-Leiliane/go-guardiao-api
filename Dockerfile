#1. Estágio de Build (Compilação)
FROM golang:1.24-alpine AS builder

WORKDIR /app

#Copia os arquivos de configuração do módulo
COPY go.mod .
COPY go.sum .

#Baixa as dependências e faz cache delas
RUN go mod download

#Copia o código fonte da aplicação
COPY . .

#Compila os binários finais para a API e o Worker
ENV CGO_ENABLED=0
ENV GOOS=linux

#Compila a API
RUN go build -o /api ./cmd/api/main.go

#Compila o Worker
RUN go build -o /worker ./cmd/gamification_worker/main.go

#2. Estágio Final (Imagem de Produção)
FROM alpine

WORKDIR /app

#Copia os binários compilados do estágio 'builder'
COPY --from=builder /api /app/api
COPY --from=builder /worker /app/worker

ENTRYPOINT ["/app/api"]
CMD [""]