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

// Service representa o serviço de Hábitos.
type Service struct {
	DBClient *db.Client
}

func NewService(dbClient *db.Client) *Service {
	return &Service{DBClient: dbClient}
}

// Helpers para respostas padronizadas
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

// HandleCreateHabit lida com a criação de um novo hábito.
func (s *Service) HandleCreateHabit(w http.ResponseWriter, r *http.Request) {
	userID, err := auth.GetUserIDFromContext(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "Acesso negado: UserID ausente.")
		return
	}

	var newHabit models.Habit
	if err := json.NewDecoder(r.Body).Decode(&newHabit); err != nil {
		writeError(w, http.StatusBadRequest, "Requisição inválida.")
		return
	}

	newHabit.UserID = userID
	habitID, err := s.DBClient.CreateHabit(r.Context(), newHabit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("Falha ao criar hábito: %v", err))
		return
	}

	writeJSON(w, http.StatusCreated, map[string]string{
		"message":  "Hábito criado com sucesso.",
		"habit_id": habitID,
	})
}

// HandleGetHabits busca todos os hábitos de um usuário.
func (s *Service) HandleGetHabits(w http.ResponseWriter, r *http.Request) {
	userID, err := auth.GetUserIDFromContext(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "Acesso negado: UserID ausente.")
		return
	}

	habits, err := s.DBClient.GetHabitsByUserID(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("Falha ao buscar hábitos: %v", err))
		return
	}

	writeJSON(w, http.StatusOK, habits)
}

// HandleLogHabit lida com o registro de um progresso em um hábito.
func (s *Service) HandleLogHabit(w http.ResponseWriter, r *http.Request) {
	userID, err := auth.GetUserIDFromContext(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "Acesso negado: UserID ausente.")
		return
	}

	vars := mux.Vars(r)
	habitID := vars["habitId"]
	if habitID == "" {
		writeError(w, http.StatusBadRequest, "Parâmetro habitId ausente na rota.")
		return
	}

	var logData models.HabitLog
	if err := json.NewDecoder(r.Body).Decode(&logData); err != nil {
		writeError(w, http.StatusBadRequest, "Requisição inválida.")
		return
	}

	logData.UserID = userID
	logData.HabitID = habitID

	if err := s.DBClient.LogHabit(r.Context(), logData); err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("Falha ao registrar log de hábito: %v", err))
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"message": "Log de hábito registrado com sucesso.",
	})
}

// HandleGetHabitLogs busca os logs de um hábito específico.
func (s *Service) HandleGetHabitLogs(w http.ResponseWriter, r *http.Request) {
	_, err := auth.GetUserIDFromContext(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "Acesso negado: UserID ausente.")
		return
	}

	vars := mux.Vars(r)
	habitID := vars["habitId"]
	if habitID == "" {
		writeError(w, http.StatusBadRequest, "Parâmetro habitId ausente na rota.")
		return
	}

	logs, err := s.DBClient.GetHabitLogs(r.Context(), habitID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("Falha ao buscar logs do hábito: %v", err))
		return
	}

	writeJSON(w, http.StatusOK, logs)
}
