package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"github.com/jackc/pgx/v5"

	"go-guardiao-api/internal/platforms/db"
	"go-guardiao-api/pkg/models"
)

// ===== Context keys =====

type contextKey string

const (
	userIDKey contextKey = "userID"
	emailKey  contextKey = "userEmail"
)

// ===== JWT secret =====

var jwtKey = []byte(getEnv("JWT_SECRET", "sua-chave-secreta-muito-forte"))

func getEnv(k, fallback string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return fallback
}

// ===== Claims =====

type Claims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

// ===== JWT helpers =====

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

func GetUserIDFromContext(r *http.Request) (string, error) {
	userID, ok := r.Context().Value(userIDKey).(string)
	if !ok || userID == "" {
		return "", errors.New("user ID não encontrado no contexto")
	}
	return userID, nil
}

// ===== Password rules + bcrypt =====

// Regras: mínimo 8, 1 maiúscula, 1 minúscula, 1 dígito, 1 especial
var (
	reUpper   = regexp.MustCompile(`[A-Z]`)
	reLower   = regexp.MustCompile(`[a-z]`)
	reDigit   = regexp.MustCompile(`[0-9]`)
	reSpecial = regexp.MustCompile(`[!@#\$%\^&\*\(\)_\+\-=\{\}\[\]:;"'<>,\.\?/\\|~]`)
)

func ValidatePassword(pw string) error {
	if len(pw) < 8 {
		return errors.New("senha deve ter pelo menos 8 caracteres")
	}
	if !reUpper.MatchString(pw) {
		return errors.New("senha deve conter ao menos uma letra maiúscula")
	}
	if !reLower.MatchString(pw) {
		return errors.New("senha deve conter ao menos uma letra minúscula")
	}
	if !reDigit.MatchString(pw) {
		return errors.New("senha deve conter ao menos um dígito")
	}
	if !reSpecial.MatchString(pw) {
		return errors.New("senha deve conter ao menos um caractere especial")
	}
	return nil
}

func HashPassword(pw string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)
	return string(b), err
}

func CheckPassword(pw, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(pw))
}

// ===== Middleware =====

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
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("método de assinatura inválido")
			}
			return jwtKey, nil
		})

		if err != nil || !token.Valid {
			errorJSON(w, http.StatusUnauthorized, "Token inválido ou expirado")
			return
		}

		ctx := context.WithValue(r.Context(), userIDKey, claims.UserID)
		ctx = context.WithValue(ctx, emailKey, claims.Email)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// ===== JSON helpers =====

func errorJSON(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"error": message,
	})
}

func okJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

// ===== Auth payload =====

type AuthPayload struct {
	Email    string `json:"email"`
	Name     string `json:"name,omitempty"`
	Password string `json:"password,omitempty"`
}

// ===== Handlers (com DB) =====

// Registro: exige email + senha forte, salva hash e retorna token
func HandleRegisterWithDB(w http.ResponseWriter, r *http.Request, dbClient *db.Client) {
	var p AuthPayload
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil || p.Email == "" {
		errorJSON(w, http.StatusBadRequest, "Email é obrigatório.")
		return
	}
	if err := ValidatePassword(p.Password); err != nil {
		errorJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	u, err := dbClient.GetUserByEmail(r.Context(), p.Email)
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		hash, err := HashPassword(p.Password)
		if err != nil {
			errorJSON(w, http.StatusInternalServerError, "Falha ao processar senha")
			return
		}
		newUser := models.User{
			Email:        p.Email,
			Name:         p.Name,
			Theme:        "",
			PasswordHash: hash,
		}
		if err = dbClient.CreateUser(r.Context(), newUser); err != nil {
			errorJSON(w, http.StatusInternalServerError, "Falha ao criar usuário")
			return
		}
		// buscar com ID preenchido
		u, _ = dbClient.GetUserByEmail(r.Context(), p.Email)

	case err == nil:
		errorJSON(w, http.StatusConflict, "Email já cadastrado")
		return

	default:
		errorJSON(w, http.StatusInternalServerError, "Falha ao consultar usuário")
		return
	}

	token, err := GenerateToken(u.ID, u.Email, 24*time.Hour)
	if err != nil {
		errorJSON(w, http.StatusInternalServerError, "Falha ao gerar token")
		return
	}

	okJSON(w, http.StatusCreated, map[string]string{
		"message": "Usuário registrado com sucesso",
		"token":   token,
		"user_id": u.ID,
	})
}

// Login: exige email + senha, valida bcrypt e retorna token
func HandleLoginWithDB(w http.ResponseWriter, r *http.Request, dbClient *db.Client) {
	var p AuthPayload
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil || p.Email == "" || p.Password == "" {
		errorJSON(w, http.StatusBadRequest, "Email e senha são obrigatórios.")
		return
	}

	u, err := dbClient.GetUserByEmail(r.Context(), p.Email)
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		errorJSON(w, http.StatusUnauthorized, "Usuário não encontrado")
		return
	case err != nil:
		errorJSON(w, http.StatusInternalServerError, "Falha ao consultar usuário")
		return
	}

	if err := CheckPassword(p.Password, u.PasswordHash); err != nil {
		errorJSON(w, http.StatusUnauthorized, "Senha inválida")
		return
	}

	token, err := GenerateToken(u.ID, u.Email, 24*time.Hour)
	if err != nil {
		errorJSON(w, http.StatusInternalServerError, "Falha ao gerar token")
		return
	}

	okJSON(w, http.StatusOK, map[string]string{
		"message": "Login realizado com sucesso",
		"token":   token,
		"user_id": u.ID,
	})
}
