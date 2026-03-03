package terminal

import (
	"fmt"
	"sync"
	"time"
)

type Manager struct {
	mu        sync.RWMutex
	terminals map[string]Terminal
}

func NewManager() *Manager {
	return &Manager{
		terminals: make(map[string]Terminal),
	}
}

func (m *Manager) Add(id string, term Terminal) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.terminals[id] = term
}

func (m *Manager) Get(id string) (Terminal, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	t, ok := m.terminals[id]
	return t, ok
}

func (m *Manager) Remove(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if t, ok := m.terminals[id]; ok {
		t.Close()
		delete(m.terminals, id)
	}
}

// GenerateID creates a simple unique ID
func GenerateID() string {
	return fmt.Sprintf("term_%d", time.Now().UnixNano())
}
