package handlers

import (
	"time"

	"github.com/GoSec-Labs/mMAD/engines/internal/circuits"
	"github.com/GoSec-Labs/mMAD/engines/internal/events"
	"github.com/GoSec-Labs/mMAD/engines/internal/zkproof"
	"github.com/GoSec-Labs/mMAD/engines/pkg/config"
)

// Handler contains all API handlers
type Handler struct {
    config      *config.Config
    prover      *zkproof.ProofEngine
    circuits    *circuits.CircuitRegistry
    events      *events.Manager
    startTime   time.Time
}

// New creates a new API handler
func New(
    cfg *config.Config,
    prover *zkproof.ProofEngine,
    circuits *circuits.CircuitRegistry,
    events *events.Manager,
) *Handler {
    return &Handler{
        config:    cfg,
        prover:    prover,
        circuits:  circuits,
        events:    events,
        startTime: time.Now(),
    }
}

// HealthStatus represents service health
type HealthStatus struct {
    Status    string            `json:"status"`
    Timestamp time.Time         `json:"timestamp"`
    Version   string            `json:"version"`
    Services  map[string]string `json:"services"`
    Uptime    time.Duration     `json:"uptime"`
}