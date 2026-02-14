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

// --- Helpers para respostas padronizadas ---

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

// --- Handlers de API ---

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
		writeError(w, http.StatusUnauthorized, "Acesso negado.")
		return
	}

	habits, err := s.DBClient.GetHabitsByUserID(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Erro ao buscar hábitos.")
		return
	}

	writeJSON(w, http.StatusOK, habits)
}

// HandleGetHabitById busca um único hábito (Necessário para o cache do Angular).
func (s *Service) HandleGetHabitById(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	habitID := vars["habitId"]

	habit, err := s.DBClient.GetHabitById(r.Context(), habitID)
	if err != nil {
		writeError(w, http.StatusNotFound, "Hábito não encontrado.")
		return
	}

	writeJSON(w, http.StatusOK, habit)
}

// HandleLogHabit registra um progresso (Log) em um hábito.
func (s *Service) HandleLogHabit(w http.ResponseWriter, r *http.Request) {
	userID, err := auth.GetUserIDFromContext(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "Acesso negado.")
		return
	}

	vars := mux.Vars(r)
	habitID := vars["habitId"]

	var logData models.HabitLog
	_ = json.NewDecoder(r.Body).Decode(&logData)

	logData.UserID = userID
	logData.HabitID = habitID

	if err := s.DBClient.LogHabit(r.Context(), logData); err != nil {
		writeError(w, http.StatusInternalServerError, "Erro ao registrar log.")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "Log registrado."})
}

// HandleGetHabitLogs busca o histórico de um hábito.
func (s *Service) HandleGetHabitLogs(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	habitID := vars["habitId"]

	logs, err := s.DBClient.GetHabitLogs(r.Context(), habitID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Erro ao buscar histórico.")
		return
	}

	writeJSON(w, http.StatusOK, logs)
}

// HandleDeleteHabit remove o hábito com segurança.
func (s *Service) HandleDeleteHabit(w http.ResponseWriter, r *http.Request) {
	userID, err := auth.GetUserIDFromContext(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "Acesso negado.")
		return
	}

	vars := mux.Vars(r)
	habitID := vars["habitId"]

	if err := s.DBClient.DeleteHabit(r.Context(), habitID, userID); err != nil {
		writeError(w, http.StatusForbidden, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "Hábito excluído."})
}
