package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// contextKey define um tipo para evitar colisão de chaves no contexto.
type contextKey string

const (
	userIDKey contextKey = "userID"
	emailKey  contextKey = "userEmail"
)

// jwtKey busca a chave secreta de variável de ambiente.
var jwtKey = []byte(getEnv("JWT_SECRET", "sua-chave-secreta-muito-forte"))

// getEnv busca variável de ambiente ou retorna fallback.
func getEnv(k, fallback string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return fallback
}

// Claims define a estrutura dos dados que serão armazenados no token JWT.
type Claims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

// GetUserIDFromContext obtém o ID do usuário do contexto.
func GetUserIDFromContext(r *http.Request) (string, error) {
	userID, ok := r.Context().Value(userIDKey).(string)
	if !ok {
		return "", errors.New("user ID não encontrado no contexto")
	}
	return userID, nil
}

// GenerateToken cria e assina um novo token JWT para um usuário.
func GenerateToken(userID, email string, expiration time.Duration) (string, error) {
	claims := &Claims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiration)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtKey)
}

// JWTAuthMiddleware protege rotas verificando o token JWT.
func JWTAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if !strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
			errorJSON(w, http.StatusUnauthorized, "Token inválido ou ausente")
			return
		}
		tokenString := strings.TrimSpace(authHeader[7:])

		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			// Verifica se o algoritmo é realmente HS256
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("método de assinatura inválido")
			}
			return jwtKey, nil
		})

		if err != nil || !token.Valid {
			errorJSON(w, http.StatusUnauthorized, "Token inválido ou expirado")
			return
		}

		// Adiciona dados ao contexto
		ctx := context.WithValue(r.Context(), userIDKey, claims.UserID)
		ctx = context.WithValue(ctx, emailKey, claims.Email)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// errorJSON envia um erro padronizado em JSON.
func errorJSON(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"error": message,
	})
}

// HandleRegister simula registro de usuário e retorna um token JWT.
func HandleRegister(w http.ResponseWriter, r *http.Request) {
	userID := "mock-user-456"
	email := "novo_usuario@guardiao.com"

	token, err := GenerateToken(userID, email, 24*time.Hour)
	if err != nil {
		errorJSON(w, http.StatusInternalServerError, "Falha ao gerar token")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"message": "Usuário registrado com sucesso",
		"token":   token,
		"user_id": userID,
	})
}

// HandleLogin simula login de usuário e retorna um token JWT.
func HandleLogin(w http.ResponseWriter, r *http.Request) {
	userID := "mock-user-456"
	email := "novo_usuario@guardiao.com"

	token, err := GenerateToken(userID, email, 24*time.Hour)
	if err != nil {
		errorJSON(w, http.StatusInternalServerError, "Falha ao gerar token")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{
		"message": "Login realizado com sucesso",
		"token":   token,
		"user_id": userID,
	})
}
