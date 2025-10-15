package gamification

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"go-guardiao-api/internal/auth"
	"go-guardiao-api/internal/platforms/cache"
	"go-guardiao-api/internal/platforms/db"
	"go-guardiao-api/pkg/models"
)

// Service representa o serviço de Gamificação.
type Service struct {
	DBClient    *db.Client
	CacheClient *cache.Client
}

func NewService(dbClient *db.Client, cacheClient *cache.Client) *Service {
	return &Service{
		DBClient:    dbClient,
		CacheClient: cacheClient,
	}
}

// Helpers para respostas padronizadas
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}
func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

// HandleGetManaBalance consulta o saldo de Mana, priorizando o Cache.
func (s *Service) HandleGetManaBalance(w http.ResponseWriter, r *http.Request) {
	userID, err := auth.GetUserIDFromContext(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "Acesso negado: UserID ausente.")
		return
	}

	// 1. Tenta buscar no Cache
	balance, err := s.CacheClient.GetManaBalance(r.Context(), userID)
	if err != nil {
		log.Printf("INFO: Cache Miss ou Redis inoperante para %s. Buscando no DB.", userID)
		// Busca no DB
		balance, err = s.DBClient.GetManaBalance(r.Context(), userID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, fmt.Sprintf("Falha ao consultar o saldo de Mana no DB: %v", err))
			return
		}
		// Atualiza o Cache (não crítico)
		if setErr := s.CacheClient.SetManaBalance(r.Context(), userID, balance); setErr != nil {
			log.Printf("AVISO: Falha ao atualizar cache após DB lookup: %v", setErr)
		}
	}

	manaBalance := models.UserMana{
		UserID:  userID,
		Balance: balance,
	}
	writeJSON(w, http.StatusOK, manaBalance)
}

// HandleRedeemReward utiliza o DB para transação crítica e invalida o Cache/Leaderboard.
func (s *Service) HandleRedeemReward(w http.ResponseWriter, r *http.Request) {
	userID, err := auth.GetUserIDFromContext(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "Acesso negado: UserID ausente.")
		return
	}

	var redemptionRequest struct {
		RewardID string `json:"reward_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&redemptionRequest); err != nil {
		writeError(w, http.StatusBadRequest, "Requisição inválida (RewardID ausente).")
		return
	}

	tx := models.ManaTransaction{
		UserID:      userID,
		Type:        "REWARD_REDEEM",
		Amount:      -500, // Custo mock
		ReferenceID: redemptionRequest.RewardID,
	}
	if err := s.DBClient.CreateManaTransaction(r.Context(), tx); err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("Falha na transação de resgate: %v", err))
		return
	}

	newBalance, err := s.DBClient.GetManaBalance(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Falha ao buscar novo saldo após resgate.")
		return
	}

	// Atualiza cache/leaderboard (não críticos)
	if setErr := s.CacheClient.SetManaBalance(r.Context(), userID, newBalance); setErr != nil {
		log.Printf("AVISO: Falha ao atualizar cache de Mana: %v", setErr)
	}
	if lbErr := s.CacheClient.UpdateLeaderboard(r.Context(), userID, newBalance); lbErr != nil {
		log.Printf("AVISO: Falha ao atualizar leaderboard: %v", lbErr)
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"message":     fmt.Sprintf("Recompensa %s resgatada com sucesso! Mana debitada.", redemptionRequest.RewardID),
		"user_id":     userID,
		"new_balance": fmt.Sprintf("%d", newBalance),
	})
}

// HandleGetLeaderboard consulta o Placar de Líderes, priorizando o Cache e atualiza em lote com TTL.
func (s *Service) HandleGetLeaderboard(w http.ResponseWriter, r *http.Request) {
	_, err := auth.GetUserIDFromContext(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "Acesso negado: UserID ausente.")
		return
	}

	limit := 10
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	leaderboard, err := s.CacheClient.GetLeaderboard(r.Context(), int64(limit))
	if err == nil && len(leaderboard) > 0 {
		writeJSON(w, http.StatusOK, leaderboard)
		return
	}

	leaderboard, err = s.DBClient.GetTopManaUsers(r.Context(), limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("Falha ao buscar leaderboard: %v", err))
		return
	}

	// Atualização em lote + TTL (exemplo: 2 minutos = 120 segundos)
	batchErr := s.CacheClient.UpdateLeaderboardBatch(r.Context(), leaderboard, 120)
	if batchErr != nil {
		log.Printf("AVISO: Falha ao atualizar leaderboard em lote no cache: %v", batchErr)
	}

	writeJSON(w, http.StatusOK, leaderboard)
}

// HandleListChallenges retorna desafios (mock, sem cache).
func (s *Service) HandleListChallenges(w http.ResponseWriter, r *http.Request) {
	_, err := auth.GetUserIDFromContext(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "Acesso negado: UserID ausente.")
		return
	}

	challenges := []models.Challenge{
		{ID: "c1", Name: "Caminhada Semanal", GoalType: "STEPS", ManaReward: 100},
		{ID: "c2", Name: "Log de Humor Diário", GoalType: "LOG_ENTRY", ManaReward: 50},
	}
	writeJSON(w, http.StatusOK, challenges)
}
