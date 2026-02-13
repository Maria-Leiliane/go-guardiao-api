package users

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"

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

// ===== Helpers JSON =====

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

// ===== Handlers Perfil =====

// GET /user/profile — retorna o perfil do usuário autenticado
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

// PUT /user/profile — atualiza nome e/ou theme (avatar)
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
	})
}

// ===== NOVOS: Email e Senha =====

var emailRx = regexp.MustCompile(`^[^@\s]+@[^@\s]+\.[^@\s]+$`)

// PUT /user/email { "email": "novo@exemplo.com" }
func (s *Service) HandleUpdateEmail(w http.ResponseWriter, r *http.Request) {
	userID, err := auth.GetUserIDFromContext(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "Acesso negado: UserID ausente.")
		return
	}

	var payload struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, "Requisição inválida.")
		return
	}

	email := strings.TrimSpace(payload.Email)
	if email == "" || !emailRx.MatchString(email) {
		writeError(w, http.StatusBadRequest, "E-mail inválido.")
		return
	}

	if err := s.DBClient.UpdateUserEmail(r.Context(), userID, email); errors.Is(err, pgx.ErrNoRows) {
		writeError(w, http.StatusNotFound, "Usuário não encontrado para atualizar e-mail.")
		return
	} else if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("Falha ao atualizar e-mail: %v", err))
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"message": "E-mail atualizado com sucesso.",
		"user_id": userID,
	})
}

// PUT /user/password { "current":"Senha@123", "next":"NovaSenha@123" }
func (s *Service) HandleUpdatePassword(w http.ResponseWriter, r *http.Request) {
	userID, err := auth.GetUserIDFromContext(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "Acesso negado: UserID ausente.")
		return
	}

	var payload struct {
		Current string `json:"current"`
		Next    string `json:"next"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, "Requisição inválida.")
		return
	}

	if err := auth.ValidatePassword(payload.Next); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Confere senha atual
	hash, err := s.DBClient.GetUserPasswordHash(r.Context(), userID)
	if errors.Is(err, pgx.ErrNoRows) {
		writeError(w, http.StatusNotFound, "Usuário não encontrado.")
		return
	} else if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("Falha ao buscar senha: %v", err))
		return
	}

	if err := auth.CheckPassword(payload.Current, hash); err != nil {
		writeError(w, http.StatusUnauthorized, "Senha atual incorreta.")
		return
	}

	// Grava nova senha
	newHash, err := auth.HashPassword(payload.Next)
	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("Falha ao gerar hash: %v", err))
		return
	}
	if err := s.DBClient.SetUserPassword(r.Context(), userID, newHash); err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("Falha ao atualizar senha: %v", err))
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"message": "Senha alterada com sucesso.",
		"user_id": userID,
	})
}

// ===== Support Contacts =====

// POST /user/support-contact
func (s *Service) HandleAddSupportContact(w http.ResponseWriter, r *http.Request) {
	userID, err := auth.GetUserIDFromContext(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "Acesso negado: UserID ausente.")
		return
	}

	var payload struct {
		Name     string `json:"name"`     // obrigatório
		Phone    string `json:"phone"`    // opcional
		Email    string `json:"email"`    // opcional (caso queira notificar por email)
		Relation string `json:"relation"` // opcional (ex.: mãe, amiga, médico)
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, "Requisição inválida.")
		return
	}
	if strings.TrimSpace(payload.Name) == "" {
		writeError(w, http.StatusBadRequest, "Nome é obrigatório.")
		return
	}

	contact := models.SupportContact{
		UserID:                 userID,
		ContactEmail:           strings.TrimSpace(payload.Email),
		Phone:                  strings.TrimSpace(payload.Phone),
		Nickname:               strings.TrimSpace(payload.Name),
		NotificationPreference: strings.TrimSpace(payload.Relation),
		Name:                   strings.TrimSpace(payload.Name),
	}

	if err := s.DBClient.CreateSupportContact(r.Context(), contact); err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("Falha ao adicionar contato: %v", err))
		return
	}

	writeJSON(w, http.StatusCreated, map[string]string{
		"message": "Contato de apoio adicionado.",
		"user_id": userID,
	})
}

// GET /user/support-contact — (id, name, phone, relation)
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

	type Out struct {
		ID       string `json:"id"`
		Name     string `json:"name"`
		Phone    string `json:"phone,omitempty"`
		Relation string `json:"relation,omitempty"`
	}
	out := make([]Out, 0, len(contacts))
	for _, c := range contacts {
		out = append(out, Out{
			ID:       c.ContactID,
			Name:     firstNonEmpty(c.Nickname, c.Name),
			Phone:    c.Phone,
			Relation: c.NotificationPreference,
		})
	}
	writeJSON(w, http.StatusOK, out)
}

func firstNonEmpty(a, b string) string {
	if strings.TrimSpace(a) != "" {
		return a
	}
	if strings.TrimSpace(b) != "" {
		return b
	}
	return ""
}

// DELETE /user/support-contact/{contactId}
func (s *Service) HandleDeleteSupportContact(w http.ResponseWriter, r *http.Request) {
	userID, err := auth.GetUserIDFromContext(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "Acesso negado: UserID ausente.")
		return
	}

	vars := mux.Vars(r)
	contactID := strings.TrimSpace(vars["contactId"])
	if contactID == "" {
		writeError(w, http.StatusBadRequest, "contactId é obrigatório.")
		return
	}

	// Segurança: garante que o contato pertence ao usuário logado
	if err := s.DBClient.DeleteSupportContactByUser(r.Context(), userID, contactID); err != nil {
		// quando não encontrado, retorne 404
		if strings.Contains(strings.ToLower(err.Error()), "não encontrado") || strings.Contains(strings.ToLower(err.Error()), "no rows") {
			writeError(w, http.StatusNotFound, "Contato não encontrado.")
			return
		}
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("Falha ao remover contato: %v", err))
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "Contato removido."})
}
