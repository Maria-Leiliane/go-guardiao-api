package db

import (
	"context"
	"fmt"
	"log"
	"time"

	"go-guardiao-api/pkg/models"

	"github.com/jackc/pgx/v5"
)

// Client é a estrutura que contém a conexão real com o PostgreSQL.
type Client struct {
	conn *pgx.Conn
}

// NewDBClient estabelece a conexão real com o banco de dados PostgreSQL.
// Retorna um erro se a conexão falhar após múltiplas tentativas.
func NewDBClient(dsn string) (*Client, error) {
	const maxRetries = 5
	var err error
	var conn *pgx.Conn

	log.Printf("Tentando conectar ao PostgreSQL... DSN: %s", dsn)

	// Tentativa de reconexão com backoff
	for i := 0; i < maxRetries; i++ {
		conn, err = pgx.Connect(context.Background(), dsn)
		if err == nil {
			// Sucesso na conexão
			log.Println("Conexão com PostgreSQL estabelecida com sucesso.")
			return &Client{conn: conn}, nil
		}

		log.Printf("Falha na conexão (tentativa %d/%d): %v", i+1, maxRetries, err)
		if i < maxRetries-1 {
			time.Sleep(time.Second * time.Duration(1<<i)) // Backoff exponencial
		}
	}

	// Falha após todas as tentativas
	return nil, fmt.Errorf("falha crítica ao conectar ao PostgreSQL após %d tentativas: %w", maxRetries, err)
}

// Close fecha a conexão com o banco de dados.
func (c *Client) Close() {
	if c.conn != nil {
		c.conn.Close(context.Background())
		log.Println("Conexão com o PostgreSQL fechada.")
	}
}

// --- Métodos Reais de Implementação e Mock de Dados ---
// Projeto piloto, mantemos o mock DE DADOS, mas a estrutura do código (context, query, scan) é a mesma que a real.

// GetUserByID busca um usuário pelo ID.
func (c *Client) GetUserByID(ctx context.Context, userID string) (models.User, error) {
	log.Printf("DB MOCK: Buscando usuário %s", userID)
	// Implementação real usaria: c.conn.QueryRow(ctx, "SELECT...")
	if userID == "mock-user-456" {
		return models.User{
			ID:    userID,
			Email: "mocked@example.com",
			Name:  "Usuário Real",
			Theme: "Padrão",
		}, nil
	}
	return models.User{}, pgx.ErrNoRows
}

// CreateManaTransaction registra uma nova transação de Mana.
func (c *Client) CreateManaTransaction(ctx context.Context, tx models.ManaTransaction) error {
	log.Printf("DB MOCK: Registrando transação de Mana (Tipo: %s, Valor: %d)", tx.Type, tx.Amount)
	// Implementação real usaria: c.conn.Exec(ctx, "INSERT INTO mana_transactions...")
	return nil
}

// GetManaBalance busca o saldo atual de Mana do usuário.
func (c *Client) GetManaBalance(ctx context.Context, userID string) (int, error) {
	log.Printf("DB MOCK: Buscando saldo de Mana para o usuário %s", userID)
	// Implementação real usaria: c.conn.QueryRow(ctx, "SELECT balance...")
	return 1500, nil
}

// UpdateUser atualiza as informações do perfil do usuário.
func (c *Client) UpdateUser(ctx context.Context, user models.User) error {
	log.Printf("DB MOCK: Atualizando perfil do usuário %s", user.ID)
	// Implementação real usaria: c.conn.Exec(ctx, "UPDATE users...")
	return nil
}

// GetSupportContactsByUserID busca a rede de apoio de um usuário.
func (c *Client) GetSupportContactsByUserID(ctx context.Context, userID string) ([]models.SupportContact, error) {
	log.Printf("DB MOCK: Buscando contatos de apoio para %s", userID)
	// Implementação real usaria: c.conn.Query(ctx, "SELECT...")
	return []models.SupportContact{
		{ContactID: "c1", UserID: userID, ContactEmail: "suporte1@email.com", Nickname: "Anjo"},
	}, nil
}

// DeleteSupportContact simula a exclusão de um contato de apoio.
func (c *Client) DeleteSupportContact(ctx context.Context, contactID string) error {
	log.Printf("DB MOCK: Deletando contato de apoio ID: %s", contactID)
	// Implementação real usaria: c.conn.Exec(ctx, "DELETE...")
	return nil
}
