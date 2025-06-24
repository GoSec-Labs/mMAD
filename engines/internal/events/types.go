package events

import (
	"time"

	"github.com/GoSec-Labs/mMAD/engines/internal/zkproof/types"
)

// EventType represents different types of events
type EventType string

const (
	// Proof lifecycle events
	EventProofRequested EventType = "proof.requested"
	EventProofGenerated EventType = "proof.generated"
	EventProofVerified  EventType = "proof.verified"
	EventProofFailed    EventType = "proof.failed"

	// Circuit events
	EventCircuitCompiled EventType = "circuit.compiled"
	EventCircuitCached   EventType = "circuit.cached"

	// System events
	EventSystemStarted EventType = "system.started"
	EventSystemStopped EventType = "system.stopped"
	EventError         EventType = "system.error"
)

// Event represents a system event
type Event struct {
	ID        string                 `json:"id"`
	Type      EventType              `json:"type"`
	Timestamp time.Time              `json:"timestamp"`
	Source    string                 `json:"source"`
	Data      map[string]interface{} `json:"data"`
	UserID    string                 `json:"user_id,omitempty"`
	RequestID string                 `json:"request_id,omitempty"`
}

// ProofEvent contains proof-specific event data
type ProofEvent struct {
	ProofID   string          `json:"proof_id"`
	ProofType types.ProofType `json:"proof_type"`
	Status    string          `json:"status"`
	Duration  time.Duration   `json:"duration,omitempty"`
	Error     string          `json:"error,omitempty"`
	ProofSize int             `json:"proof_size,omitempty"`
}

// CircuitEvent contains circuit-specific event data
type CircuitEvent struct {
	CircuitID   string        `json:"circuit_id"`
	CircuitName string        `json:"circuit_name"`
	Constraints int           `json:"constraints"`
	CompileTime time.Duration `json:"compile_time,omitempty"`
}

// EventHandler processes events
type EventHandler interface {
	Handle(event Event) error
	CanHandle(eventType EventType) bool
}

// EventPublisher publishes events
type EventPublisher interface {
	Publish(event Event) error
	PublishAsync(event Event)
}

// EventSubscriber subscribes to events
type EventSubscriber interface {
	Subscribe(eventType EventType, handler EventHandler) error
	Unsubscribe(eventType EventType, handler EventHandler) error
}
