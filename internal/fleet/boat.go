package fleet

import (
	"fmt"
	"sync"
	"time"
)

// BoatState represents the current state of a Boat in the fleet
type BoatState int

const (
	StateIdle        BoatState = iota // Blue — clean, broadcasting bait
	StateProbing                      // Yellow — being scanned/fuzzed
	StateInfiltrated                  // Orange — attacker inside, logging keystrokes
	StateNuked                        // Red — kill switch triggered, recycling
	StateStopped                      // Gray — paused by operator
)

func (s BoatState) String() string {
	switch s {
	case StateIdle:
		return "IDLE"
	case StateProbing:
		return "PROBING"
	case StateInfiltrated:
		return "INFILTRATED"
	case StateNuked:
		return "NUKED"
	case StateStopped:
		return "STOPPED"
	default:
		return "UNKNOWN"
	}
}

// Boat represents a single micro-isolated container in the fleet
type Boat struct {
	mu          sync.Mutex
	ID          string    // Docker container ID
	Name        string    // Human-readable name e.g. "boat-1"
	State       BoatState // Current state
	Role        string    // "frontline", "shadow", "graveyard"
	CreatedAt   time.Time
	TriggeredAt *time.Time // When the kill switch fired
}

// NewBoat creates a new Boat instance
func NewBoat(name string, role string) *Boat {
	return &Boat{
		Name:      name,
		Role:      role,
		State:     StateIdle,
		CreatedAt: time.Now(),
	}
}

// SetState safely transitions a Boat to a new state
func (b *Boat) SetState(state BoatState) {
	b.mu.Lock()
	defer b.mu.Unlock()

	fmt.Printf("[WAMAI] %s: %s → %s\n", b.Name, b.State, state)
	b.State = state

	if state == StateNuked {
		now := time.Now()
		b.TriggeredAt = &now
	}
}

// GetState safely returns the current state
func (b *Boat) GetState() BoatState {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.State
}

// IsActive returns true if the boat is frontline
func (b *Boat) IsActive() bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.Role == "frontline"
}

// Promote elevates a shadow boat to frontline
func (b *Boat) Promote() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.Role = "frontline"
	b.State = StateIdle
	fmt.Printf("[WAMAI] %s promoted to frontline.\n", b.Name)
}

// Demote sends a boat to the graveyard
func (b *Boat) Demote() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.Role = "graveyard"
	fmt.Printf("[WAMAI] %s sent to graveyard.\n", b.Name)
}
