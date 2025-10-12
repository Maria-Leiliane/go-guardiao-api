package auth

import (
	"encoding/json"
	"net/http"

	// Caminhos corrigidos para a sua estrutura atual
	"go-guardiao-api/internal/auth"
	"go-guardiao-api/pkg/models"
)

// HandleCreateHabit cria um novo hábito para o usuário logado.
func HandleCreateHabit(w http.ResponseWriter, r *http.Request) {
	// 1. Obter UserID do contexto
	userID, err := auth.GetUserIDFromContext(r)
	if err != nil {
		http.Error(w, "Acesso negado.", http.StatusUnauthorized)
		return
	}

	// Lógica real: Deserializar models.Habit, salvar no RDS.

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Hábito 'Tomar Medicamento' criado.",
		"user_id": userID,
	})
}

// HandleLogHabit registra um log de progresso para um hábito específico.
// Este log pode acionar o cálculo de Mana se for a conclusão de um desafio.
func HandleLogHabit(w http.ResponseWriter, r *http.Request) {
	// Lógica real: Deserializar models.HabitLog, salvar no DynamoDB (para logs de alto volume).

	// Nota: Em um projeto real, esta ação PODE enviar uma mensagem para uma fila SQS
	// para que o Serviço de Gamificação processe o potencial ganho de Mana assincronamente.

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Progresso de hábito registrado com sucesso. Mana em processamento (via SQS).",
	})
}

// HandleGetHabits busca todos os hábitos do usuário.
func HandleGetHabits(w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.GetUserIDFromContext(r)

	// Mock: Retorna uma lista de hábitos.
	habits := []models.Habit{
		{ID: "h1", UserID: userID, Name: "Beber Água", GoalType: "Hydration"},
		{ID: "h2", UserID: userID, Name: "Caminhada de 30min", GoalType: "Activity"},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(habits)
}
