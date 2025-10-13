package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"

	// Importa os pacotes internos de Autenticação e Usuários
	"go-guardiao-api/internal/auth"
	"go-guardiao-api/internal/users" // <- Importa o seu novo serviço
)

// defineAuthRoutes configura as rotas públicas (sem JWT).
func defineAuthRoutes(router *mux.Router) {
	router.HandleFunc("/register", auth.HandleRegister).Methods("POST")
	router.HandleFunc("/login", auth.HandleLogin).Methods("POST")
}

// main é a função principal que inicializa o servidor HTTP.
func main() {
	r := mux.NewRouter()

	// 1. Rotas de Autenticação (Públicas)
	authRouter := r.PathPrefix("/api/v1/auth").Subrouter()
	defineAuthRoutes(authRouter)

	// 2. Rotas Protegidas (Requerem JWT)
	apiRouter := r.PathPrefix("/api/v1").Subrouter()
	apiRouter.Use(auth.JWTAuthMiddleware)

	// Rotas do Serviço de Usuários (agora roteadas a partir do pacote 'users')
	apiRouter.HandleFunc("/user/profile", users.HandleGetUserProfile).Methods("GET")
	apiRouter.HandleFunc("/user/profile", users.HandleUpdateProfile).Methods("PUT")
	apiRouter.HandleFunc("/user/support-contact", users.HandleAddSupportContact).Methods("POST")
	apiRouter.HandleFunc("/user/support-contact", users.HandleGetSupportContacts).Methods("GET")
	apiRouter.HandleFunc("/user/support-contact/{contactId}", users.HandleDeleteSupportContact).Methods("DELETE")

	// Rota de exemplo para testar a proteção (será substituída por outros serviços)
	apiRouter.HandleFunc("/protected", func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte("Olá, mundo protegido!"))
		if err != nil {
			return
		}
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
