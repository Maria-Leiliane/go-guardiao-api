package models

import (
	"time"
)

// User representa o perfil básico do usuário.
type User struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
	Theme string `json:"theme"` // Ex: "OutubroRosa", "Padrão"
	// Adicionar outros campos conforme necessário (Ex: AvatarURL)
}

// UserMana representa o saldo atual de Mana do usuário.
// Idealmente armazenado no RDS (PostgreSQL) para integridade transacional.
type UserMana struct {
	UserID    string    `json:"user_id"`
	Balance   int       `json:"balance"` // Saldo atual de Mana
	UpdatedAt time.Time `json:"updated_at"`
}

// ManaTransaction registra cada ganho ou perda de Mana.
// Essencial para auditoria e histórico (armazenado no RDS).
type ManaTransaction struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	Type        string    `json:"type"`         // Ex: "CHALLENGE_COMPLETE", "ACTIVITY_GRANT", "REWARD_REDEEM"
	Amount      int       `json:"amount"`       // Valor da transação (+ para ganho, - para custo)
	ReferenceID string    `json:"reference_id"` // ID do desafio, prêmio ou log de atividade
	CreatedAt   time.Time `json:"created_at"`
}

// Habit representa um hábito ou meta que o usuário está monitorando.
type Habit struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Name      string    `json:"name"`
	GoalType  string    `json:"goal_type"` // Ex: "STEPS", "MEDICATION_LOG", "ACTIVITY"
	Frequency string    `json:"frequency"` // Ex: "Daily", "Weekly"
	CreatedAt time.Time `json:"created_at"`
}

// HabitLog registra o progresso diário de um hábito.
// Idealmente armazenado no DynamoDB devido ao alto volume de writes.
type HabitLog struct {
	ID        string    `json:"id"`
	HabitID   string    `json:"habit_id"`
	UserID    string    `json:"user_id"`
	Timestamp time.Time `json:"timestamp"`
	Value     int       `json:"value"` // Ex: número de passos ou 1 para concluído
}

// Challenge representa um desafio que pode ser completado.
// Pode ser armazenado no DynamoDB ou RDS.
type Challenge struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	ManaReward  int    `json:"mana_reward"` // Quantidade de Mana que o usuário ganha
	GoalType    string `json:"goal_type"`   // Ex: "STEPS", "MEDICATION_LOG", "EDUCATIONAL_READ"
	GoalValue   int    `json:"goal_value"`  // Ex: 10000 passos, 7 dias de registro
	IsActive    bool   `json:"is_active"`
}

// Reward representa um prêmio resgatável na "loja".
type Reward struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Cost        int    `json:"cost"` // Custo em Mana
	IsAvailable bool   `json:"is_available"`
}

// LeaderboardEntry representa uma entrada no placar de líderes (armazenado no ElastiCache Redis).
type LeaderboardEntry struct {
	UserID   string `json:"user_id"`
	UserName string `json:"user_name"`
	Mana     int    `json:"mana"`
}

// SupportContact representa um membro da rede de apoio do usuário.
type SupportContact struct {
	UserID                 string `json:"user_id"`
	ContactID              string `json:"contact_id"`    // Novo campo ID para remoção
	ContactEmail           string `json:"contact_email"` // Novo campo
	Nickname               string `json:"nickname"`      // Novo campo
	NotificationPreference string `json:"notification_preference"`
}
