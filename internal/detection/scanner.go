package detection

import (
	"fmt"
	"regexp"
	"sync"
	"time"
)

// ThreatLevel represents how dangerous the detected pattern is
type ThreatLevel int

const (
	ThreatLow    ThreatLevel = iota // Mild probing
	ThreatMedium                    // Active fuzzing
	ThreatHigh                      // Confirmed attack payload
)

func (t ThreatLevel) String() string {
	switch t {
	case ThreatLow:
		return "LOW"
	case ThreatMedium:
		return "MEDIUM"
	case ThreatHigh:
		return "HIGH"
	default:
		return "UNKNOWN"
	}
}

// Hit represents a single detected threat event
type Hit struct {
	Pattern   string
	Payload   string
	Level     ThreatLevel
	Timestamp time.Time
}

// Scanner watches raw traffic streams for malicious patterns
type Scanner struct {
	mu         sync.Mutex
	signatures []*Signature
	hits       []Hit
	threshold  int // Number of hits before triggering nuke
	hitCount   int
	OnTrigger  func() // Callback fired when threshold is breached
}

// NewScanner initialises a Scanner with loaded signatures
func NewScanner(threshold int, onTrigger func()) *Scanner {
	return &Scanner{
		signatures: LoadSignatures(),
		threshold:  threshold,
		OnTrigger:  onTrigger,
	}
}

// Scan checks a raw payload string against all signatures
func (s *Scanner) Scan(payload string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, sig := range s.signatures {
		if sig.Regex.MatchString(payload) {
			hit := Hit{
				Pattern:   sig.Name,
				Payload:   truncate(payload, 80),
				Level:     sig.Level,
				Timestamp: time.Now(),
			}
			s.hits = append(s.hits, hit)
			s.hitCount++

			fmt.Printf("[WAMAI] HIT — Pattern: %s | Level: %s | Payload: %s\n",
				hit.Pattern, hit.Level, hit.Payload)

			// Trigger nuke if threshold breached or high threat detected
			if s.hitCount >= s.threshold || sig.Level == ThreatHigh {
				fmt.Println("[WAMAI] Threshold breached — triggering nuke.")
				if s.OnTrigger != nil {
					go s.OnTrigger()
				}
				s.hitCount = 0 // Reset after trigger
			}
			return
		}
	}
}

// GetHits returns all recorded hits
func (s *Scanner) GetHits() []Hit {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.hits
}

// truncate shortens a string for clean logging
func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}

// Signature defines a single detection rule
type Signature struct {
	Name  string
	Regex *regexp.Regexp
	Level ThreatLevel
}
