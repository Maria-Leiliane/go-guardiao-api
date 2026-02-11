package auth

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5"

	"go-guardiao-api/internal/platforms/db"
	"go-guardiao-api/pkg/models"
)

type AuthPayload struct {
	Email    string `json:"email"`
	Name     string `json:"name,omitempty"`
	Password string `json:"password,omitempty"`
}

// Registro: cria o usuário com UUID (se não existir) e retorna token com esse UUID.
func HandleRegisterWithDB(w http.ResponseWriter, r *http.Request, dbClient *db.Client) {
	var p AuthPayload
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil || p.Email == "" {
		errorJSON(w, http.StatusBadRequest, "Requisição inválida: email obrigatório.")
		return
	}

	u, err := dbClient.GetUserByEmail(r.Context(), p.Email)
	if errors.Is(err, pgx.ErrNoRows) {
		u = models.User{
			Email: p.Email,
			Name:  p.Name,
			Theme: "",
		}
		if err = dbClient.CreateUser(r.Context(), u); err != nil {
			errorJSON(w, http.StatusInternalServerError, "Falha ao criar usuário")
			return
		}
		// Busca novamente para garantir campos atualizados, inclusive ID
		u, _ = dbClient.GetUserByEmail(r.Context(), p.Email)
	} else if err != nil {
		errorJSON(w, http.StatusInternalServerError, "Falha ao consultar usuário")
		return
	}

	token, err := GenerateToken(u.ID, u.Email, 24*time.Hour)
	if err != nil {
		errorJSON(w, http.StatusInternalServerError, "Falha ao gerar token")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"message": "Usuário registrado com sucesso",
		"token":   token,
		"user_id": u.ID,
	})
}

// Login: busca por email (auto-cria opcional) e retorna token com UUID.
func HandleLoginWithDB(w http.ResponseWriter, r *http.Request, dbClient *db.Client) {
	var p AuthPayload
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil || p.Email == "" {
		errorJSON(w, http.StatusBadRequest, "Requisição inválida: email obrigatório.")
		return
	}

	u, err := dbClient.GetUserByEmail(r.Context(), p.Email)
	if errors.Is(err, pgx.ErrNoRows) {
		// Fluxo simples: cria se não existir
		u = models.User{
			Email: p.Email,
			Name:  p.Name,
			Theme: "",
		}
		if err = dbClient.CreateUser(r.Context(), u); err != nil {
			errorJSON(w, http.StatusInternalServerError, "Falha ao criar usuário")
			return
		}
		u, _ = dbClient.GetUserByEmail(r.Context(), p.Email)
	} else if err != nil {
		errorJSON(w, http.StatusInternalServerError, "Falha ao consultar usuário")
		return
	}

	token, err := GenerateToken(u.ID, u.Email, 24*time.Hour)
	if err != nil {
		errorJSON(w, http.StatusInternalServerError, "Falha ao gerar token")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{
		"message": "Login realizado com sucesso",
		"token":   token,
		"user_id": u.ID,
	})
}
