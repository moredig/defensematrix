package proxy

import (
	"fmt"
	"sync"
	"time"
)

// ConnectionState tracks an individual connection through the gateway
type ConnectionState int

const (
	ConnClean      ConnectionState = iota // Normal traffic
	ConnSuspicious                        // Flagged but not confirmed
	ConnMalicious                         // Confirmed attack
)

// Connection represents a tracked network connection
type Connection struct {
	ID        string
	SourceIP  string
	State     ConnectionState
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Multiplexer tracks all active connections and manages routing state
type Multiplexer struct {
	mu          sync.RWMutex
	connections map[string]*Connection
	gateway     *Gateway
	onNuke      func() // Fired when malicious connection confirmed
}

// NewMultiplexer creates a new connection tracker
func NewMultiplexer(gateway *Gateway, onNuke func()) *Multiplexer {
	return &Multiplexer{
		connections: make(map[string]*Connection),
		gateway:     gateway,
		onNuke:      onNuke,
	}
}

// Track registers a new incoming connection
func (m *Multiplexer) Track(id string, sourceIP string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.connections[id] = &Connection{
		ID:        id,
		SourceIP:  sourceIP,
		State:     ConnClean,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	fmt.Printf("[WAMAI] Connection tracked: %s from %s\n", id, sourceIP)
}

// Flag marks a connection as suspicious or malicious
func (m *Multiplexer) Flag(id string, state ConnectionState) {
	m.mu.Lock()
	defer m.mu.Unlock()

	conn, exists := m.connections[id]
	if !exists {
		return
	}

	conn.State = state
	conn.UpdatedAt = time.Now()

	switch state {
	case ConnSuspicious:
		fmt.Printf("[WAMAI] Connection suspicious: %s (%s)\n", id, conn.SourceIP)
	case ConnMalicious:
		fmt.Printf("[WAMAI] Connection malicious: %s (%s) — triggering nuke.\n", id, conn.SourceIP)
		if m.onNuke != nil {
			go m.onNuke()
		}
	}
}

// Drop removes a connection from the tracker
func (m *Multiplexer) Drop(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.connections, id)
	fmt.Printf("[WAMAI] Connection dropped: %s\n", id)
}

// ActiveCount returns the number of currently tracked connections
func (m *Multiplexer) ActiveCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.connections)
}

// GetAll returns a snapshot of all active connections
func (m *Multiplexer) GetAll() []*Connection {
	m.mu.RLock()
	defer m.mu.RUnlock()

	conns := make([]*Connection, 0, len(m.connections))
	for _, c := range m.connections {
		conns = append(conns, c)
	}
	return conns
}
