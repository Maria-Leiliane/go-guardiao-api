package users

import (
	"encoding/json"
	"fmt"
	"net/http"

	"go-guardiao-api/internal/auth"
	"go-guardiao-api/pkg/models"
)

// HandleGetUserProfile busca e retorna o perfil do usuário logado.
func HandleGetUserProfile(w http.ResponseWriter, r *http.Request) {
	// 1. Obter UserID do contexto (inserido pelo JWTAuthMiddleware)
	userID, err := auth.GetUserIDFromContext(r)
	if err != nil {
		http.Error(w, "Acesso negado: UserID ausente.", http.StatusUnauthorized)
		return
	}

	// Mock: Retorna um perfil de exemplo
	userProfile := models.User{
		ID:    userID,
		Email: "usuario_logado@email.com",
		Name:  "Maria da Silva",
		Theme: "OutubroRosa",
	}

	w.Header().Set("Content-Type", "application/json")
	// Tratamento de erro de encoding adicionado
	if err := json.NewEncoder(w).Encode(userProfile); err != nil {
		http.Error(w, "Erro ao serializar resposta.", http.StatusInternalServerError)
	}
}

// HandleUpdateProfile permite ao usuário atualizar seus dados.
func HandleUpdateProfile(w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.GetUserIDFromContext(r)
	var updateData models.User

	// Tratamento de erro de decoding adicionado
	if err := json.NewDecoder(r.Body).Decode(&updateData); err != nil {
		http.Error(w, "Requisição inválida.", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(map[string]string{
		"message": "Perfil atualizado com sucesso",
		"user_id": userID,
		"name":    updateData.Name,
	}); err != nil {
		http.Error(w, "Erro ao serializar resposta.", http.StatusInternalServerError)
	}
}

// HandleAddSupportContact adiciona um novo contato à rede de apoio.
func HandleAddSupportContact(w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.GetUserIDFromContext(r)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(map[string]string{
		"message": "Contato de apoio adicionado",
		"user_id": userID,
	}); err != nil {
		http.Error(w, "Erro ao serializar resposta.", http.StatusInternalServerError)
	}
}

// HandleGetSupportContacts busca todos os contatos da rede de apoio.
func HandleGetSupportContacts(w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.GetUserIDFromContext(r)

	// Mock: Retorna lista de contatos (usando a struct corrigida)
	contacts := []models.SupportContact{
		{ContactID: "c1", UserID: userID, ContactEmail: "joao@email.com", Nickname: "João"},
		{ContactID: "c2", UserID: userID, ContactEmail: "ana@email.com", Nickname: "Ana"},
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(contacts); err != nil {
		http.Error(w, "Erro ao serializar resposta.", http.StatusInternalServerError)
	}
}

// HandleDeleteSupportContact remove um contato da rede de apoio.
func HandleDeleteSupportContact(w http.ResponseWriter, r *http.Request) {
	// O mux.Vars será usado no main.go para roteamento, mas não neste handler.
	// Simplesmente indica que a operação foi bem-sucedida.

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(map[string]string{
		"message": fmt.Sprintf("Contato removido com sucesso."),
	}); err != nil {
		http.Error(w, "Erro ao serializar resposta.", http.StatusInternalServerError)
	}
}
