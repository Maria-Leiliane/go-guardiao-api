package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"time"

	"go-guardiao-api/internal/platforms/db"
	"go-guardiao-api/pkg/models"
)

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
	// Exemplo de regra: se for o hábito 'h1' (ex: Beber Água) e o valor for 1, ganha Mana.
	if logData.Value >= 1 && logData.HabitID == "h1" {
		manaGained = 25
	} else if logData.Value >= 30 && logData.HabitID == "h2" { // Ex: Hábito de 'Caminhar' (30 minutos)
		manaGained = 50
	} else {
		log.Printf("INFO: Hábito %s não atendeu aos critérios para Mana.", logData.HabitID)
		return nil // Nenhuma Mana gerada, mas o processamento foi bem-sucedido
	}

	log.Printf("CALCULADO: Usuário %s ganhou %d Mana.", logData.UserID, manaGained)

	// 3. Persistência: Criar Transação de Mana (usando a função transacional)
	tx := models.ManaTransaction{
		UserID:      logData.UserID,
		Type:        "HABIT_COMPLETION",
		Amount:      manaGained,
		ReferenceID: logData.HabitID,
		CreatedAt:   time.Now(),
	}

	if err := dbClient.CreateManaTransaction(ctx, tx); err != nil {
		log.Printf("ERRO CRÍTICO DB: Falha ao registrar transação de Mana: %v", err)
		return err // Sinaliza falha para reprocessamento (simulando a visibilidade do SQS)
	}

	log.Printf("SUCESSO: Transação de Mana registrada e saldo atualizado para %s.", logData.UserID)

	// 4. Aqui, a lógica real enviaria uma notificação push via AWS SNS para o usuário!

	return nil
}

// simulateConsumption simula o loop infinito de consumo de fila (SQS Long Polling).
func simulateConsumption(dbClient *db.Client) {
	log.Println("Worker iniciado. Simulando consumo de fila SQS...")

	// Mock de mensagens que viriam da API (via SQS)
	mockMessages := []string{
		`{"habit_id": "h1", "user_id": "mock-user-456", "value": 1}`,  // Gera 25 Mana
		`{"habit_id": "h2", "user_id": "mock-user-456", "value": 30}`, // Gera 50 Mana
		`{"habit_id": "h2", "user_id": "mock-user-456", "value": 10}`, // Não gera Mana (abaixo de 30)
	}
	messageIndex := 0

	for {
		time.Sleep(3 * time.Second) // Simula o intervalo de polling

		if messageIndex < len(mockMessages) {
			payload := []byte(mockMessages[messageIndex])

			log.Println("WORKER: Mensagem mock recebida. Processando...")

			// Executa a lógica de processamento
			err := WorkerProcessar(context.Background(), dbClient, payload)

			if err != nil {
				log.Printf("WORKER: Falha no processamento. Manter mensagem na fila (Erro: %v)", err)
			} else {
				log.Println("WORKER: Processamento concluído. Mensagem removida da fila.")
			}

			messageIndex++
		} else {
			log.Println("WORKER: Fila de mensagens mockadas esgotada. Rodando em loop infinito (esperando SQS).")
			time.Sleep(10 * time.Second)
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

	// 2. Inicia o loop de consumo (assíncrono)
	simulateConsumption(dbClient)
}
