package events

import (
	"time"

	"github.com/GoSec-Labs/mMAD/engines/internal/zkproof/types"
)

// EventEmitter provides helper methods to emit common events
type EventEmitter struct {
	bus    EventPublisher
	source string
}

// NewEventEmitter creates a new event emitter
func NewEventEmitter(bus EventPublisher, source string) *EventEmitter {
	return &EventEmitter{
		bus:    bus,
		source: source,
	}
}

// EmitProofRequested emits a proof requested event
func (e *EventEmitter) EmitProofRequested(proofID string, proofType types.ProofType, userID, requestID string) {
	event := Event{
		Type:      EventProofRequested,
		Source:    e.source,
		UserID:    userID,
		RequestID: requestID,
		Data: map[string]interface{}{
			"proof_id":   proofID,
			"proof_type": proofType,
		},
	}

	e.bus.PublishAsync(event)
}

// EmitProofGenerated emits a proof generated event
func (e *EventEmitter) EmitProofGenerated(proofID string, proofType types.ProofType, duration time.Duration, proofSize int, userID, requestID string) {
	event := Event{
		Type:      EventProofGenerated,
		Source:    e.source,
		UserID:    userID,
		RequestID: requestID,
		Data: map[string]interface{}{
			"proof_id":   proofID,
			"proof_type": proofType,
			"duration":   duration.String(),
			"proof_size": proofSize,
		},
	}

	e.bus.PublishAsync(event)
}

// EmitProofVerified emits a proof verified event
func (e *EventEmitter) EmitProofVerified(proofID string, proofType types.ProofType, duration time.Duration, userID, requestID string) {
	event := Event{
		Type:      EventProofVerified,
		Source:    e.source,
		UserID:    userID,
		RequestID: requestID,
		Data: map[string]interface{}{
			"proof_id":        proofID,
			"proof_type":      proofType,
			"verify_duration": duration.String(),
		},
	}

	e.bus.PublishAsync(event)
}

// EmitProofFailed emits a proof failed event
func (e *EventEmitter) EmitProofFailed(proofID string, proofType types.ProofType, err error, userID, requestID string) {
	event := Event{
		Type:      EventProofFailed,
		Source:    e.source,
		UserID:    userID,
		RequestID: requestID,
		Data: map[string]interface{}{
			"proof_id":   proofID,
			"proof_type": proofType,
			"error":      err.Error(),
		},
	}

	e.bus.PublishAsync(event)
}

// EmitCircuitCompiled emits a circuit compiled event
func (e *EventEmitter) EmitCircuitCompiled(circuitID, circuitName string, constraints int, compileTime time.Duration) {
	event := Event{
		Type:   EventCircuitCompiled,
		Source: e.source,
		Data: map[string]interface{}{
			"circuit_id":   circuitID,
			"circuit_name": circuitName,
			"constraints":  constraints,
			"compile_time": compileTime.String(),
		},
	}

	e.bus.PublishAsync(event)
}

// EmitError emits a system error event
func (e *EventEmitter) EmitError(component string, err error, userID, requestID string) {
	event := Event{
		Type:      EventError,
		Source:    e.source,
		UserID:    userID,
		RequestID: requestID,
		Data: map[string]interface{}{
			"component": component,
			"error":     err.Error(),
		},
	}

	e.bus.PublishAsync(event)
}

// EmitSystemStarted emits a system started event
func (e *EventEmitter) EmitSystemStarted(version string) {
	event := Event{
		Type:   EventSystemStarted,
		Source: e.source,
		Data: map[string]interface{}{
			"version": version,
		},
	}

	e.bus.PublishAsync(event)
}

// EmitSystemStopped emits a system stopped event
func (e *EventEmitter) EmitSystemStopped(reason string) {
	event := Event{
		Type:   EventSystemStopped,
		Source: e.source,
		Data: map[string]interface{}{
			"reason": reason,
		},
	}

	e.bus.PublishAsync(event)
}
