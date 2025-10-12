package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// jwtKey simula a chave secreta. Em produção, usar AWS Secrets Manager!
var jwtKey = []byte("sua-chave-secreta-muito-forte")

// Claims define a estrutura dos dados no token JWT.
type Claims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

// Helper para obter o ID do usuário do contexto.
// ESTA FUNÇÃO É CRÍTICA E É USADA POR TODOS OS HANDLERS PROTEGIDOS.
func GetUserIDFromContext(r *http.Request) (string, error) {
	// A chave "userID" foi definida no JWTAuthMiddleware.
	userID, ok := r.Context().Value("userID").(string)
	if !ok {
		return "", fmt.Errorf("user ID não encontrado no contexto")
	}
	return userID, nil
}

// GenerateToken cria um novo token JWT.
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

// JWTAuthMiddleware verifica o token JWT em cada requisição protegida.
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

		// Anexa o UserID ao Contexto
		ctx := context.WithValue(r.Context(), "userID", claims.UserID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// HandleRegister (Mock) - Simula o registro de usuário.
func HandleRegister(w http.ResponseWriter, r *http.Request) {
	// Lógica real: salvar usuário no RDS, gerar token.
	mockUserID := "mock-user-123"
	token, _ := GenerateToken(mockUserID, "teste@guardiao.com")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Usuário registrado com sucesso",
		"token":   token,
	})
}

// HandleLogin (Mock) - Simula o login de usuário.
func HandleLogin(w http.ResponseWriter, r *http.Request) {
	// Lógica real: validar senha, gerar token.
	mockUserID := "mock-user-123"
	token, _ := GenerateToken(mockUserID, "teste@guardiao.com")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Login realizado com sucesso",
		"token":   token,
	})
}
