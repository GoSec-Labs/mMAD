package models

import (
	"time"

	"github.com/GoSec-Labs/mMAD/engines/internal/zkproof/types"
)

// APIResponse is the standard API response wrapper
type APIResponse struct {
	Success   bool        `json:"success"`
	Data      interface{} `json:"data,omitempty"`
	Error     *APIError   `json:"error,omitempty"`
	RequestID string      `json:"request_id,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
}

// APIError represents an API error
type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// ProofResponse represents a proof generation response
type ProofResponse struct {
	ProofID      string          `json:"proof_id"`
	Status       string          `json:"status"`
	ProofData    string          `json:"proof_data,omitempty"`
	ProofType    types.ProofType `json:"proof_type"`
	CircuitID    string          `json:"circuit_id"`
	GeneratedAt  time.Time       `json:"generated_at,omitempty"`
	Duration     time.Duration   `json:"duration,omitempty"`
	ProofSize    int             `json:"proof_size,omitempty"`
	VerifyingKey string          `json:"verifying_key,omitempty"`
}

// VerifyResponse represents a proof verification response
type VerifyResponse struct {
	Valid      bool          `json:"valid"`
	ProofID    string        `json:"proof_id,omitempty"`
	VerifiedAt time.Time     `json:"verified_at"`
	Duration   time.Duration `json:"duration"`
	CircuitID  string        `json:"circuit_id"`
}

// CircuitInfo represents circuit information
type CircuitInfo struct {
	ID            string          `json:"id"`
	Name          string          `json:"name"`
	Description   string          `json:"description"`
	Version       string          `json:"version"`
	ProofType     types.ProofType `json:"proof_type"`
	Constraints   int             `json:"constraints"`
	IsCompiled    bool            `json:"is_compiled"`
	CompiledAt    time.Time       `json:"compiled_at,omitempty"`
	EstimatedTime string          `json:"estimated_time"`
}

// BatchResponse represents batch operation response
type BatchResponse struct {
	BatchID     string        `json:"batch_id"`
	Status      string        `json:"status"`
	Total       int           `json:"total"`
	Completed   int           `json:"completed"`
	Failed      int           `json:"failed"`
	Results     interface{}   `json:"results,omitempty"`
	StartedAt   time.Time     `json:"started_at"`
	CompletedAt time.Time     `json:"completed_at,omitempty"`
	Duration    time.Duration `json:"duration,omitempty"`
}

// SystemMetrics represents system metrics
type SystemMetrics struct {
	ProofStats   ProofStats   `json:"proof_stats"`
	CircuitStats CircuitStats `json:"circuit_stats"`
	SystemStats  SystemStats  `json:"system_stats"`
	EventStats   EventStats   `json:"event_stats"`
}

type ProofStats struct {
	TotalGenerated  int64         `json:"total_generated"`
	TotalVerified   int64         `json:"total_verified"`
	TotalFailed     int64         `json:"total_failed"`
	AvgGenerateTime time.Duration `json:"avg_generate_time"`
	AvgVerifyTime   time.Duration `json:"avg_verify_time"`
	ActiveProofs    int           `json:"active_proofs"`
}

type CircuitStats struct {
	TotalCircuits    int           `json:"total_circuits"`
	CompiledCircuits int           `json:"compiled_circuits"`
	AvgCompileTime   time.Duration `json:"avg_compile_time"`
	CacheHitRate     float64       `json:"cache_hit_rate"`
}

type SystemStats struct {
	Uptime         time.Duration `json:"uptime"`
	MemoryUsage    int64         `json:"memory_usage"`
	CPUUsage       float64       `json:"cpu_usage"`
	ActiveRequests int           `json:"active_requests"`
	QueueSize      int           `json:"queue_size"`
}

type EventStats struct {
	TotalEvents  int64            `json:"total_events"`
	EventsByType map[string]int64 `json:"events_by_type"`
	RecentEvents int              `json:"recent_events"`
	ErrorRate    float64          `json:"error_rate"`
}
