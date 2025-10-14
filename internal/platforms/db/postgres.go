package db

import (
	"context"
	"errors"
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
func NewDBClient(dsn string) (*Client, error) {
	const maxRetries = 5
	var err error
	var conn *pgx.Conn

	log.Printf("Tentando conectar ao PostgreSQL... DSN: %s", dsn)

	for i := 0; i < maxRetries; i++ {
		conn, err = pgx.Connect(context.Background(), dsn)
		if err == nil {
			log.Println("Conexão com PostgreSQL estabelecida com sucesso.")
			client := &Client{conn: conn}

			if initErr := client.InitSchema(context.Background()); initErr != nil {
				if closeErr := conn.Close(context.Background()); closeErr != nil {
					log.Printf("ERRO: Falha ao fechar conexão após falha de InitSchema: %v", closeErr)
				}
				return nil, fmt.Errorf("falha ao inicializar o esquema do DB: %w", initErr)
			}

			return client, nil
		}

		log.Printf("Falha na conexão (tentativa %d/%d): %v", i+1, maxRetries, err)
		if i < maxRetries-1 {
			time.Sleep(time.Second * time.Duration(1<<i))
		}
	}

	return nil, fmt.Errorf("falha crítica ao conectar ao PostgreSQL após %d tentativas: %w", maxRetries, err)
}

// Close fecha a conexão com o banco de dados.
func (c *Client) Close() {
	if c.conn != nil {
		if err := c.conn.Close(context.Background()); err != nil {
			log.Printf("ERRO: Falha ao fechar conexão com PostgreSQL: %v", err)
			return
		}
		log.Println("Conexão com o PostgreSQL fechada.")
	}
}

// InitSchema verifica a existência das tabelas e as cria se necessário.
func (c *Client) InitSchema(ctx context.Context) error {
	log.Println("Verificando e inicializando o esquema do banco de dados...")

	schemaSQL := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id VARCHAR(255) PRIMARY KEY,
			email VARCHAR(255) UNIQUE NOT NULL,
			name VARCHAR(255),
			theme VARCHAR(50),
			created_at TIMESTAMP WITHOUT TIME ZONE DEFAULT NOW()
		);`,
		`CREATE TABLE IF NOT EXISTS user_mana (
			user_id VARCHAR(255) PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
			balance INTEGER NOT NULL DEFAULT 0,
			updated_at TIMESTAMP WITHOUT TIME ZONE DEFAULT NOW()
		);`,
		`CREATE TABLE IF NOT EXISTS mana_transactions (
			id SERIAL PRIMARY KEY,
			user_id VARCHAR(255) REFERENCES users(id) ON DELETE CASCADE,
			type VARCHAR(50) NOT NULL,
			amount INTEGER NOT NULL,
			reference_id VARCHAR(255),
			created_at TIMESTAMP WITHOUT TIME ZONE DEFAULT NOW()
		);`,
		`CREATE TABLE IF NOT EXISTS support_contacts (
			contact_id VARCHAR(255) PRIMARY KEY,
			user_id VARCHAR(255) REFERENCES users(id) ON DELETE CASCADE,
			contact_email VARCHAR(255) NOT NULL,
			nickname VARCHAR(255),
			notification_preference VARCHAR(50),
			created_at TIMESTAMP WITHOUT TIME ZONE DEFAULT NOW()
		);`,
		`CREATE TABLE IF NOT EXISTS habits (
			id VARCHAR(255) PRIMARY KEY,
			user_id VARCHAR(255) REFERENCES users(id) ON DELETE CASCADE,
			name VARCHAR(255) NOT NULL,
			goal_type VARCHAR(50),
			frequency VARCHAR(50),
			created_at TIMESTAMP WITHOUT TIME ZONE DEFAULT NOW()
		);`,
		`CREATE TABLE IF NOT EXISTS habit_logs (
			id SERIAL PRIMARY KEY,
			habit_id VARCHAR(255) REFERENCES habits(id) ON DELETE CASCADE,
			user_id VARCHAR(255) REFERENCES users(id) ON DELETE CASCADE,
			"value" INTEGER NOT NULL,
			"timestamp" TIMESTAMP WITHOUT TIME ZONE DEFAULT NOW()
		);`,
	}

	for _, sql := range schemaSQL {
		if _, err := c.conn.Exec(ctx, sql); err != nil {
			return fmt.Errorf("falha ao executar SQL de inicialização: %w", err)
		}
	}

	log.Println("Esquema do DB inicializado com sucesso. Tabelas verificadas/criadas.")
	return nil
}

// --- Funções Reais de Persistência (Serviço de Usuários) ---

// CreateUser insere um novo usuário e inicializa o saldo de Mana.
func (c *Client) CreateUser(ctx context.Context, user models.User) error {
	tx, err := c.conn.Begin(ctx)
	if err != nil {
		return fmt.Errorf("falha ao iniciar transação: %w", err)
	}
	defer func() {
		if rberr := tx.Rollback(ctx); rberr != nil && !errors.Is(rberr, pgx.ErrTxClosed) {
			log.Printf("ERRO CRÍTICO: Falha ao executar Rollback: %v", rberr)
		}
	}()

	userInsertSQL := `INSERT INTO users (id, email, name, theme) VALUES ($1, $2, $3, $4)`
	if _, err := tx.Exec(ctx, userInsertSQL, user.ID, user.Email, user.Name, user.Theme); err != nil {
		return fmt.Errorf("falha ao inserir usuário: %w", err)
	}

	manaInsertSQL := `INSERT INTO user_mana (user_id, balance) VALUES ($1, 0)`
	if _, err := tx.Exec(ctx, manaInsertSQL, user.ID); err != nil {
		return fmt.Errorf("falha ao inicializar saldo de mana: %w", err)
	}

	return tx.Commit(ctx)
}

// GetUserByID busca um usuário pelo ID.
func (c *Client) GetUserByID(ctx context.Context, userID string) (models.User, error) {
	user := models.User{}
	sql := `SELECT id, email, name, theme FROM users WHERE id = $1`

	err := c.conn.QueryRow(ctx, sql, userID).Scan(
		&user.ID,
		&user.Email,
		&user.Name,
		&user.Theme,
	)
	if err != nil {
		return models.User{}, err
	}
	return user, nil
}

// UpdateUser atualiza as informações do perfil do usuário.
func (c *Client) UpdateUser(ctx context.Context, user models.User) error {
	sql := `UPDATE users SET name = $2, theme = $3 WHERE id = $1`
	cmdTag, err := c.conn.Exec(ctx, sql, user.ID, user.Name, user.Theme)
	if err != nil {
		return err
	}
	if cmdTag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

// CreateSupportContact insere um novo contato de apoio.
func (c *Client) CreateSupportContact(ctx context.Context, contact models.SupportContact) error {
	sql := `INSERT INTO support_contacts (contact_id, user_id, contact_email, nickname, notification_preference)
            VALUES ($1, $2, $3, $4, $5)`

	contact.ContactID = fmt.Sprintf("contact-%d", time.Now().UnixNano())

	_, err := c.conn.Exec(ctx, sql, contact.ContactID, contact.UserID, contact.ContactEmail, contact.Nickname, contact.NotificationPreference)
	return err
}

// GetSupportContactsByUserID busca a rede de apoio de um usuário.
func (c *Client) GetSupportContactsByUserID(ctx context.Context, userID string) ([]models.SupportContact, error) {
	sql := `SELECT contact_id, user_id, contact_email, nickname, notification_preference FROM support_contacts WHERE user_id = $1`
	rows, err := c.conn.Query(ctx, sql, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var contacts []models.SupportContact
	for rows.Next() {
		contact := models.SupportContact{}
		if err := rows.Scan(&contact.ContactID, &contact.UserID, &contact.ContactEmail, &contact.Nickname, &contact.NotificationPreference); err != nil {
			return nil, err
		}
		contacts = append(contacts, contact)
	}
	return contacts, nil
}

// DeleteSupportContact simula a exclusão de um contato de apoio.
func (c *Client) DeleteSupportContact(ctx context.Context, contactID string) error {
	sql := `DELETE FROM support_contacts WHERE contact_id = $1`
	cmdTag, err := c.conn.Exec(ctx, sql, contactID)
	if err != nil {
		return err
	}
	if cmdTag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

// --- Funções Reais de Persistência (Serviço de Hábitos) ---

// CreateHabit insere um novo hábito.
func (c *Client) CreateHabit(ctx context.Context, habit models.Habit) (string, error) {
	habit.ID = fmt.Sprintf("habit-%d", time.Now().UnixNano())
	sql := `INSERT INTO habits (id, user_id, name, goal_type, frequency) VALUES ($1, $2, $3, $4, $5)`

	_, err := c.conn.Exec(ctx, sql, habit.ID, habit.UserID, habit.Name, habit.GoalType, habit.Frequency)
	if err != nil {
		return "", err
	}
	return habit.ID, nil
}

// GetHabitsByUserID busca todos os hábitos de um usuário.
func (c *Client) GetHabitsByUserID(ctx context.Context, userID string) ([]models.Habit, error) {
	sql := `SELECT id, user_id, name, goal_type, frequency, created_at FROM habits WHERE user_id = $1 ORDER BY created_at DESC`
	rows, err := c.conn.Query(ctx, sql, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var habits []models.Habit
	for rows.Next() {
		habit := models.Habit{}
		if err := rows.Scan(&habit.ID, &habit.UserID, &habit.Name, &habit.GoalType, &habit.Frequency, &habit.CreatedAt); err != nil {
			return nil, err
		}
		habits = append(habits, habit)
	}
	return habits, nil
}

// LogHabit registra um log de progresso para um hábito.
func (c *Client) LogHabit(ctx context.Context, logData models.HabitLog) error {
	sql := `INSERT INTO habit_logs (habit_id, user_id, value) VALUES ($1, $2, $3)`

	_, err := c.conn.Exec(ctx, sql, logData.HabitID, logData.UserID, logData.Value)
	if err != nil {
		return err
	}

	log.Printf("DB Ação: Log de hábito %s registrado.", logData.HabitID)
	return nil
}

// GetHabitLogs busca os logs de um hábito específico.
func (c *Client) GetHabitLogs(ctx context.Context, habitID string) ([]models.HabitLog, error) {
	sql := `SELECT id, habit_id, user_id, value, "timestamp" FROM habit_logs WHERE habit_id = $1 ORDER BY "timestamp" DESC`
	rows, err := c.conn.Query(ctx, sql, habitID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []models.HabitLog
	for rows.Next() {
		logItem := models.HabitLog{}
		if err := rows.Scan(&logItem.ID, &logItem.HabitID, &logItem.UserID, &logItem.Value, &logItem.Timestamp); err != nil {
			return nil, err
		}
		logs = append(logs, logItem)
	}
	return logs, nil
}

// --- Funções Reais de Persistência (Serviço de Gamificação - Transacional) ---

// GetManaBalance busca o saldo atual de Mana do usuário.
func (c *Client) GetManaBalance(ctx context.Context, userID string) (int, error) {
	var balance int
	sql := `SELECT balance FROM user_mana WHERE user_id = $1`

	err := c.conn.QueryRow(ctx, sql, userID).Scan(&balance)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, nil
		}
		return 0, err
	}
	return balance, nil
}

// UpdateManaBalance atualiza o saldo e registra a transação.
func (c *Client) UpdateManaBalance(ctx context.Context, txData models.ManaTransaction) error {
	tx, err := c.conn.Begin(ctx)
	if err != nil {
		return fmt.Errorf("falha ao iniciar transação de mana: %w", err)
	}
	defer func() {
		if rberr := tx.Rollback(ctx); rberr != nil && !errors.Is(rberr, pgx.ErrTxClosed) {
			log.Printf("ERRO CRÍTICO: Falha ao executar Rollback na Mana: %v", rberr)
		}
	}()

	// 1. Atualiza o saldo (dedução/adição)
	updateSQL := `UPDATE user_mana SET balance = balance + $1, updated_at = NOW() WHERE user_id = $2 RETURNING balance`
	var newBalance int

	err = tx.QueryRow(ctx, updateSQL, txData.Amount, txData.UserID).Scan(&newBalance)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("usuário de mana não encontrado: %w", err)
		}
		return fmt.Errorf("falha ao atualizar saldo: %w", err)
	}

	// 2. Registra a transação de auditoria
	insertSQL := `INSERT INTO mana_transactions (user_id, type, amount, reference_id) VALUES ($1, $2, $3, $4)`
	if _, err := tx.Exec(ctx, insertSQL, txData.UserID, txData.Type, txData.Amount, txData.ReferenceID); err != nil {
		return fmt.Errorf("falha ao registrar transação de mana: %w", err)
	}

	// 3. Commit
	return tx.Commit(ctx)
}

// CreateManaTransaction é mantida como um alias para a função transacional
func (c *Client) CreateManaTransaction(ctx context.Context, tx models.ManaTransaction) error {
	return c.UpdateManaBalance(ctx, tx)
}
