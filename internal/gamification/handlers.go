package gamification

import (
	"encoding/json"
	"fmt"
	"log" // Adicionado para logging de cache
	"net/http"

	"go-guardiao-api/internal/auth"
	"go-guardiao-api/internal/platforms/cache"
	"go-guardiao-api/internal/platforms/db"
	"go-guardiao-api/pkg/models"
)

// Service representa o serviço de Gamificação. Contém DB e Cache.
type Service struct {
	DBClient    *db.Client
	CacheClient *cache.Client
}

// NewService cria uma nova instância do serviço, injetando DB e Cache.
func NewService(dbClient *db.Client, cacheClient *cache.Client) *Service {
	return &Service{
		DBClient:    dbClient,
		CacheClient: cacheClient,
	}
}

// HandleGetManaBalance consulta o saldo de Mana, priorizando o Cache.
func (s *Service) HandleGetManaBalance(w http.ResponseWriter, r *http.Request) {
	userID, err := auth.GetUserIDFromContext(r)
	if err != nil {
		http.Error(w, "Acesso negado: UserID ausente.", http.StatusUnauthorized)
		return
	}

	var balance int

	// 1. Tenta buscar no Cache
	balance, err = s.CacheClient.GetManaBalance(r.Context(), userID)

	if err != nil {
		// Se o cache falhar (cache miss ou Redis inoperante):
		log.Printf("INFO: Cache Miss ou Redis inoperante para %s. Buscando no DB.", userID)

		// Busca no DB
		balance, err = s.DBClient.GetManaBalance(r.Context(), userID)
		if err != nil {
			http.Error(w, fmt.Sprintf("Falha ao consultar o saldo de Mana no DB: %v", err), http.StatusInternalServerError)
			return
		}

		// Atualiza o Cache (deve ser feito de forma não crítica)
		if setErr := s.CacheClient.SetManaBalance(r.Context(), userID, balance); setErr != nil {
			log.Printf("AVISO: Falha ao atualizar cache após DB lookup: %v", setErr)
			// Não retornamos, permitindo que a resposta continue
		}
	}

	manaBalance := models.UserMana{
		UserID:  userID,
		Balance: balance,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(manaBalance); err != nil {
		http.Error(w, "Erro ao serializar resposta.", http.StatusInternalServerError)
	}
}

// HandleRedeemReward utiliza o DB para transação crítica e invalida o Cache.
func (s *Service) HandleRedeemReward(w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.GetUserIDFromContext(r)

	var redemptionRequest struct {
		RewardID string `json:"reward_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&redemptionRequest); err != nil {
		http.Error(w, "Requisição inválida (RewardID ausente).", http.StatusBadRequest)
		return
	}

	// 1. Executa a transação crítica (débito de saldo e registro de transação)
	tx := models.ManaTransaction{
		UserID:      userID,
		Type:        "REWARD_REDEEM",
		Amount:      -500, // Custo mock da recompensa
		ReferenceID: redemptionRequest.RewardID,
	}

	if err := s.DBClient.CreateManaTransaction(r.Context(), tx); err != nil {
		http.Error(w, fmt.Sprintf("Falha na transação de resgate: %v", err), http.StatusInternalServerError)
		return
	}

	// 2. Após sucesso, obtém o novo saldo e atualiza o Cache e o Leaderboard
	newBalance, _ := s.DBClient.GetManaBalance(r.Context(), userID)

	// ATUALIZAÇÃO DE CACHE (Não Crítica)
	if setErr := s.CacheClient.SetManaBalance(r.Context(), userID, newBalance); setErr != nil {
		log.Printf("AVISO: Falha ao atualizar cache de Mana: %v", setErr)
	}
	if lbErr := s.CacheClient.UpdateLeaderboard(r.Context(), userID, newBalance); lbErr != nil {
		log.Printf("AVISO: Falha ao atualizar leaderboard: %v", lbErr)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(map[string]string{
		"message":     fmt.Sprintf("Recompensa %s resgatada com sucesso! Mana debitada.", redemptionRequest.RewardID),
		"user_id":     userID,
		"new_balance": fmt.Sprintf("%d", newBalance),
	}); err != nil {
		http.Error(w, "Erro ao serializar resposta.", http.StatusInternalServerError)
	}
}

// HandleGetLeaderboard consulta o Placar de Líderes, priorizando o Cache.
func (s *Service) HandleGetLeaderboard(w http.ResponseWriter, r *http.Request) {
	_, err := auth.GetUserIDFromContext(r)
	if err != nil {
		http.Error(w, "Acesso negado: UserID ausente.", http.StatusUnauthorized)
		return
	}

	// Tenta buscar no Cache
	leaderboard, err := s.CacheClient.GetLeaderboard(r.Context(), 10)
	if err == nil && len(leaderboard) > 0 {
		// Cache Hit
	} else {
		// Cache Miss: Usamos um mock, mas a lógica real buscaria no DB ou construiria.
		leaderboard = []models.LeaderboardEntry{
			{UserID: "u1", UserName: "Campeão", Mana: 5000},
			{UserID: "u2", UserName: "Aspirante", Mana: 3200},
		}
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(leaderboard); err != nil {
		http.Error(w, "Erro ao serializar resposta.", http.StatusInternalServerError)
	}
}

// HandleListChallenges (Não usa cache)
func (s *Service) HandleListChallenges(w http.ResponseWriter, r *http.Request) {
	_, err := auth.GetUserIDFromContext(r)
	if err != nil {
		http.Error(w, "Acesso negado: UserID ausente.", http.StatusUnauthorized)
		return
	}

	// Mock: Retorna desafios de exemplo (normalmente viria do DB)
	challenges := []models.Challenge{
		{ID: "c1", Name: "Caminhada Semanal", GoalType: "STEPS", ManaReward: 100},
		{ID: "c2", Name: "Log de Humor Diário", GoalType: "LOG_ENTRY", ManaReward: 50},
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(challenges); err != nil {
		http.Error(w, "Erro ao serializar resposta.", http.StatusInternalServerError)
	}
}
