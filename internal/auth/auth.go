package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// A chave secreta para assinar os tokens (em produção, deve vir de um gerenciador de segredos).
var jwtKey = []byte("sua-chave-secreta-muito-forte")

// Claims define a estrutura dos dados que serão armazenados no token JWT.
type Claims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

// GetUserIDFromContext é uma função auxiliar para obter o ID do usuário
// do contexto da requisição, após a validação do JWT.
func GetUserIDFromContext(r *http.Request) (string, error) {
	userID, ok := r.Context().Value("userID").(string)
	if !ok {
		return "", fmt.Errorf("user ID não encontrado no contexto")
	}
	return userID, nil
}

// GenerateToken cria e assina um novo token JWT para um usuário.
func GenerateToken(userID, email string) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtKey)
}

// JWTAuthMiddleware é um middleware para proteger rotas. Ele verifica
// a validade do token JWT em cada requisição.
func JWTAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" || len(authHeader) < 7 || authHeader[:7] != "Bearer " {
			http.Error(w, "Token inválido ou ausente", http.StatusUnauthorized)
			return
		}
		tokenString := authHeader[7:]
		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return jwtKey, nil
		})

		if err != nil || !token.Valid {
			http.Error(w, "Token inválido ou expirado", http.StatusUnauthorized)
			return
		}

		// Se o token for válido, anexa o ID do usuário ao contexto da requisição
		// para que os handlers possam acessá-lo.
		ctx := context.WithValue(r.Context(), "userID", claims.UserID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// HandleRegister simula o processo de registro de usuário.
// Retorna um token JWT para o cliente.
func HandleRegister(w http.ResponseWriter, r *http.Request) {
	// Na implementação real, você validaria os dados e criaria um novo registro no banco de dados.
	mockUserID := "mock-user-456"
	token, _ := GenerateToken(mockUserID, "novo_usuario@guardiao.com")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Usuário registrado com sucesso",
		"token":   token,
		"user_id": mockUserID,
	})
}

// HandleLogin simula o processo de login.
// Retorna um token JWT se as credenciais estiverem corretas.
func HandleLogin(w http.ResponseWriter, r *http.Request) {
	// Na implementação real, você verificaria as credenciais contra o banco de dados.
	mockUserID := "mock-user-456"
	token, _ := GenerateToken(mockUserID, "novo_usuario@guardiao.com")

	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(map[string]string{
		"message": "Login realizado com sucesso",
		"token":   token,
		"user_id": mockUserID,
	})
	if err != nil {
		return
	}
}
