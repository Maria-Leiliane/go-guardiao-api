package users

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5"

	"go-guardiao-api/internal/auth"
	"go-guardiao-api/internal/platforms/db"
	"go-guardiao-api/pkg/models"
)

// Service representa o serviço de Usuários.
type Service struct {
	DBClient *db.Client
}

func NewService(dbClient *db.Client) *Service {
	return &Service{DBClient: dbClient}
}

// JSON helpers para padronizar respostas e erros
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

// HandleGetUserProfile retorna o perfil do usuário autenticado.
func (s *Service) HandleGetUserProfile(w http.ResponseWriter, r *http.Request) {
	userID, err := auth.GetUserIDFromContext(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "Acesso negado: UserID ausente.")
		return
	}

	userProfile, err := s.DBClient.GetUserByID(r.Context(), userID)
	if errors.Is(err, pgx.ErrNoRows) {
		writeError(w, http.StatusNotFound, "Perfil de usuário não encontrado.")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("Falha interna ao buscar perfil: %v", err))
		return
	}

	writeJSON(w, http.StatusOK, userProfile)
}

// HandleUpdateProfile permite ao usuário atualizar seus dados.
func (s *Service) HandleUpdateProfile(w http.ResponseWriter, r *http.Request) {
	userID, err := auth.GetUserIDFromContext(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "Acesso negado: UserID ausente.")
		return
	}

	var updateData models.User
	if err := json.NewDecoder(r.Body).Decode(&updateData); err != nil {
		writeError(w, http.StatusBadRequest, "Requisição inválida.")
		return
	}

	updateData.ID = userID
	err = s.DBClient.UpdateUser(r.Context(), updateData)
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		writeError(w, http.StatusNotFound, "Perfil não encontrado para atualização.")
		return
	case err != nil:
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("Falha interna ao atualizar perfil: %v", err))
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"message": "Perfil atualizado com sucesso.",
		"user_id": userID,
		"name":    updateData.Name,
	})
}

// HandleAddSupportContact adiciona um novo contato de apoio.
func (s *Service) HandleAddSupportContact(w http.ResponseWriter, r *http.Request) {
	userID, err := auth.GetUserIDFromContext(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "Acesso negado: UserID ausente.")
		return
	}

	var contact models.SupportContact
	if err := json.NewDecoder(r.Body).Decode(&contact); err != nil {
		writeError(w, http.StatusBadRequest, "Requisição inválida.")
		return
	}

	contact.UserID = userID
	if err := s.DBClient.CreateSupportContact(r.Context(), contact); err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("Falha ao adicionar contato: %v", err))
		return
	}

	writeJSON(w, http.StatusCreated, map[string]string{
		"message": "Contato de apoio adicionado.",
		"user_id": userID,
	})
}

// HandleGetSupportContacts retorna todos os contatos de apoio do usuário logado.
func (s *Service) HandleGetSupportContacts(w http.ResponseWriter, r *http.Request) {
	userID, err := auth.GetUserIDFromContext(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "Acesso negado: UserID ausente.")
		return
	}

	contacts, err := s.DBClient.GetSupportContactsByUserID(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("Falha ao buscar contatos: %v", err))
		return
	}

	writeJSON(w, http.StatusOK, contacts)
}

// HandleDeleteSupportContact remove um contato de apoio do usuário.
func (s *Service) HandleDeleteSupportContact(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	contactId := vars["contactId"]

	if contactId == "" {
		writeError(w, http.StatusBadRequest, "ID do contato ausente.")
		return
	}

	err := s.DBClient.DeleteSupportContact(r.Context(), contactId)
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		writeError(w, http.StatusNotFound, "Contato não encontrado.")
		return
	case err != nil:
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("Falha ao deletar contato: %v", err))
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"message": fmt.Sprintf("Contato %s removido com sucesso.", contactId),
	})
}
