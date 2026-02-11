package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/mux"

	"go-guardiao-api/internal/auth"
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
	cfg := &Config{
		DBURL:     getenv("DATABASE_URL", "postgres://user:password@db:5432/guardiaodb?sslmode=disable"),
		RedisAddr: getenv("REDIS_ADDR", "cache:6379"),
		Port:      getenv("PORT", "8080"),
	}
	return cfg
}

func getenv(env, fallback string) string {
	val := os.Getenv(env)
	if val == "" {
		return fallback
	}
	return val
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
	// Inicializa os Serviços de Negócio (Injeção de Dependência)
	userService := users.NewService(dbClient)
	habitService := habits.NewService(dbClient)
	gamificationService := gamification.NewService(dbClient, cacheClient)

	// --- ROTAS DO SERVIÇO DE USUÁRIOS ---
	router.HandleFunc("/user/profile", userService.HandleGetUserProfile).Methods("GET")
	router.HandleFunc("/user/profile", userService.HandleUpdateProfile).Methods("PUT")
	router.HandleFunc("/user/support-contact", userService.HandleAddSupportContact).Methods("POST")
	router.HandleFunc("/user/support-contact", userService.HandleGetSupportContacts).Methods("GET")
	router.HandleFunc("/user/support-contact/{contactId}", userService.HandleDeleteSupportContact).Methods("DELETE")

	// --- ROTAS DO SERVIÇO DE HÁBITOS & METAS ---
	router.HandleFunc("/habits", habitService.HandleCreateHabit).Methods("POST")
	router.HandleFunc("/habits", habitService.HandleGetHabits).Methods("GET")
	// FIX: rota agora fornece {habitId} como o handler espera
	router.HandleFunc("/habits/{habitId}/log", habitService.HandleLogHabit).Methods("POST")
	router.HandleFunc("/habits/{habitId}/logs", habitService.HandleGetHabitLogs).Methods("GET")

	// --- ROTAS DO SERVIÇO DE GAMIFICAÇÃO ---
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

	// Health endpoint
	r.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}).Methods("GET")

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

	srv := &http.Server{
		Handler:      r,
		Addr:         addr,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Canal para shutdown elegante
	idleConnsClosed := make(chan struct{})
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		<-c
		log.Println("🛑 Recebido sinal de interrupção, encerrando servidor...")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			log.Printf("Erro ao encerrar servidor: %v", err)
		}
		close(idleConnsClosed)
	}()

	log.Printf("🚀 Servidor Guardião da Saúde iniciado em http://localhost%s", addr)
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("❌ Erro ao iniciar o servidor: %v", err)
	}

	<-idleConnsClosed
	log.Println("Servidor encerrado com segurança.")
}
