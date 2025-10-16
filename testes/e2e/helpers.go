package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type Client struct {
	BaseURL string
	Token   string
	HTTP    HTTPClient
}

func NewClient() *Client {
	base := os.Getenv("E2E_BASE_URL")
	if strings.TrimSpace(base) == "" {
		base = "http://localhost:18080/api/v1"
	}
	return &Client{
		BaseURL: strings.TrimRight(base, "/"),
		HTTP:    &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *Client) WithToken(token string) *Client {
	c.Token = token
	return c
}

func (c *Client) do(method, path string, body any, auth bool, headers map[string]string) (*http.Response, []byte, error) {
	var rdr io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, nil, fmt.Errorf("marshal body: %w", err)
		}
		rdr = bytes.NewReader(b)
	}

	req, err := http.NewRequest(method, c.BaseURL+path, rdr)
	if err != nil {
		return nil, nil, fmt.Errorf("new request: %w", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if auth && c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	res, err := c.HTTP.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("do: %w", err)
	}

	// Lê o corpo antes de fechar e trata erros de leitura e de fechamento.
	b, readErr := io.ReadAll(res.Body)
	closeErr := res.Body.Close()
	if readErr != nil {
		return res, nil, fmt.Errorf("read body: %w", readErr)
	}
	if closeErr != nil {
		// Preferimos retornar o body válido e reportar erro no close
		return res, b, fmt.Errorf("close body: %w", closeErr)
	}
	return res, b, nil
}

func (c *Client) post(path string, body any, auth bool) (*http.Response, []byte, error) {
	return c.do(http.MethodPost, path, body, auth, nil)
}
func (c *Client) get(path string, auth bool) (*http.Response, []byte, error) {
	return c.do(http.MethodGet, path, nil, auth, nil)
}
func (c *Client) put(path string, body any, auth bool) (*http.Response, []byte, error) {
	return c.do(http.MethodPut, path, body, auth, nil)
}

func MustStatus(res *http.Response, want int, body []byte) {
	if res.StatusCode != want {
		panic(fmt.Errorf("unexpected status: got=%d want=%d body=%s", res.StatusCode, want, string(body)))
	}
}

func decode[T any](b []byte, into *T) error {
	return json.Unmarshal(b, into)
}

// Retry com backoff simples para esperar efeitos eventualmente consistentes
func Retry(attempts int, sleep time.Duration, fn func() error) error {
	var err error
	for i := 0; i < attempts; i++ {
		err = fn()
		if err == nil {
			return nil
		}
		time.Sleep(sleep)
	}
	return err
}

// bytesReader é usado apenas nos testes para enviar JSON inválido bruto
func bytesReader(b []byte) *bytes.Reader { return bytes.NewReader(b) }
