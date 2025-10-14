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

// Service representa o serviço de Usuários. Ele contém a dependência do DB.
type Service struct {
	DBClient *db.Client
}

// NewService cria uma nova instância do serviço de Usuários, injetando o cliente DB.
func NewService(dbClient *db.Client) *Service {
	return &Service{DBClient: dbClient}
}

// HandleGetUserProfile busca e retorna o perfil do usuário logado usando o DB.
func (s *Service) HandleGetUserProfile(w http.ResponseWriter, r *http.Request) {
	// 1. Obter UserID do contexto
	userID, err := auth.GetUserIDFromContext(r)
	if err != nil {
		http.Error(w, "Acesso negado: UserID ausente.", http.StatusUnauthorized)
		return
	}

	// 2. Buscar no DB
	userProfile, err := s.DBClient.GetUserByID(r.Context(), userID) // <--- USANDO GetUserByID
	if errors.Is(err, pgx.ErrNoRows) {
		http.Error(w, "Perfil de usuário não encontrado.", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, fmt.Sprintf("Falha interna ao buscar perfil: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(userProfile); err != nil {
		http.Error(w, "Erro ao serializar resposta.", http.StatusInternalServerError)
	}
}

// HandleUpdateProfile permite ao usuário atualizar seus dados no DB.
func (s *Service) HandleUpdateProfile(w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.GetUserIDFromContext(r)
	var updateData models.User

	if err := json.NewDecoder(r.Body).Decode(&updateData); err != nil {
		http.Error(w, "Requisição inválida.", http.StatusBadRequest)
		return
	}

	// Preenche o ID e atualiza no DB
	updateData.ID = userID
	if err := s.DBClient.UpdateUser(r.Context(), updateData); errors.Is(err, pgx.ErrNoRows) { // <--- USANDO UpdateUser
		http.Error(w, "Perfil não encontrado para atualização.", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, fmt.Sprintf("Falha interna ao atualizar perfil: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(map[string]string{
		"message": "Perfil atualizado com sucesso.",
		"user_id": userID,
		"name":    updateData.Name,
	}); err != nil {
		http.Error(w, "Erro ao serializar resposta.", http.StatusInternalServerError)
	}
}

// HandleAddSupportContact adiciona um novo contato à rede de apoio no DB.
func (s *Service) HandleAddSupportContact(w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.GetUserIDFromContext(r)
	var contact models.SupportContact

	if err := json.NewDecoder(r.Body).Decode(&contact); err != nil {
		http.Error(w, "Requisição inválida.", http.StatusBadRequest)
		return
	}

	contact.UserID = userID
	if err := s.DBClient.CreateSupportContact(r.Context(), contact); err != nil { // <--- USANDO CreateSupportContact
		http.Error(w, fmt.Sprintf("Falha ao adicionar contato: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(map[string]string{
		"message": "Contato de apoio adicionado.",
		"user_id": userID,
	}); err != nil {
		http.Error(w, "Erro ao serializar resposta.", http.StatusInternalServerError)
	}
}

// HandleGetSupportContacts busca todos os contatos da rede de apoio no DB.
func (s *Service) HandleGetSupportContacts(w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.GetUserIDFromContext(r)

	contacts, err := s.DBClient.GetSupportContactsByUserID(r.Context(), userID) // <--- USANDO GetSupportContactsByUserID
	if err != nil {
		http.Error(w, fmt.Sprintf("Falha ao buscar contatos: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(contacts); err != nil {
		http.Error(w, "Erro ao serializar resposta.", http.StatusInternalServerError)
	}
}

// HandleDeleteSupportContact remove um contato da rede de apoio no DB.
func (s *Service) HandleDeleteSupportContact(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	contactId := vars["contactId"]

	if err := s.DBClient.DeleteSupportContact(r.Context(), contactId); errors.Is(err, pgx.ErrNoRows) { // <--- USANDO DeleteSupportContact
		http.Error(w, "Contato não encontrado.", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, fmt.Sprintf("Falha ao deletar contato: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(map[string]string{
		"message": fmt.Sprintf("Contato %s removido com sucesso.", contactId),
	}); err != nil {
		http.Error(w, "Erro ao serializar resposta.", http.StatusInternalServerError)
	}
}
