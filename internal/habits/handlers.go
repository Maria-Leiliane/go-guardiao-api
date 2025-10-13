package habits

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"go-guardiao-api/internal/auth"
	"go-guardiao-api/pkg/models"

	"github.com/gorilla/mux"
)

// HandleCreateHabit cria um novo hábito/meta para o usuário.
func HandleCreateHabit(w http.ResponseWriter, r *http.Request) {
	// 1. Obter UserID do contexto
	userID, err := auth.GetUserIDFromContext(r)
	if err != nil {
		http.Error(w, "Acesso negado.", http.StatusUnauthorized)
		return
	}

	var newHabit models.Habit
	if err := json.NewDecoder(r.Body).Decode(&newHabit); err != nil {
		http.Error(w, "Requisição inválida.", http.StatusBadRequest)
		return
	}

	// Lógica real: Salvar o newHabit no RDS.

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(map[string]string{
		"message":  fmt.Sprintf("Hábito '%s' criado com sucesso.", newHabit.Name),
		"habit_id": "mock-habit-123", // ID retornado após salvar no DB
		"user_id":  userID,
	}); err != nil {
		http.Error(w, "Erro ao serializar resposta.", http.StatusInternalServerError)
	}
}

// HandleLogHabit registra o progresso diário de um hábito.
// Esta ação é crítica, pois irá alimentar o Serviço de Gamificação.
func HandleLogHabit(w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.GetUserIDFromContext(r)
	var logData models.HabitLog

	if err := json.NewDecoder(r.Body).Decode(&logData); err != nil {
		http.Error(w, "Requisição inválida.", http.StatusBadRequest)
		return
	}

	// Lógica real: Salvar o logData no DynamoDB (para alto volume).
	// Em um sistema real, essa ação acionaria uma mensagem SQS
	// para o Worker de Gamificação calcular a Mana.

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(map[string]string{
		"message": fmt.Sprintf("Progresso registrado para o hábito %s. Mana em processamento.", logData.HabitID),
		"user_id": userID,
	}); err != nil {
		http.Error(w, "Erro ao serializar resposta.", http.StatusInternalServerError)
	}
}

// HandleGetHabits busca todos os hábitos e metas do usuário.
func HandleGetHabits(w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.GetUserIDFromContext(r)

	// Lógica real: Consultar lista de hábitos no RDS.

	// Mock: Retorna uma lista de hábitos.
	habits := []models.Habit{
		{ID: "h1", UserID: userID, Name: "Beber Água (2L)", GoalType: "Hydration", Frequency: "Daily"},
		{ID: "h2", UserID: userID, Name: "Tomar Medicação", GoalType: "Medication", Frequency: "Daily"},
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(habits); err != nil {
		http.Error(w, "Erro ao serializar resposta.", http.StatusInternalServerError)
	}
}

// HandleGetHabitLogs busca o histórico de logs para um hábito específico.
func HandleGetHabitLogs(w http.ResponseWriter, r *http.Request) {
	// Obtém o HabitID da URL.
	vars := mux.Vars(r)
	habitID := vars["habitId"]

	// Mock: Retorna logs de exemplo
	logs := []models.HabitLog{
		{HabitID: habitID, Value: 1, Timestamp: time.Now().Add(-24 * time.Hour)},
		{HabitID: habitID, Value: 1, Timestamp: time.Now()},
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(logs); err != nil {
		http.Error(w, "Erro ao serializar resposta.", http.StatusInternalServerError)
	}
}
