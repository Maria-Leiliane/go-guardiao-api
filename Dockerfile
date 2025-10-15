# 1. Build stage
FROM golang:1.24-alpine AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ENV CGO_ENABLED=0 GOOS=linux

RUN go build -o /api ./cmd/api/main.go
RUN go build -o /worker ./cmd/gamification_worker/main.go

# 2. Run stage
FROM alpine:latest

RUN addgroup -S appgroup && adduser -S appuser -G appgroup
WORKDIR /app

COPY --from=builder /api /app/api
COPY --from=builder /worker /app/worker

RUN chown -R appuser:appgroup /app

USER appuser

ENTRYPOINT ["/app/api"]
CMD [""]