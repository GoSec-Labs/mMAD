package audit

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"github.com/GoSec-Labs/mMAD/engines/internal/compliance"
	"github.com/GoSec-Labs/mMAD/engines/pkg/logger"
)

// AuditEvent represents an auditable event
type AuditEvent struct {
	ID             string                        `json:"id"`
	Timestamp      time.Time                     `json:"timestamp"`
	UserID         string                        `json:"user_id"`
	SessionID      string                        `json:"session_id,omitempty"`
	IPAddress      string                        `json:"ip_address,omitempty"`
	UserAgent      string                        `json:"user_agent,omitempty"`
	Action         string                        `json:"action"`
	Resource       string                        `json:"resource"`
	ResourceID     string                        `json:"resource_id,omitempty"`
	Details        map[string]interface{}        `json:"details"`
	Result         string                        `json:"result"`
	Risk           compliance.RiskLevel          `json:"risk"`
	Classification compliance.DataClassification `json:"classification"`
	Hash           string                        `json:"hash"`
	PreviousHash   string                        `json:"previous_hash,omitempty"`
	Metadata       map[string]interface{}        `json:"metadata"`
}

// AuditChain represents a blockchain-like audit trail
type AuditChain struct {
	events    []*AuditEvent
	storage   AuditStorage
	mu        sync.RWMutex
	lastHash  string
	eventChan chan *AuditEvent
	stopChan  chan struct{}
	wg        sync.WaitGroup
}

// AuditStorage defines the interface for audit storage
type AuditStorage interface {
	Store(ctx context.Context, event *AuditEvent) error
	Get(ctx context.Context, eventID string) (*AuditEvent, error)
	List(ctx context.Context, filters AuditFilters) ([]*AuditEvent, error)
	GetChain(ctx context.Context, startHash string, limit int) ([]*AuditEvent, error)
	Close() error
}

// AuditFilters represents filters for querying audit events
type AuditFilters struct {
	UserID         string                        `json:"user_id,omitempty"`
	Action         string                        `json:"action,omitempty"`
	Resource       string                        `json:"resource,omitempty"`
	Risk           compliance.RiskLevel          `json:"risk,omitempty"`
	Classification compliance.DataClassification `json:"classification,omitempty"`
	StartDate      *time.Time                    `json:"start_date,omitempty"`
	EndDate        *time.Time                    `json:"end_date,omitempty"`
	Limit          int                           `json:"limit,omitempty"`
	Offset         int                           `json:"offset,omitempty"`
}

// NewAuditChain creates a new audit chain
func NewAuditChain(storage AuditStorage) *AuditChain {
	return &AuditChain{
		events:    make([]*AuditEvent, 0),
		storage:   storage,
		eventChan: make(chan *AuditEvent, 1000),
		stopChan:  make(chan struct{}),
	}
}

// Start starts the audit chain processing
func (ac *AuditChain) Start(ctx context.Context) error {
	ac.wg.Add(1)
	go ac.processEvents(ctx)

	logger.Info("Audit chain started")
	return nil
}

// Stop stops the audit chain processing
func (ac *AuditChain) Stop(ctx context.Context) error {
	close(ac.stopChan)
	ac.wg.Wait()

	logger.Info("Audit chain stopped")
	return nil
}

// LogEvent logs an audit event
func (ac *AuditChain) LogEvent(event *AuditEvent) error {
	// Set timestamp if not provided
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	// Calculate hash
	event.Hash = ac.calculateHash(event)
	event.PreviousHash = ac.lastHash

	// Send to processing channel
	select {
	case ac.eventChan <- event:
		return nil
	default:
		return fmt.Errorf("audit event channel is full")
	}
}

// processEvents processes audit events in the background
func (ac *AuditChain) processEvents(ctx context.Context) {
	defer ac.wg.Done()

	for {
		select {
		case event := <-ac.eventChan:
			if err := ac.processEvent(ctx, event); err != nil {
				logger.Error("Failed to process audit event", "error", err)
			}

		case <-ac.stopChan:
			// Process remaining events
			for len(ac.eventChan) > 0 {
				event := <-ac.eventChan
				if err := ac.processEvent(ctx, event); err != nil {
					logger.Error("Failed to process audit event during shutdown", "error", err)
				}
			}
			return

		case <-ctx.Done():
			return
		}
	}
}

// processEvent processes a single audit event
func (ac *AuditChain) processEvent(ctx context.Context, event *AuditEvent) error {
	ac.mu.Lock()
	defer ac.mu.Unlock()

	// Store event
	if err := ac.storage.Store(ctx, event); err != nil {
		return fmt.Errorf("failed to store audit event: %w", err)
	}

	// Add to chain
	ac.events = append(ac.events, event)
	ac.lastHash = event.Hash

	logger.Debug("Audit event processed",
		"event_id", event.ID,
		"user_id", event.UserID,
		"action", event.Action,
		"resource", event.Resource)

	return nil
}

// calculateHash calculates the hash of an audit event
func (ac *AuditChain) calculateHash(event *AuditEvent) string {
	data := fmt.Sprintf("%s:%s:%s:%s:%s:%s",
		event.Timestamp.Format(time.RFC3339Nano),
		event.UserID,
		event.Action,
		event.Resource,
		event.Result,
		event.PreviousHash,
	)

	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// GetEvents returns audit events with filters
func (ac *AuditChain) GetEvents(ctx context.Context, filters AuditFilters) ([]*AuditEvent, error) {
	return ac.storage.List(ctx, filters)
}

// VerifyChain verifies the integrity of the audit chain
func (ac *AuditChain) VerifyChain(ctx context.Context, startHash string, limit int) (bool, error) {
	events, err := ac.storage.GetChain(ctx, startHash, limit)
	if err != nil {
		return false, fmt.Errorf("failed to get audit chain: %w", err)
	}

	var previousHash string
	for i, event := range events {
		if i == 0 {
			previousHash = event.PreviousHash
		} else {
			if event.PreviousHash != previousHash {
				return false, fmt.Errorf("chain integrity violation at event %s", event.ID)
			}
		}

		// Verify event hash
		expectedHash := ac.calculateEventHash(event)
		if event.Hash != expectedHash {
			return false, fmt.Errorf("event hash mismatch for event %s", event.ID)
		}

		previousHash = event.Hash
	}

	return true, nil
}

// calculateEventHash calculates hash for verification
func (ac *AuditChain) calculateEventHash(event *AuditEvent) string {
	data := fmt.Sprintf("%s:%s:%s:%s:%s:%s",
		event.Timestamp.Format(time.RFC3339Nano),
		event.UserID,
		event.Action,
		event.Resource,
		event.Result,
		event.PreviousHash,
	)

	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// GenerateAuditReport generates an audit report
func (ac *AuditChain) GenerateAuditReport(ctx context.Context, filters AuditFilters) (*AuditReport, error) {
	events, err := ac.GetEvents(ctx, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to get events for report: %w", err)
	}

	report := &AuditReport{
		ID:          fmt.Sprintf("audit_report_%d", time.Now().Unix()),
		GeneratedAt: time.Now(),
		Period: ReportPeriod{
			StartDate: *filters.StartDate,
			EndDate:   *filters.EndDate,
		},
		Events:  events,
		Metrics: ac.calculateMetrics(events),
	}

	return report, nil
}

// calculateMetrics calculates audit metrics
func (ac *AuditChain) calculateMetrics(events []*AuditEvent) AuditMetrics {
	metrics := AuditMetrics{
		TotalEvents:      len(events),
		ActionCounts:     make(map[string]int),
		ResourceCounts:   make(map[string]int),
		RiskDistribution: make(map[compliance.RiskLevel]int),
		UserActivity:     make(map[string]int),
	}

	for _, event := range events {
		metrics.ActionCounts[event.Action]++
		metrics.ResourceCounts[event.Resource]++
		metrics.RiskDistribution[event.Risk]++
		metrics.UserActivity[event.UserID]++
	}

	return metrics
}

// AuditReport represents an audit report
type AuditReport struct {
	ID          string        `json:"id"`
	GeneratedAt time.Time     `json:"generated_at"`
	Period      ReportPeriod  `json:"period"`
	Events      []*AuditEvent `json:"events"`
	Metrics     AuditMetrics  `json:"metrics"`
}

// ReportPeriod represents a time period for reporting
type ReportPeriod struct {
	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`
}

// AuditMetrics represents audit metrics
type AuditMetrics struct {
	TotalEvents      int                          `json:"total_events"`
	ActionCounts     map[string]int               `json:"action_counts"`
	ResourceCounts   map[string]int               `json:"resource_counts"`
	RiskDistribution map[compliance.RiskLevel]int `json:"risk_distribution"`
	UserActivity     map[string]int               `json:"user_activity"`
}
