package models

import "time"

// User representa o perfil básico do usuário.
type User struct {
	ID        string    `json:"id,omitempty"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	Theme     string    `json:"theme,omitempty"` // Ex: "OutubroRosa", "Padrao"
	CreatedAt time.Time `json:"created_at,omitempty"`
	UpdatedAt time.Time `json:"updated_at,omitempty"`
}

// UserMana representa o saldo atual de Mana do usuário.
type UserMana struct {
	UserID    string    `json:"user_id"`
	Balance   int       `json:"balance"`              // Saldo atual de Mana
	UpdatedAt time.Time `json:"updated_at,omitempty"` // Pode ser setado pelo DB
}

// Tipagem para tipos de transações de Mana.
type ManaTransactionType string

const (
	ManaTypeHabitCompletion ManaTransactionType = "HABIT_COMPLETION"
	ManaTypeRewardRedeem    ManaTransactionType = "REWARD_REDEEM"
	ManaTypeActivityGrant   ManaTransactionType = "ACTIVITY_GRANT"
	ManaTypeChallengeDone   ManaTransactionType = "CHALLENGE_COMPLETE"
)

// ManaTransaction registra cada ganho ou perda de Mana.
type ManaTransaction struct {
	ID          string              `json:"id,omitempty"`
	UserID      string              `json:"user_id"`
	Type        ManaTransactionType `json:"type"`         // Use constantes acima
	Amount      int                 `json:"amount"`       // + ganho, - custo
	ReferenceID string              `json:"reference_id"` // ID do desafio, prêmio ou log de atividade
	CreatedAt   time.Time           `json:"created_at"`   // Definido no momento da gravação
}

// Habit representa um hábito/meta.
type Habit struct {
	ID        string    `json:"id,omitempty"`
	UserID    string    `json:"user_id,omitempty"`
	Name      string    `json:"name"`
	GoalType  string    `json:"goal_type,omitempty"` // Ex: "STEPS", "MEDICATION_LOG", "ACTIVITY"
	Frequency string    `json:"frequency,omitempty"` // Ex: "Daily", "Weekly"
	CreatedAt time.Time `json:"created_at,omitempty"`
}

// HabitLog registra o progresso de um hábito.
type HabitLog struct {
	ID        string    `json:"id,omitempty"`
	HabitID   string    `json:"habit_id"`
	UserID    string    `json:"user_id"`
	Timestamp time.Time `json:"log_date"` // era "timestamp"; trocado para compatibilidade do cliente
	Value     int       `json:"value"`    // Ex: número de passos ou 1 para concluído
}

// Challenge representa um desafio.
type Challenge struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	ManaReward  int    `json:"mana_reward"` // Quantidade de Mana que o usuário ganha
	GoalType    string `json:"goal_type"`   // Ex: "STEPS", "MEDICATION_LOG", "EDUCATIONAL_READ"
	GoalValue   int    `json:"goal_value"`  // Ex: 10000 passos, 7 dias de registro
	IsActive    bool   `json:"is_active"`
}

// Reward representa um prêmio resgatável.
type Reward struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Cost        int    `json:"cost"`         // Custo em Mana
	IsAvailable bool   `json:"is_available"` // Disponível para resgate
}

// LeaderboardEntry representa uma entrada no placar (geralmente no Redis).
type LeaderboardEntry struct {
	UserID   string `json:"user_id"`
	UserName string `json:"user_name"`
	Mana     int    `json:"mana"`
}

// SupportContact representa um membro da rede de apoio.
type SupportContact struct {
	UserID                 string `json:"user_id,omitempty"`
	ContactID              string `json:"contact_id,omitempty"` // usado para remoção
	Name                   string `json:"name,omitempty"`
	Phone                  string `json:"phone,omitempty"`
	ContactEmail           string `json:"contact_email,omitempty"`
	Nickname               string `json:"nickname,omitempty"`
	NotificationPreference string `json:"notification_preference,omitempty"`
}
