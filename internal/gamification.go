package auth

import (
	"encoding/json"
	"net/http"

	"go-guardiao-api/internal/auth"
	"go-guardiao-api/pkg/models"
)

// HandleGetManaBalance consulta o saldo de Mana do usuário logado.
// Idealmente, consulta o ElastiCache (Redis) para performance, com fallback para o RDS.
func HandleGetManaBalance(w http.ResponseWriter, r *http.Request) {
	userID, err := auth.GetUserIDFromContext(r)
	if err != nil {
		http.Error(w, "Acesso negado.", http.StatusUnauthorized)
		return
	}

	// Lógica real: Consultar ElastiCache ou RDS.

	// Mock: Retorna um saldo de Mana
	manaBalance := models.UserMana{
		UserID:  userID,
		Balance: 1850,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(manaBalance)
}

// HandleRedeemReward permite ao usuário gastar Mana por um prêmio.
// Esta é uma transação crítica que deve usar o RDS para garantir forte consistência.
func HandleRedeemReward(w http.ResponseWriter, r *http.Request) {
	// Lógica real:
	// 1. Verificar saldo de Mana (RDS)
	// 2. Se OK, subtrair Mana e registrar models.ManaTransaction (Transação no RDS)
	// 3. Registrar prêmio resgatado em outra tabela.

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Prêmio 'Medalha de Diamante' resgatado com sucesso! Saldo atualizado.",
	})
}
