package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"

	"go-guardiao-api/internal/auth"
	"go-guardiao-api/internal/gamification"
	"go-guardiao-api/internal/habits"
	"go-guardiao-api/internal/platforms/db"
	"go-guardiao-api/internal/users"
)

// defineAuthRoutes configura as rotas públicas (sem JWT).
func defineAuthRoutes(router *mux.Router) {
	router.HandleFunc("/register", auth.HandleRegister).Methods("POST")
	router.HandleFunc("/login", auth.HandleLogin).Methods("POST")
}

// defineServiceRoutes configura todas as rotas protegidas e injeta a dependência do DB.
func defineServiceRoutes(router *mux.Router, dbClient *db.Client) {
	// 1. Inicializa os Serviços de Negócio (Injeção de Dependência)
	userService := users.NewService(dbClient)
	habitService := habits.NewService(dbClient)
	gamificationService := gamification.NewService(dbClient) // <--- INSTÂNCIA DO SERVIÇO DE GAMIFICAÇÃO

	// --- ROTAS DO SERVIÇO DE USUÁRIOS ---
	router.HandleFunc("/user/profile", userService.HandleGetUserProfile).Methods("GET")
	router.HandleFunc("/user/profile", userService.HandleUpdateProfile).Methods("PUT")
	router.HandleFunc("/user/support-contact", userService.HandleAddSupportContact).Methods("POST")
	router.HandleFunc("/user/support-contact", userService.HandleGetSupportContacts).Methods("GET")
	router.HandleFunc("/user/support-contact/{contactId}", userService.HandleDeleteSupportContact).Methods("DELETE")

	// --- ROTAS DO SERVIÇO DE HÁBITOS & METAS ---
	router.HandleFunc("/habits", habitService.HandleCreateHabit).Methods("POST")
	router.HandleFunc("/habits", habitService.HandleGetHabits).Methods("GET")
	router.HandleFunc("/habits/log", habitService.HandleLogHabit).Methods("POST")
	router.HandleFunc("/habits/{habitId}/logs", habitService.HandleGetHabitLogs).Methods("GET")

	// --- ROTAS DO SERVIÇO DE GAMIFICAÇÃO (CHAMANDO MÉTODOS DA STRUCT) ---
	router.HandleFunc("/mana/balance", gamificationService.HandleGetManaBalance).Methods("GET")
	router.HandleFunc("/mana/redeem", gamificationService.HandleRedeemReward).Methods("POST")
	router.HandleFunc("/challenges", gamificationService.HandleListChallenges).Methods("GET")
	router.HandleFunc("/leaderboard", gamificationService.HandleGetLeaderboard).Methods("GET")

	// Rota de exemplo para testar a proteção
	router.HandleFunc("/protected", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte("Olá, mundo protegido!"))
		if err != nil {
			log.Printf("Erro ao escrever a resposta na rota protegida: %v", err)
			return
		}
	}).Methods("GET")
}

// main é a função principal que inicializa o servidor HTTP.
func main() {
	// SIMULAÇÃO DE CRIAÇÃO DO CLIENTE DB PARA USO LOCAL
	dbClient, _ := db.NewDBClient("MOCK_DSN")
	defer func() {
		if dbClient != nil {
			dbClient.Close()
		}
	}()

	r := mux.NewRouter()

	// 1. Rotas de Autenticação (Públicas)
	authRouter := r.PathPrefix("/api/v1/auth").Subrouter()
	defineAuthRoutes(authRouter)

	// 2. Rotas Protegidas (Injeção de Dependência)
	apiRouter := r.PathPrefix("/api/v1").Subrouter()
	apiRouter.Use(auth.JWTAuthMiddleware)

	defineServiceRoutes(apiRouter, dbClient)

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
