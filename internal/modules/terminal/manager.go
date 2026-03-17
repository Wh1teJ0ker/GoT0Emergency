// Package terminal provides terminal session management
package terminal

import (
	"GoT0Emergency/internal/pkg/log"
	"sync"
	"time"
)

// Manager manages multiple terminal sessions
type Manager struct {
	mu        sync.RWMutex
	terminals map[string]Terminal // Active terminal sessions by ID
}

// NewManager creates a new terminal manager instance
func NewManager() *Manager {
	return &Manager{
		terminals: make(map[string]Terminal),
	}
}

// Add adds a terminal session to the manager
// id: unique terminal identifier
// term: Terminal implementation instance
func (m *Manager) Add(id string, term Terminal) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.terminals[id] = term
}

// Get retrieves a terminal session by ID
// id: unique terminal identifier
// Returns: Terminal instance and boolean indicating if found
func (m *Manager) Get(id string) (Terminal, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	t, ok := m.terminals[id]
	return t, ok
}

// Remove removes and closes a terminal session
// id: unique terminal identifier
func (m *Manager) Remove(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if t, ok := m.terminals[id]; ok {
		t.Close()
		delete(m.terminals, id)
	}
}

// GenerateID creates a unique terminal identifier based on timestamp
func GenerateID() string {
	return log.Sprintf("term_%d", time.Now().UnixNano())
}
