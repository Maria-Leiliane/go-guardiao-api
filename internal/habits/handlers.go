package habits

import (
	"encoding/json"
	"fmt"
	"net/http"

	"go-guardiao-api/internal/auth"
	"go-guardiao-api/internal/platforms/db"
	"go-guardiao-api/pkg/models"

	"github.com/gorilla/mux"
)

// Service representa o serviço de Hábitos, contendo a dependência do DB.
type Service struct {
	DBClient *db.Client
}

// NewService cria uma nova instância do serviço de Hábitos.
func NewService(dbClient *db.Client) *Service {
	return &Service{DBClient: dbClient}
}

// HandleCreateHabit lida com a criação de um novo hábito.
func (s *Service) HandleCreateHabit(w http.ResponseWriter, r *http.Request) {
	userID, err := auth.GetUserIDFromContext(r)
	if err != nil {
		http.Error(w, "Acesso negado: UserID ausente.", http.StatusUnauthorized)
		return
	}

	var newHabit models.Habit
	if err := json.NewDecoder(r.Body).Decode(&newHabit); err != nil {
		http.Error(w, "Requisição inválida.", http.StatusBadRequest)
		return
	}

	newHabit.UserID = userID
	habitID, err := s.DBClient.CreateHabit(r.Context(), newHabit)
	if err != nil {
		http.Error(w, fmt.Sprintf("Falha ao criar hábito: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(map[string]string{
		"message":  "Hábito criado com sucesso.",
		"habit_id": habitID,
	}); err != nil {
		http.Error(w, "Erro ao serializar resposta.", http.StatusInternalServerError)
	}
}

// HandleGetHabits busca todos os hábitos de um usuário.
func (s *Service) HandleGetHabits(w http.ResponseWriter, r *http.Request) {
	userID, err := auth.GetUserIDFromContext(r)
	if err != nil {
		http.Error(w, "Acesso negado: UserID ausente.", http.StatusUnauthorized)
		return
	}

	habits, err := s.DBClient.GetHabitsByUserID(r.Context(), userID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Falha ao buscar hábitos: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(habits); err != nil {
		http.Error(w, "Erro ao serializar resposta.", http.StatusInternalServerError)
	}
}

// HandleLogHabit lida com o registro de um progresso em um hábito.
func (s *Service) HandleLogHabit(w http.ResponseWriter, r *http.Request) {
	userID, err := auth.GetUserIDFromContext(r)
	if err != nil {
		http.Error(w, "Acesso negado: UserID ausente.", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	habitID := vars["habitId"]

	var logData models.HabitLog
	if err := json.NewDecoder(r.Body).Decode(&logData); err != nil {
		http.Error(w, "Requisição inválida.", http.StatusBadRequest)
		return
	}

	logData.UserID = userID
	logData.HabitID = habitID

	if err := s.DBClient.LogHabit(r.Context(), logData); err != nil {
		http.Error(w, fmt.Sprintf("Falha ao registrar log de hábito: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(map[string]string{
		"message": "Log de hábito registrado com sucesso.",
	}); err != nil {
		http.Error(w, "Erro ao serializar resposta.", http.StatusInternalServerError)
	}
}

// HandleGetHabitLogs busca os logs de um hábito específico.
func (s *Service) HandleGetHabitLogs(w http.ResponseWriter, r *http.Request) {
	// A variável userID não é usada aqui, mas é extraída para garantir a
	// consistência de que a rota é protegida e o usuário está autenticado.
	// O erro está sendo tratado, então o warning 'unused variable' é irrelevante,
	// mas pode ser ignorado no Go com o '_' se fosse necessário.
	_, err := auth.GetUserIDFromContext(r)
	if err != nil {
		http.Error(w, "Acesso negado: UserID ausente.", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	habitID := vars["habitId"]

	logs, err := s.DBClient.GetHabitLogs(r.Context(), habitID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Falha ao buscar logs do hábito: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(logs); err != nil {
		http.Error(w, "Erro ao serializar resposta.", http.StatusInternalServerError)
	}
}
