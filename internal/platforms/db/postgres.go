package db

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
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

// ensureColumn adiciona uma coluna se ela não existir (compatível com Neon/Postgres)
func ensureColumn(ctx context.Context, tx pgx.Tx, table, column, definition string) error {
	var exists bool
	const q = `
		SELECT EXISTS (
			SELECT 1
			FROM information_schema.columns
			WHERE table_name = $1 AND column_name = $2
		);`
	if err := tx.QueryRow(ctx, q, table, column).Scan(&exists); err != nil {
		return fmt.Errorf("falha ao checar coluna %s.%s: %w", table, column, err)
	}
	if exists {
		return nil
	}
	stmt := fmt.Sprintf(`ALTER TABLE %s ADD COLUMN %s %s`, table, column, definition)
	if _, err := tx.Exec(ctx, stmt); err != nil {
		return fmt.Errorf("falha ao adicionar coluna %s.%s: %w", table, column, err)
	}
	return nil
}

func (c *Client) InitSchema(ctx context.Context) error {
	log.Println("Verificando e inicializando o esquema do banco de dados...")
	tx, err := c.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	// Tabela users
	if _, err = tx.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS users (
			id UUID PRIMARY KEY,
			email VARCHAR(255) UNIQUE NOT NULL,
			name VARCHAR(255),
			theme VARCHAR(50),
			created_at TIMESTAMP DEFAULT NOW()
		);`); err != nil {
		return fmt.Errorf("falha ao criar tabela users: %w", err)
	}
	// Coluna de senha (bcrypt ~60 chars; usamos VARCHAR(72))
	if err = ensureColumn(ctx, tx, "users", "password_hash", "VARCHAR(72)"); err != nil {
		return err
	}
	// Índice único (idempotente)
	if _, err = tx.Exec(ctx, `CREATE UNIQUE INDEX IF NOT EXISTS users_email_unique_idx ON users (email);`); err != nil {
		return fmt.Errorf("falha ao criar índice de email: %w", err)
	}

	// Demais tabelas
	if _, err = tx.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS user_mana (
			user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
			balance INTEGER NOT NULL DEFAULT 0,
			updated_at TIMESTAMP DEFAULT NOW()
		);`); err != nil {
		return fmt.Errorf("falha ao criar tabela user_mana: %w", err)
	}
	if _, err = tx.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS mana_transactions (
			id SERIAL PRIMARY KEY,
			user_id UUID REFERENCES users(id) ON DELETE CASCADE,
			type VARCHAR(50) NOT NULL,
			amount INTEGER NOT NULL,
			reference_id VARCHAR(255),
			created_at TIMESTAMP DEFAULT NOW()
		);`); err != nil {
		return fmt.Errorf("falha ao criar tabela mana_transactions: %w", err)
	}

	// Tabela support_contacts
	if _, err = tx.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS support_contacts (
			contact_id UUID PRIMARY KEY,
			user_id UUID REFERENCES users(id) ON DELETE CASCADE,
			contact_email VARCHAR(255) NOT NULL,
			nickname VARCHAR(255),
			notification_preference VARCHAR(50),
			created_at TIMESTAMP DEFAULT NOW()
		);`); err != nil {
		return fmt.Errorf("falha ao criar tabela support_contacts: %w", err)
	}
	// garante coluna phone (idempotente)
	if err = ensureColumn(ctx, tx, "support_contacts", "phone", "VARCHAR(30)"); err != nil {
		return err
	}

	if _, err = tx.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS habits (
			id UUID PRIMARY KEY,
			user_id UUID REFERENCES users(id) ON DELETE CASCADE,
			name VARCHAR(255) NOT NULL,
			goal_type VARCHAR(50),
			frequency VARCHAR(50),
			created_at TIMESTAMP DEFAULT NOW()
		);`); err != nil {
		return fmt.Errorf("falha ao criar tabela habits: %w", err)
	}
	if _, err = tx.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS habit_logs (
			id SERIAL PRIMARY KEY,
			habit_id UUID REFERENCES habits(id) ON DELETE CASCADE,
			user_id UUID REFERENCES users(id) ON DELETE CASCADE,
			value INTEGER NOT NULL,
			timestamp TIMESTAMP DEFAULT NOW()
		);`); err != nil {
		return fmt.Errorf("falha ao criar tabela habit_logs: %w", err)
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
	if strings.TrimSpace(user.ID) == "" {
		user.ID = uuid.New().String()
	}
	// Inclui password_hash no insert
	const sqlUser = `INSERT INTO users (id, email, name, theme, password_hash) VALUES ($1, $2, $3, $4, $5)`
	if _, err = tx.Exec(ctx, sqlUser, strings.ToLower(strings.TrimSpace(user.ID)), strings.ToLower(strings.TrimSpace(user.Email)), strings.TrimSpace(user.Name), strings.TrimSpace(user.Theme), user.PasswordHash); err != nil {
		return fmt.Errorf("falha ao inserir usuário: %w", err)
	}
	const sqlMana = `INSERT INTO user_mana (user_id, balance) VALUES ($1, 0)`
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

// Busca usuário por email (inclui password_hash) — necessário para login/registro
func (c *Client) GetUserByEmail(ctx context.Context, email string) (models.User, error) {
	u := models.User{}
	// Case-insensitive
	sql := `SELECT id, email, name, theme, password_hash FROM users WHERE LOWER(email) = LOWER($1) LIMIT 1`
	err := c.pool.QueryRow(ctx, sql, strings.TrimSpace(email)).Scan(&u.ID, &u.Email, &u.Name, &u.Theme, &u.PasswordHash)
	if err != nil {
		return models.User{}, err
	}
	return u, nil
}

// Obtém o hash de senha atual do usuário
func (c *Client) GetUserPasswordHash(ctx context.Context, userID string) (string, error) {
	var hash string
	const q = `SELECT password_hash FROM users WHERE id = $1`
	err := c.pool.QueryRow(ctx, q, userID).Scan(&hash)
	if err != nil {
		return "", err
	}
	return hash, nil
}

// Define/atualiza senha (hash) para usuário existente
func (c *Client) SetUserPassword(ctx context.Context, userID, hash string) error {
	const q = `UPDATE users SET password_hash = $2 WHERE id = $1`
	cmdTag, err := c.pool.Exec(ctx, q, userID, hash)
	if err != nil {
		return fmt.Errorf("falha ao atualizar senha: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return errors.New("usuário não encontrado para atualização de senha")
	}
	return nil
}

// Atualiza nome e tema do usuário
func (c *Client) UpdateUser(ctx context.Context, user models.User) error {
	sql := `UPDATE users SET name = $2, theme = $3 WHERE id = $1`
	cmdTag, err := c.pool.Exec(ctx, sql, user.ID, strings.TrimSpace(user.Name), strings.TrimSpace(user.Theme))
	if err != nil {
		return err
	}
	if cmdTag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

// Atualiza e-mail do usuário (deve existir índice único para evitar duplicidade)
func (c *Client) UpdateUserEmail(ctx context.Context, userID, email string) error {
	const q = `UPDATE users SET email = $2 WHERE id = $1`
	cmdTag, err := c.pool.Exec(ctx, q, userID, strings.ToLower(strings.TrimSpace(email)))
	if err != nil {
		return fmt.Errorf("falha ao atualizar e-mail: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (c *Client) CreateSupportContact(ctx context.Context, contact models.SupportContact) error {
	// Inclui phone no INSERT
	sql := `INSERT INTO support_contacts (contact_id, user_id, contact_email, phone, nickname, notification_preference)
	        VALUES ($1, $2, $3, $4, $5, $6)`
	if strings.TrimSpace(contact.ContactID) == "" {
		contact.ContactID = uuid.New().String()
	}
	_, err := c.pool.Exec(ctx, sql,
		contact.ContactID,
		contact.UserID,
		strings.TrimSpace(contact.ContactEmail),
		strings.TrimSpace(contact.Phone),
		strings.TrimSpace(contact.Nickname),
		strings.TrimSpace(contact.NotificationPreference),
	)
	return err
}

func (c *Client) GetSupportContactsByUserID(ctx context.Context, userID string) ([]models.SupportContact, error) {
	// Seleciona phone
	sql := `SELECT contact_id, user_id, contact_email, phone, nickname, notification_preference FROM support_contacts WHERE user_id = $1`
	rows, err := c.pool.Query(ctx, sql, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var contacts []models.SupportContact
	for rows.Next() {
		contact := models.SupportContact{}
		if err := rows.Scan(
			&contact.ContactID,
			&contact.UserID,
			&contact.ContactEmail,
			&contact.Phone,
			&contact.Nickname,
			&contact.NotificationPreference,
		); err != nil {
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
	if strings.TrimSpace(habit.ID) == "" {
		habit.ID = uuid.New().String()
	}
	sql := `INSERT INTO habits (id, user_id, name, goal_type, frequency) VALUES ($1, $2, $3, $4, $5)`
	_, err := c.pool.Exec(ctx, sql, habit.ID, habit.UserID, strings.TrimSpace(habit.Name), strings.TrimSpace(habit.GoalType), strings.TrimSpace(habit.Frequency))
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

// Remove contato de suporte garantindo que pertence ao userID
func (c *Client) DeleteSupportContactByUser(ctx context.Context, userID, contactID string) error {
	const sql = `DELETE FROM support_contacts WHERE contact_id = $1 AND user_id = $2`
	cmdTag, err := c.pool.Exec(ctx, sql, contactID, userID)
	if err != nil {
		return err
	}
	if cmdTag.RowsAffected() == 0 {
		return errors.New("contato não encontrado")
	}
	return nil
}
