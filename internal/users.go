package auth

import (
	"/auth.go"
	"encoding/json"
	"go-guardiao-api/pkg"
	"net/http"
)

// HandleGetUserProfile busca e retorna o perfil do usuário logado.
func HandleGetUserProfile(w http.ResponseWriter, r *http.Request) {
	// 1. Obter UserID do contexto da requisição (inserido pelo JWTAuthMiddleware)
	userID, err := auth.GetUserIDFromContext(r)
	if err != nil {
		http.Error(w, "Acesso negado: UserID ausente.", http.StatusUnauthorized)
		return
	}

	// Lógica real: Consultar o RDS (PostgreSQL) usando o userID

	// Mock: Retorna um perfil de exemplo
	userProfile := models.User{
		ID:    userID,
		Email: "usuario_logado@email.com",
		Name:  "Maria da Silva",
		Theme: "OutubroRosa",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(userProfile)
}

// HandleUpdateProfile permite ao usuário atualizar seus dados.
func HandleUpdateProfile(w http.ResponseWriter, r *http.Request) {
	// Lógica real: Deserializar payload, validar, atualizar no RDS.
	userID, _ := auth.GetUserIDFromContext(r)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Perfil atualizado com sucesso",
		"user_id": userID,
	})
}

// HandleAddSupportContact adiciona um novo contato à rede de apoio.
func HandleAddSupportContact(w http.ResponseWriter, r *http.Request) {
	// Lógica real: Deserializar payload (models.SupportContact), salvar no RDS.
	userID, _ := auth.GetUserIDFromContext(r)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Contato de apoio adicionado",
		"user_id": userID,
	})
}
