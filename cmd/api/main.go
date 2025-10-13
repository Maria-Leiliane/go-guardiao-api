package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"

	"go-guardiao-api/internal/auth"
	"go-guardiao-api/internal/habits"
	"go-guardiao-api/internal/users"
)

// defineAuthRoutes configura as rotas públicas (sem JWT).
func defineAuthRoutes(router *mux.Router) {
	router.HandleFunc("/register", auth.HandleRegister).Methods("POST")
	router.HandleFunc("/login", auth.HandleLogin).Methods("POST")
}

// main função principal que inicializa o servidor HTTP.
func main() {
	r := mux.NewRouter()

	// 1. Rotas de Autenticação (Públicas)
	authRouter := r.PathPrefix("/api/v1/auth").Subrouter()
	defineAuthRoutes(authRouter)

	// 2. Rotas Protegidas (Requerem JWT)
	apiRouter := r.PathPrefix("/api/v1").Subrouter()
	apiRouter.Use(auth.JWTAuthMiddleware) // Aplica a proteção JWT a todas as rotas abaixo

	// --- ROTAS DO SERVIÇO DE USUÁRIOS ---
	apiRouter.HandleFunc("/user/profile", users.HandleGetUserProfile).Methods("GET")
	apiRouter.HandleFunc("/user/profile", users.HandleUpdateProfile).Methods("PUT")
	apiRouter.HandleFunc("/user/support-contact", users.HandleAddSupportContact).Methods("POST")
	apiRouter.HandleFunc("/user/support-contact", users.HandleGetSupportContacts).Methods("GET")
	apiRouter.HandleFunc("/user/support-contact/{contactId}", users.HandleDeleteSupportContact).Methods("DELETE")

	// --- ROTAS DO SERVIÇO DE HÁBITOS & METAS (NOVO) ---
	apiRouter.HandleFunc("/habits", habits.HandleCreateHabit).Methods("POST")
	apiRouter.HandleFunc("/habits", habits.HandleGetHabits).Methods("GET")
	apiRouter.HandleFunc("/habits/log", habits.HandleLogHabit).Methods("POST")
	apiRouter.HandleFunc("/habits/{habitId}/logs", habits.HandleGetHabitLogs).Methods("GET")

	// Rota de exemplo para testar a proteção
	apiRouter.HandleFunc("/protected", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Olá, mundo protegido!"))
	}).Methods("GET")

	// Inicia o servidor
	port := "8080"
	log.Printf("Servidor Guardião da Saúde iniciado em http://localhost:%s", port)

	srv := &http.Server{
		Handler:      r,
		Addr:         fmt.Sprintf(":%s", port),
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("Não foi possível iniciar o servidor: %v", err)
	}
}
