package aws

import (
	"context"
	"errors"
	"sync"
)

// MockQueue é uma implementação em memória de QueueClient para DEV/TEST.
type MockQueue struct {
	mu       sync.Mutex
	queue    []QueueMessage
	nextID   int
	closed   bool
	queueURL string
}

func NewMockQueue(queueURL string) *MockQueue {
	return &MockQueue{queueURL: queueURL}
}

func (m *MockQueue) Send(_ context.Context, body string, attrs map[string]string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.closed {
		return "", errors.New("mock queue closed")
	}
	m.nextID++
	id := m.queueURL + "#msg-" + itoa(m.nextID)
	m.queue = append(m.queue, QueueMessage{
		MessageID:     id,
		Body:          body,
		ReceiptHandle: id + "-rh",
		Attributes:    attrs,
	})
	return id, nil
}

func (m *MockQueue) Receive(_ context.Context, max int32, _ int32, _ int32) ([]QueueMessage, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.queue) == 0 {
		return nil, nil
	}
	if max <= 0 || int(max) > len(m.queue) {
		max = int32(len(m.queue))
	}
	out := make([]QueueMessage, max)
	copy(out, m.queue[:max])
	return out, nil
}

func (m *MockQueue) Delete(_ context.Context, receiptHandle string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for i, msg := range m.queue {
		if msg.ReceiptHandle == receiptHandle {
			m.queue = append(m.queue[:i], m.queue[i+1:]...)
			return nil
		}
	}
	return errors.New("receipt handle not found")
}

func (m *MockQueue) Purge(_ context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.queue = nil
	return nil
}

func itoa(n int) string {
	// simples conversão para evitar dependências
	if n == 0 {
		return "0"
	}
	var d []byte
	for n > 0 {
		d = append([]byte{byte('0' + n%10)}, d...)
		n /= 10
	}
	return string(d)
}
