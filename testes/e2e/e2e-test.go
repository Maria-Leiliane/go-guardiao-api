package e2e

import (
	"fmt"
	"net/http"
	"testing"
	"time"
)

func TestHealth(t *testing.T) {
	c := NewClient()

	// Aguarda o /health ficar OK com backoff (evita flakiness em CI)
	if err := Retry(30, 2*time.Second, func() error {
		res, body, err := c.do(http.MethodGet, "/health", nil, false, nil)
		if err != nil {
			return err
		}
		if res.StatusCode != 200 || string(body) != "ok" {
			return fmt.Errorf("health unexpected: status=%d body=%s", res.StatusCode, string(body))
		}
		return nil
	}); err != nil {
		t.Fatalf("health check failed: %v", err)
	}
}

func TestAuthRegisterLoginAndProtected(t *testing.T) {
	c := NewClient()

	// 1) Register
	email := fmt.Sprintf("e2e+%d@example.com", time.Now().UnixNano())
	reg := map[string]any{
		"name":     "E2E User",
		"email":    email,
		"password": "123456",
	}
	res, b, err := c.post("/auth/register", reg, false)
	if err != nil {
		t.Fatalf("register: %v", err)
	}
	if res.StatusCode != 200 && res.StatusCode != 201 && res.StatusCode != 409 {
		t.Fatalf("register unexpected status=%d body=%s", res.StatusCode, string(b))
	}

	// 2) Login
	login := map[string]any{"email": email, "password": "123456"}
	res, b, err = c.post("/auth/login", login, false)
	if err != nil {
		t.Fatalf("login: %v", err)
	}
	MustStatus(res, 200, b)

	var loginResp map[string]any
	if err := decode(b, &loginResp); err != nil {
		t.Fatalf("login decode: %v", err)
	}
	token, _ := loginResp["token"].(string)
	if token == "" {
		t.Fatalf("missing token in login response: %s", string(b))
	}
	c.WithToken(token)

	// 3) Protected without token -> 401
	cNoAuth := NewClient()
	res, b, err = cNoAuth.get("/user/profile", true)
	if err != nil {
		t.Fatalf("profile without token: %v", err)
	}
	if res.StatusCode != 401 {
		t.Fatalf("expected 401 for profile without token, got %d body=%s", res.StatusCode, string(b))
	}

	// 4) Protected with token -> 200
	res, b, err = c.get("/user/profile", true)
	if err != nil {
		t.Fatalf("profile with token: %v", err)
	}
	MustStatus(res, 200, b)
}

func TestUserProfileUpdate(t *testing.T) {
	c := NewClient()

	// cria usuário e autentica
	email := fmt.Sprintf("e2e+%d@example.com", time.Now().UnixNano())
	_, _, _ = c.post("/auth/register", map[string]any{
		"name":     "E2E User",
		"email":    email,
		"password": "123456",
	}, false)

	res, b, err := c.post("/auth/login", map[string]any{"email": email, "password": "123456"}, false)
	if err != nil {
		t.Fatalf("login: %v", err)
	}
	MustStatus(res, 200, b)

	var lr map[string]any
	_ = decode(b, &lr)
	token := lr["token"].(string)
	c.WithToken(token)

	// Update Profile
	payload := map[string]any{"name": "E2E Updated", "email": email}
	res, b, err = c.put("/user/profile", payload, true)
	if err != nil {
		t.Fatalf("update profile: %v", err)
	}
	MustStatus(res, 200, b)
}

func TestHabitsCRUDAndLogs(t *testing.T) {
	c := NewClient()

	// auth
	email := fmt.Sprintf("e2e+%d@example.com", time.Now().UnixNano())
	_, _, _ = c.post("/auth/register", map[string]any{
		"name":     "E2E Habits",
		"email":    email,
		"password": "123456",
	}, false)

	res, b, err := c.post("/auth/login", map[string]any{"email": email, "password": "123456"}, false)
	if err != nil {
		t.Fatalf("login: %v", err)
	}
	MustStatus(res, 200, b)

	var lr map[string]any
	_ = decode(b, &lr)
	c.WithToken(lr["token"].(string))

	// Create Habit
	create := map[string]any{"name": "Beber agua", "description": "2L/dia"}
	res, b, err = c.post("/habits", create, true)
	if err != nil {
		t.Fatalf("create habit: %v", err)
	}
	MustStatus(res, 201, b)

	var cr map[string]any
	_ = decode(b, &cr)
	habitID, _ := cr["habit_id"].(string)
	if habitID == "" {
		t.Fatalf("missing habit_id in response: %s", string(b))
	}

	// List Habits
	res, b, err = c.get("/habits", true)
	if err != nil {
		t.Fatalf("list habits: %v", err)
	}
	MustStatus(res, 200, b)

	// Log Habit (rota corrigida /habits/{habitId}/log)
	logBody := map[string]any{
		"value":    1,
		"log_date": time.Now().UTC().Format(time.RFC3339),
	}
	res, b, err = c.post(fmt.Sprintf("/habits/%s/log", habitID), logBody, true)
	if err != nil {
		t.Fatalf("log habit: %v", err)
	}
	MustStatus(res, 200, b)

	// Get Habit Logs
	res, b, err = c.get(fmt.Sprintf("/habits/%s/logs", habitID), true)
	if err != nil {
		t.Fatalf("get habit logs: %v", err)
	}
	MustStatus(res, 200, b)
}

func TestErrorCases(t *testing.T) {
	c := NewClient()

	// Sem token => 401
	res, b, err := c.get("/user/profile", true)
	if err != nil {
		t.Fatalf("unauthorized profile: %v", err)
	}
	if res.StatusCode != 401 {
		t.Fatalf("expected 401, got %d body=%s", res.StatusCode, string(b))
	}

	// JSON inválido no PUT profile => 400 (autenticado)
	email := fmt.Sprintf("e2e+%d@example.com", time.Now().UnixNano())
	_, _, _ = c.post("/auth/register", map[string]any{"name": "E2E Err", "email": email, "password": "123456"}, false)

	resLogin, bodyLogin, err := c.post("/auth/login", map[string]any{"email": email, "password": "123456"}, false)
	if err != nil {
		t.Fatalf("login: %v", err)
	}
	MustStatus(resLogin, 200, bodyLogin)

	var lr map[string]any
	_ = decode(bodyLogin, &lr)
	token := lr["token"].(string)

	req, _ := http.NewRequest(http.MethodPut, c.BaseURL+"/user/profile", bytesReader([]byte("{ name: sem-aspas }")))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.HTTP.Do(req)
	if err != nil {
		t.Fatalf("bad json request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 400 && resp.StatusCode != 422 {
		t.Fatalf("expected 400/422 for bad json, got %d", resp.StatusCode)
	}
}
