package db

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"go-guardiao-api/pkg/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Client struct {
	pool *pgxpool.Pool
}

// NewDBClient cria pool de conexões com retries exponenciais.
func NewDBClient(dsn string) (*Client, error) {
	const maxRetries = 5
	var err error
	var pool *pgxpool.Pool

	log.Printf("Tentando conectar ao PostgreSQL... DSN: %s", dsn)

	for i := 0; i < maxRetries; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		pool, err = pgxpool.New(ctx, dsn)
		cancel()
		if err == nil {
			ctx, cancel2 := context.WithTimeout(context.Background(), 10*time.Second)
			err = pool.Ping(ctx)
			cancel2()
			if err == nil {
				client := &Client{pool: pool}
				if initErr := client.InitSchema(context.Background()); initErr != nil {
					pool.Close()
					return nil, fmt.Errorf("falha ao inicializar o esquema do DB: %w", initErr)
				}
				log.Println("Conexão com PostgreSQL estabelecida com sucesso.")
				return client, nil
			}
		}
		log.Printf("Falha na conexão (tentativa %d/%d): %v", i+1, maxRetries, err)
		if i < maxRetries-1 {
			time.Sleep(time.Second * time.Duration(1<<i))
		}
	}
	return nil, fmt.Errorf("falha crítica ao conectar ao PostgreSQL após %d tentativas: %w", maxRetries, err)
}

func (c *Client) Close() {
	if c.pool != nil {
		c.pool.Close()
		log.Println("Conexão com o PostgreSQL fechada.")
	}
}

func (c *Client) InitSchema(ctx context.Context) error {
	log.Println("Verificando e inicializando o esquema do banco de dados...")
	schemaSQL := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id UUID PRIMARY KEY,
			email VARCHAR(255) UNIQUE NOT NULL,
			name VARCHAR(255),
			theme VARCHAR(50),
			created_at TIMESTAMP DEFAULT NOW()
		);`,
		`CREATE TABLE IF NOT EXISTS user_mana (
			user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
			balance INTEGER NOT NULL DEFAULT 0,
			updated_at TIMESTAMP DEFAULT NOW()
		);`,
		`CREATE TABLE IF NOT EXISTS mana_transactions (
			id SERIAL PRIMARY KEY,
			user_id UUID REFERENCES users(id) ON DELETE CASCADE,
			type VARCHAR(50) NOT NULL,
			amount INTEGER NOT NULL,
			reference_id VARCHAR(255),
			created_at TIMESTAMP DEFAULT NOW()
		);`,
		`CREATE TABLE IF NOT EXISTS support_contacts (
			contact_id UUID PRIMARY KEY,
			user_id UUID REFERENCES users(id) ON DELETE CASCADE,
			contact_email VARCHAR(255) NOT NULL,
			nickname VARCHAR(255),
			notification_preference VARCHAR(50),
			created_at TIMESTAMP DEFAULT NOW()
		);`,
		`CREATE TABLE IF NOT EXISTS habits (
			id UUID PRIMARY KEY,
			user_id UUID REFERENCES users(id) ON DELETE CASCADE,
			name VARCHAR(255) NOT NULL,
			goal_type VARCHAR(50),
			frequency VARCHAR(50),
			created_at TIMESTAMP DEFAULT NOW()
		);`,
		`CREATE TABLE IF NOT EXISTS habit_logs (
			id SERIAL PRIMARY KEY,
			habit_id UUID REFERENCES habits(id) ON DELETE CASCADE,
			user_id UUID REFERENCES users(id) ON DELETE CASCADE,
			value INTEGER NOT NULL,
			timestamp TIMESTAMP DEFAULT NOW()
		);`,
	}
	tx, err := c.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()
	for _, sql := range schemaSQL {
		if _, err = tx.Exec(ctx, sql); err != nil {
			return fmt.Errorf("falha ao executar SQL de inicialização: %w", err)
		}
	}
	return tx.Commit(ctx)
}

func (c *Client) CreateUser(ctx context.Context, user models.User) error {
	tx, err := c.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("falha ao iniciar transação: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()
	if user.ID == "" {
		user.ID = uuid.New().String()
	}
	sqlUser := `INSERT INTO users (id, email, name, theme) VALUES ($1, $2, $3, $4)`
	if _, err = tx.Exec(ctx, sqlUser, user.ID, user.Email, user.Name, user.Theme); err != nil {
		return fmt.Errorf("falha ao inserir usuário: %w", err)
	}
	sqlMana := `INSERT INTO user_mana (user_id, balance) VALUES ($1, 0)`
	if _, err = tx.Exec(ctx, sqlMana, user.ID); err != nil {
		return fmt.Errorf("falha ao inicializar saldo de mana: %w", err)
	}
	return tx.Commit(ctx)
}

func (c *Client) GetUserByID(ctx context.Context, userID string) (models.User, error) {
	user := models.User{}
	sql := `SELECT id, email, name, theme FROM users WHERE id = $1`
	err := c.pool.QueryRow(ctx, sql, userID).Scan(&user.ID, &user.Email, &user.Name, &user.Theme)
	if err != nil {
		return models.User{}, err
	}
	return user, nil
}

func (c *Client) UpdateUser(ctx context.Context, user models.User) error {
	sql := `UPDATE users SET name = $2, theme = $3 WHERE id = $1`
	cmdTag, err := c.pool.Exec(ctx, sql, user.ID, user.Name, user.Theme)
	if err != nil {
		return err
	}
	if cmdTag.RowsAffected() == 0 {
		return errors.New("usuário não encontrado")
	}
	return nil
}

func (c *Client) CreateSupportContact(ctx context.Context, contact models.SupportContact) error {
	sql := `INSERT INTO support_contacts (contact_id, user_id, contact_email, nickname, notification_preference) VALUES ($1, $2, $3, $4, $5)`
	if contact.ContactID == "" {
		contact.ContactID = uuid.New().String()
	}
	_, err := c.pool.Exec(ctx, sql, contact.ContactID, contact.UserID, contact.ContactEmail, contact.Nickname, contact.NotificationPreference)
	return err
}

func (c *Client) GetSupportContactsByUserID(ctx context.Context, userID string) ([]models.SupportContact, error) {
	sql := `SELECT contact_id, user_id, contact_email, nickname, notification_preference FROM support_contacts WHERE user_id = $1`
	rows, err := c.pool.Query(ctx, sql, userID)
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

func (c *Client) DeleteSupportContact(ctx context.Context, contactID string) error {
	sql := `DELETE FROM support_contacts WHERE contact_id = $1`
	cmdTag, err := c.pool.Exec(ctx, sql, contactID)
	if err != nil {
		return err
	}
	if cmdTag.RowsAffected() == 0 {
		return errors.New("contato não encontrado")
	}
	return nil
}

func (c *Client) CreateHabit(ctx context.Context, habit models.Habit) (string, error) {
	if habit.ID == "" {
		habit.ID = uuid.New().String()
	}
	sql := `INSERT INTO habits (id, user_id, name, goal_type, frequency) VALUES ($1, $2, $3, $4, $5)`
	_, err := c.pool.Exec(ctx, sql, habit.ID, habit.UserID, habit.Name, habit.GoalType, habit.Frequency)
	if err != nil {
		return "", err
	}
	return habit.ID, nil
}

func (c *Client) GetHabitsByUserID(ctx context.Context, userID string) ([]models.Habit, error) {
	sql := `SELECT id, user_id, name, goal_type, frequency, created_at FROM habits WHERE user_id = $1 ORDER BY created_at DESC`
	rows, err := c.pool.Query(ctx, sql, userID)
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

func (c *Client) LogHabit(ctx context.Context, logData models.HabitLog) error {
	sql := `INSERT INTO habit_logs (habit_id, user_id, value) VALUES ($1, $2, $3)`
	_, err := c.pool.Exec(ctx, sql, logData.HabitID, logData.UserID, logData.Value)
	if err != nil {
		return err
	}
	log.Printf("DB Ação: Log de hábito %s registrado.", logData.HabitID)
	return nil
}

func (c *Client) GetHabitLogs(ctx context.Context, habitID string) ([]models.HabitLog, error) {
	sql := `SELECT id, habit_id, user_id, value, timestamp FROM habit_logs WHERE habit_id = $1 ORDER BY timestamp DESC`
	rows, err := c.pool.Query(ctx, sql, habitID)
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

func (c *Client) GetManaBalance(ctx context.Context, userID string) (int, error) {
	var balance int
	sql := `SELECT balance FROM user_mana WHERE user_id = $1`
	err := c.pool.QueryRow(ctx, sql, userID).Scan(&balance)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, nil
		}
		return 0, err
	}
	return balance, nil
}

func (c *Client) UpdateManaBalance(ctx context.Context, txData models.ManaTransaction) error {
	tx, err := c.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("falha ao iniciar transação de mana: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()
	updateSQL := `UPDATE user_mana SET balance = balance + $1, updated_at = NOW() WHERE user_id = $2 RETURNING balance`
	var newBalance int
	err = tx.QueryRow(ctx, updateSQL, txData.Amount, txData.UserID).Scan(&newBalance)
	if err != nil {
		return fmt.Errorf("falha ao atualizar saldo: %w", err)
	}
	insertSQL := `INSERT INTO mana_transactions (user_id, type, amount, reference_id) VALUES ($1, $2, $3, $4)`
	if _, err := tx.Exec(ctx, insertSQL, txData.UserID, txData.Type, txData.Amount, txData.ReferenceID); err != nil {
		return fmt.Errorf("falha ao registrar transação de mana: %w", err)
	}
	return tx.Commit(ctx)
}

func (c *Client) CreateManaTransaction(ctx context.Context, tx models.ManaTransaction) error {
	return c.UpdateManaBalance(ctx, tx)
}
func (c *Client) GetTopManaUsers(ctx context.Context, limit int) ([]models.LeaderboardEntry, error) {
	sql := `
		SELECT u.id, u.name, um.balance
		FROM users u
		JOIN user_mana um ON u.id = um.user_id
		ORDER BY um.balance DESC
		LIMIT $1;
	`
	rows, err := c.pool.Query(ctx, sql, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []models.LeaderboardEntry
	for rows.Next() {
		var e models.LeaderboardEntry
		if err := rows.Scan(&e.UserID, &e.UserName, &e.Mana); err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}
	return entries, nil
}
