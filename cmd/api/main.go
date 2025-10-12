package api

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"

	// Mockup de pacotes internos (seriam implementados em internal/app/...)
	"guardian/internal/app/auth"
)

// Constantes de Configuração
const (
	port = ":8080"
)

func main() {
	// 1. Inicializa o Roteador
	router := mux.NewRouter()

	// 2. Configura as Rotas
	setupRoutes(router)

	// 3. Inicializa o Servidor HTTP
	fmt.Printf("Servidor Guardião da Saúde rodando na porta %s\n", port)

	// Configurações ideais para produção (timeouts)
	srv := &http.Server{
		Handler:      router,
		Addr:         port,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Inicia o servidor (log.Fatal irá encerrar a aplicação em caso de erro)
	log.Fatal(srv.ListenAndServe())
}

// setupRoutes define todas as rotas da nossa API, separando rotas públicas e protegidas.
func setupRoutes(r *mux.Router) {
	// --- Rotas Públicas (Não requerem JWT) ---
	public := r.PathPrefix("/api/v1/public").Subrouter()

	// Rotas do Serviço de Autenticação
	public.HandleFunc("/register", auth.HandleRegister).Methods("POST")
	public.HandleFunc("/login", auth.HandleLogin).Methods("POST")

	// Rotas do Conteúdo Educacional (pode ser público)
	public.HandleFunc("/content", HandleGetContent).Methods("GET")

	// --- Rotas Protegidas (Requerem JWT - Middleware de Autenticação) ---
	protected := r.PathPrefix("/api/v1/protected").Subrouter()

	// Aplica o middleware de autenticação JWT em todas as rotas protegidas
	protected.Use(auth.JWTAuthMiddleware)

	// Rotas do Serviço de Usuários e Perfil
	protected.HandleFunc("/users/{id}", HandleGetUserProfile).Methods("GET")

	// Rotas do Serviço de Hábitos & Metas
	protected.HandleFunc("/habits", HandleCreateHabit).Methods("POST")

	// Rotas do Serviço de Gamificação (Mana)
	protected.HandleFunc("/mana/balance", HandleGetManaBalance).Methods("GET")
	protected.HandleFunc("/challenges/complete", HandleCompleteChallenge).Methods("POST")

	// Rotas do Serviço de Suporte & Acolhimento
	protected.HandleFunc("/support/contacts", HandleAddSupportContact).Methods("POST")
}

// Mock Handlers (Substituir por lógica real dos serviços)
func HandleGetContent(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "Conteúdo Educacional - Público"}`))
}

func HandleGetUserProfile(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "Perfil do Usuário - Protegido"}`))
}

func HandleCreateHabit(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(`{"message": "Hábito criado com sucesso - Protegido"}`))
}

func HandleGetManaBalance(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"mana_balance": 1500}`))
}

func HandleCompleteChallenge(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "Desafio concluído, Mana calculada e enviada para SQS - Protegido"}`))
}

func HandleAddSupportContact(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(`{"message": "Contato de suporte adicionado - Protegido"}`))
}
