package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"strconv"
	"time"

	"go-guardiao-api/internal/platforms/db"
	"go-guardiao-api/pkg/models"
)

const (
	defaultPollSeconds = 3
	defaultIdleSeconds = 10
)

// getEnv busca variável de ambiente com fallback
func getEnv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

// getEnvInt busca variável de ambiente inteira com fallback
func getEnvInt(k string, def int) int {
	v := getEnv(k, "")
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil || n <= 0 {
		return def
	}
	return n
}

// WorkerProcessar simula o recebimento de uma mensagem do SQS e executa a lógica de cálculo.
func WorkerProcessar(ctx context.Context, dbClient *db.Client, messagePayload []byte) error {
	var logData models.HabitLog

	// 1. Deserializa a mensagem (que seria o log do hábito da API)
	if err := json.Unmarshal(messagePayload, &logData); err != nil {
		log.Printf("ERRO: Falha ao desserializar payload: %v. Payload: %s", err, string(messagePayload))
		// No SQS real, esta mensagem seria movida para uma Dead Letter Queue.
		return err
	}

	// 2. Lógica de Negócio: Calcular Mana
	manaGained := 0
	switch {
	case logData.Value >= 1 && logData.HabitID == "h1":
		// Ex: Beber Água
		manaGained = 25
	case logData.Value >= 30 && logData.HabitID == "h2":
		// Ex: Caminhar 30 minutos
		manaGained = 50
	default:
		log.Printf("INFO: Hábito %s não atendeu aos critérios para Mana.", logData.HabitID)
		return nil // Nenhuma Mana gerada, mas o processamento foi bem-sucedido
	}

	log.Printf("CALCULADO: Usuário %s ganhou %d Mana (habit=%s value=%d).", logData.UserID, manaGained, logData.HabitID, logData.Value)

	// 3. Persistência: Criar Transação de Mana (usando a função transacional)
	tx := models.ManaTransaction{
		UserID:      logData.UserID,
		Type:        "HABIT_COMPLETION",
		Amount:      manaGained,
		ReferenceID: logData.HabitID,
		CreatedAt:   time.Now(),
	}

	// Contexto com timeout para operação no DB
	dbCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := dbClient.CreateManaTransaction(dbCtx, tx); err != nil {
		log.Printf("ERRO CRÍTICO DB: Falha ao registrar transação de Mana: %v", err)
		return err // Sinaliza falha para reprocessamento
	}

	log.Printf("SUCESSO: Transação de Mana registrada e saldo atualizado para %s.", logData.UserID)
	// 4. Aqui, a lógica real enviaria uma notificação push via AWS SNS para o usuário!
	return nil
}

func simulateConsumption(ctx context.Context, dbClient *db.Client) {
	log.Println("Worker iniciado. Simulando consumo de fila SQS...")

	enableMock := getEnv("WORKER_ENABLE_MOCK", "true") == "true"
	pollSec := getEnvInt("WORKER_POLL_INTERVAL_SECONDS", defaultPollSeconds)
	idleSec := getEnvInt("WORKER_IDLE_INTERVAL_SECONDS", defaultIdleSeconds)

	if !enableMock {
		log.Println("WORKER: Modo mock desabilitado. Aguardando integração real com SQS...")
		<-ctx.Done()
		return
	}

	// Mock de mensagens que viriam da API (via SQS)
	mockMessages := []string{
		`{"habit_id": "h1", "user_id": "mock-user-456", "value": 1}`,  // Gera 25 Mana
		`{"habit_id": "h2", "user_id": "mock-user-456", "value": 30}`, // Gera 50 Mana
		`{"habit_id": "h2", "user_id": "mock-user-456", "value": 10}`, // Não gera Mana (abaixo de 30)
	}
	messageIndex := 0

	pollTicker := time.NewTicker(time.Duration(pollSec) * time.Second)
	defer pollTicker.Stop()

	idleTicker := time.NewTicker(time.Duration(idleSec) * time.Second)
	defer idleTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("WORKER: Contexto cancelado. Encerrando consumo.")
			return

		case <-pollTicker.C:
			if messageIndex < len(mockMessages) {
				payload := []byte(mockMessages[messageIndex])
				log.Println("WORKER: Mensagem mock recebida. Processando...")

				if err := WorkerProcessar(ctx, dbClient, payload); err != nil {
					log.Printf("WORKER: Falha no processamento. Manter mensagem na fila (Erro: %v)", err)
				} else {
					log.Println("WORKER: Processamento concluído. Mensagem removida da fila.")
				}
				messageIndex++
			}

		case <-idleTicker.C:
			if messageIndex >= len(mockMessages) {
				log.Println("WORKER: Fila mock esgotada. Aguardando novas mensagens (simulando SQS).")
			}
		}
	}
}

func main() {
	log.Println("Iniciando Worker de Gamificação...")

	// 1. Inicializa o Cliente DB (necessário para persistir a Mana)
	dbClient, err := db.NewDBClient(os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("ERRO CRÍTICO: Falha ao inicializar o cliente DB para o worker: %v", err)
	}
	defer dbClient.Close()

	// 2. Contexto com shutdown gracioso
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	// 3. Inicia o loop de consumo (assíncrono)
	simulateConsumption(ctx, dbClient)

	log.Println("Worker encerrado com segurança.")
}
