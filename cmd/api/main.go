package main

import (
	"context"
	"errors"
	"fmt"
	"go-guardiao-api/internal/auth"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"

	"go-guardiao-api/internal/gamification"
	"go-guardiao-api/internal/habits"
	"go-guardiao-api/internal/platforms/cache"
	"go-guardiao-api/internal/platforms/db"
	"go-guardiao-api/internal/users"
)

type Config struct {
	DBURL     string
	RedisAddr string
	Port      string
}

func loadConfig() *Config {
	return &Config{
		DBURL:     getenv("DATABASE_URL", "postgres://user:password@db:5432/guardiaodb?sslmode=disable"),
		RedisAddr: getenv("REDIS_ADDR", "cache:6379"),
		Port:      getenv("PORT", "8080"),
	}
}

func getenv(env, fallback string) string {
	if v := os.Getenv(env); v != "" {
		return v
	}
	return fallback
}

func mustInitDB(dbURL string) *db.Client {
	dbClient, err := db.NewDBClient(dbURL)
	if err != nil {
		log.Fatalf("❌ Falha ao conectar ao banco: %v", err)
	}
	return dbClient
}

func tryInitCache(redisAddr string) *cache.Client {
	cacheClient, err := cache.NewCacheClient(redisAddr, "")
	if err != nil {
		log.Printf("⚠️ Redis indisponível (continuando sem cache): %v", err)
		return nil
	}
	return cacheClient
}

// defineServiceRoutes configura todas as rotas protegidas e injeta o DB e Cache.
func defineServiceRoutes(router *mux.Router, dbClient *db.Client, cacheClient *cache.Client) {
	userService := users.NewService(dbClient)
	habitService := habits.NewService(dbClient)
	gamificationService := gamification.NewService(dbClient, cacheClient)

	// --- USUÁRIOS ---
	router.HandleFunc("/user/profile", userService.HandleGetUserProfile).Methods("GET")
	router.HandleFunc("/user/profile", userService.HandleUpdateProfile).Methods("PUT")
	router.HandleFunc("/user/support-contact", userService.HandleAddSupportContact).Methods("POST")
	router.HandleFunc("/user/support-contact", userService.HandleGetSupportContacts).Methods("GET")
	router.HandleFunc("/user/support-contact/{contactId}", userService.HandleDeleteSupportContact).Methods("DELETE")

	// --- HÁBITOS ---
	router.HandleFunc("/habits", habitService.HandleCreateHabit).Methods("POST")
	router.HandleFunc("/habits", habitService.HandleGetHabits).Methods("GET")
	router.HandleFunc("/habits/{habitId}/log", habitService.HandleLogHabit).Methods("POST")
	router.HandleFunc("/habits/{habitId}/logs", habitService.HandleGetHabitLogs).Methods("GET")

	// --- GAMIFICAÇÃO ---
	router.HandleFunc("/mana/balance", gamificationService.HandleGetManaBalance).Methods("GET")
	router.HandleFunc("/mana/redeem", gamificationService.HandleRedeemReward).Methods("POST")
	router.HandleFunc("/challenges", gamificationService.HandleListChallenges).Methods("GET")
	router.HandleFunc("/leaderboard", gamificationService.HandleGetLeaderboard).Methods("GET")
}

func setupRouter(dbClient *db.Client, cacheClient *cache.Client) *mux.Router {
	r := mux.NewRouter().StrictSlash(true)

	// Rotas Públicas (Auth)
	r.HandleFunc("/api/v1/auth/register", func(w http.ResponseWriter, r *http.Request) {
		auth.HandleRegisterWithDB(w, r, dbClient)
	}).Methods("POST")
	r.HandleFunc("/api/v1/auth/login", func(w http.ResponseWriter, r *http.Request) {
		auth.HandleLoginWithDB(w, r, dbClient)
	}).Methods("POST")

	// Rotas Protegidas (API) - JWT Middleware
	apiRouter := r.PathPrefix("/api/v1").Subrouter()
	apiRouter.Use(auth.JWTAuthMiddleware)
	defineServiceRoutes(apiRouter, dbClient, cacheClient)

	// Health
	r.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}).Methods("GET")

	// (Opcional) Captura de preflight genérico
	r.PathPrefix("/").Methods(http.MethodOptions).HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	return r
}

func main() {
	cfg := loadConfig()
	addr := fmt.Sprintf(":%s", cfg.Port)

	dbClient := mustInitDB(cfg.DBURL)
	defer dbClient.Close()

	cacheClient := tryInitCache(cfg.RedisAddr)
	if cacheClient != nil {
		defer cacheClient.Close()
	}

	r := setupRouter(dbClient, cacheClient)

	// CORS (Gorilla)
	origins := []string{
		"http://localhost:4200",
		"https://seu-projeto.vercel.app",
		"https://*.vercel.app",
	}
	corsHandler := handlers.CORS(
		handlers.AllowedOrigins(origins),
		handlers.AllowedMethods([]string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}),
		handlers.AllowedHeaders([]string{"Authorization", "Content-Type"}),
		handlers.AllowCredentials(), // remova se não precisar de cookies/credenciais
		handlers.MaxAge(12*60*60),
	)

	srv := &http.Server{
		Handler:      corsHandler(r), // aplica CORS no topo da cadeia
		Addr:         addr,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Shutdown gracioso
	idleConnsClosed := make(chan struct{})
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		<-c
		log.Println("🛑 Encerrando servidor...")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			log.Printf("Erro ao encerrar servidor: %v", err)
		}
		close(idleConnsClosed)
	}()

	log.Printf("🚀 Servidor Guardião da Saúde em http://localhost%s", addr)
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("❌ Erro ao iniciar o servidor: %v", err)
	}

	<-idleConnsClosed
	log.Println("✅ Servidor encerrado com segurança.")
}
