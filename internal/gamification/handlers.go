package gamification

import (
	"encoding/json"
	"fmt"
	"net/http"

	"go-guardiao-api/internal/auth"
	"go-guardiao-api/pkg/models"
)

// HandleGetManaBalance consulta o saldo atual de Mana do usuário.
// No sistema real, consultara o ElastiCache (Redis) para performance.
func HandleGetManaBalance(w http.ResponseWriter, r *http.Request) {
	// 1. Obter UserID do contexto
	userID, err := auth.GetUserIDFromContext(r)
	if err != nil {
		http.Error(w, "Acesso negado.", http.StatusUnauthorized)
		return
	}

	// Lógica real: Consultar DB/Cache.
	// Mock: Retorna um saldo de Mana
	manaBalance := models.UserMana{
		UserID:  userID,
		Balance: 2500,
		// O campo UpdatedAt será preenchido pelo DB
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(manaBalance); err != nil {
		http.Error(w, "Erro ao serializar resposta.", http.StatusInternalServerError)
	}
}

// HandleRedeemReward permite ao usuário gastar Mana para resgatar um prêmio.
// Esta é uma transação crítica (dedução de saldo + registro de transação).
func HandleRedeemReward(w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.GetUserIDFromContext(r)

	// 1. Deserializar a recompensa que o usuário quer resgatar
	var redemptionRequest struct {
		RewardID string `json:"reward_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&redemptionRequest); err != nil {
		http.Error(w, "Requisição inválida (RewardID ausente).", http.StatusBadRequest)
		return
	}

	// Lógica real:
	// 1. Verificar custo da recompensa (DB).
	// 2. Tentar debitar o saldo do Mana do usuário (Transação RDS).
	// 3. Registrar ManaTransaction negativa.

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(map[string]string{
		"message": fmt.Sprintf("Recompensa %s resgatada com sucesso! Mana debitada.", redemptionRequest.RewardID),
		"user_id": userID,
	}); err != nil {
		http.Error(w, "Erro ao serializar resposta.", http.StatusInternalServerError)
	}
}

// HandleGetLeaderboard consulta o placar de líderes de Mana.
func HandleGetLeaderboard(w http.ResponseWriter, r *http.Request) {
	// Lógica real: Consultar o ElastiCache (Redis) para o placar de líderes.

	// Mock: Retorna um placar de líderes de exemplo
	leaderboard := []models.LeaderboardEntry{
		{UserID: "u1", UserName: "Campeão", Mana: 5000},
		{UserID: "u2", UserName: "Aspirante", Mana: 3200},
		{UserID: "u3", UserName: "Você", Mana: 2500},
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(leaderboard); err != nil {
		http.Error(w, "Erro ao serializar resposta.", http.StatusInternalServerError)
	}
}

// HandleListChallenges lista todos os desafios ativos.
func HandleListChallenges(w http.ResponseWriter, r *http.Request) {
	// Lógica real: Consultar desafios ativos no DB.

	// Mock: Retorna desafios de exemplo
	challenges := []models.Challenge{
		{ID: "c1", Name: "Caminhada Semanal", GoalType: "STEPS", ManaReward: 100},
		{ID: "c2", Name: "Log de Humor Diário", GoalType: "LOG_ENTRY", ManaReward: 50},
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(challenges); err != nil {
		http.Error(w, "Erro ao serializar resposta.", http.StatusInternalServerError)
	}
}
