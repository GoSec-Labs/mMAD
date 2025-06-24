package privacy

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/GoSec-Labs/mMAD/engines/pkg/logger"
)

// RetentionManager manages data retention policies
type RetentionManager interface {
	CreatePolicy(ctx context.Context, policy *RetentionPolicy) error
	GetPolicy(ctx context.Context, policyID string) (*RetentionPolicy, error)
	GetPolicyByType(ctx context.Context, dataType string) (*RetentionPolicy, error)
	UpdatePolicy(ctx context.Context, policy *RetentionPolicy) error
	DeletePolicy(ctx context.Context, policyID string) error
	ApplyRetention(ctx context.Context, dataID string) error
	ProcessRetention(ctx context.Context) error
	GetExpiredData(ctx context.Context) ([]*PersonalData, error)
}

// SimpleRetentionManager implements RetentionManager
type SimpleRetentionManager struct {
	policyStorage PolicyStorage
	dataStore     DataStore
	anonymizer    DataAnonymizer
}

// PolicyStorage defines the interface for retention policy storage
type PolicyStorage interface {
	Store(ctx context.Context, policy *RetentionPolicy) error
	Get(ctx context.Context, policyID string) (*RetentionPolicy, error)
	GetByType(ctx context.Context, dataType string) (*RetentionPolicy, error)
	Update(ctx context.Context, policy *RetentionPolicy) error
	Delete(ctx context.Context, policyID string) error
	List(ctx context.Context) ([]*RetentionPolicy, error)
	Close() error
}

// RetentionPolicy represents a data retention policy
type RetentionPolicy struct {
	ID              string                 `json:"id"`
	Name            string                 `json:"name"`
	Description     string                 `json:"description"`
	DataType        string                 `json:"data_type"`
	Category        DataCategory           `json:"category"`
	RetentionPeriod time.Duration          `json:"retention_period"`
	Action          RetentionAction        `json:"action"`
	LegalBasis      LegalBasis             `json:"legal_basis"`
	Conditions      []RetentionCondition   `json:"conditions"`
	Enabled         bool                   `json:"enabled"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
	Metadata        map[string]interface{} `json:"metadata"`
}

// RetentionAction represents the action to take when data expires
type RetentionAction string

const (
	ActionDelete    RetentionAction = "delete"
	ActionAnonymize RetentionAction = "anonymize"
	ActionArchive   RetentionAction = "archive"
	ActionReview    RetentionAction = "review"
)

// RetentionCondition represents a condition for retention
type RetentionCondition struct {
	Field    string      `json:"field"`
	Operator string      `json:"operator"`
	Value    interface{} `json:"value"`
}

// RetentionSchedule represents a scheduled retention task
type RetentionSchedule struct {
	ID          string                 `json:"id"`
	PolicyID    string                 `json:"policy_id"`
	DataID      string                 `json:"data_id"`
	ScheduledAt time.Time              `json:"scheduled_at"`
	Status      ScheduleStatus         `json:"status"`
	Action      RetentionAction        `json:"action"`
	ProcessedAt *time.Time             `json:"processed_at,omitempty"`
	Error       string                 `json:"error,omitempty"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// ScheduleStatus represents the status of a retention schedule
type ScheduleStatus string

const (
	SchedulePending   ScheduleStatus = "pending"
	ScheduleProcessed ScheduleStatus = "processed"
	ScheduleFailed    ScheduleStatus = "failed"
)

// NewSimpleRetentionManager creates a new retention manager
func NewSimpleRetentionManager(
	policyStorage PolicyStorage,
	dataStore DataStore,
	anonymizer DataAnonymizer,
) *SimpleRetentionManager {
	return &SimpleRetentionManager{
		policyStorage: policyStorage,
		dataStore:     dataStore,
		anonymizer:    anonymizer,
	}
}

// CreatePolicy creates a new retention policy
func (srm *SimpleRetentionManager) CreatePolicy(ctx context.Context, policy *RetentionPolicy) error {
	if policy.CreatedAt.IsZero() {
		policy.CreatedAt = time.Now()
	}
	policy.UpdatedAt = time.Now()

	if err := srm.policyStorage.Store(ctx, policy); err != nil {
		return fmt.Errorf("failed to store retention policy: %w", err)
	}

	logger.Info("Retention policy created",
		"policy_id", policy.ID,
		"data_type", policy.DataType,
		"retention_period", policy.RetentionPeriod,
		"action", policy.Action)

	return nil
}

// GetPolicy retrieves a retention policy by ID
func (srm *SimpleRetentionManager) GetPolicy(ctx context.Context, policyID string) (*RetentionPolicy, error) {
	return srm.policyStorage.Get(ctx, policyID)
}

// GetPolicyByType retrieves a retention policy by data type
func (srm *SimpleRetentionManager) GetPolicyByType(ctx context.Context, dataType string) (*RetentionPolicy, error) {
	return srm.policyStorage.GetByType(ctx, dataType)
}

// UpdatePolicy updates a retention policy
func (srm *SimpleRetentionManager) UpdatePolicy(ctx context.Context, policy *RetentionPolicy) error {
	policy.UpdatedAt = time.Now()

	if err := srm.policyStorage.Update(ctx, policy); err != nil {
		return fmt.Errorf("failed to update retention policy: %w", err)
	}

	logger.Info("Retention policy updated", "policy_id", policy.ID)
	return nil
}

// DeletePolicy deletes a retention policy
func (srm *SimpleRetentionManager) DeletePolicy(ctx context.Context, policyID string) error {
	if err := srm.policyStorage.Delete(ctx, policyID); err != nil {
		return fmt.Errorf("failed to delete retention policy: %w", err)
	}

	logger.Info("Retention policy deleted", "policy_id", policyID)
	return nil
}

// ApplyRetention applies retention policy to specific data
func (srm *SimpleRetentionManager) ApplyRetention(ctx context.Context, dataID string) error {
	// Get the data
	data, err := srm.dataStore.Get(ctx, dataID)
	if err != nil {
		return fmt.Errorf("failed to get data: %w", err)
	}

	// Get applicable policy
	policy, err := srm.GetPolicyByType(ctx, data.DataType)
	if err != nil {
		return fmt.Errorf("failed to get retention policy: %w", err)
	}

	if policy == nil || !policy.Enabled {
		return nil // No policy or policy disabled
	}

	// Check if data should be processed
	if !srm.shouldProcessData(data, policy) {
		return nil
	}

	// Calculate expiry time
	expiryTime := data.CreatedAt.Add(policy.RetentionPeriod)
	data.ExpiresAt = &expiryTime

	// Update data with expiry
	if err := srm.dataStore.Update(ctx, data); err != nil {
		return fmt.Errorf("failed to update data with expiry: %w", err)
	}

	logger.Info("Retention policy applied",
		"data_id", dataID,
		"policy_id", policy.ID,
		"expires_at", expiryTime)

	return nil
}

// ProcessRetention processes all expired data according to retention policies
func (srm *SimpleRetentionManager) ProcessRetention(ctx context.Context) error {
	expiredData, err := srm.GetExpiredData(ctx)
	if err != nil {
		return fmt.Errorf("failed to get expired data: %w", err)
	}

	logger.Info("Processing retention", "expired_count", len(expiredData))

	for _, data := range expiredData {
		if err := srm.processExpiredData(ctx, data); err != nil {
			logger.Error("Failed to process expired data",
				"data_id", data.ID,
				"error", err)
		}
	}

	return nil
}

// GetExpiredData returns all data that has expired
func (srm *SimpleRetentionManager) GetExpiredData(ctx context.Context) ([]*PersonalData, error) {
	now := time.Now()
	query := DataQuery{
		EndDate: &now,
		Limit:   1000, // Process in batches
	}

	allData, err := srm.dataStore.Search(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to search for expired data: %w", err)
	}

	var expiredData []*PersonalData
	for _, data := range allData {
		if data.ExpiresAt != nil && now.After(*data.ExpiresAt) {
			expiredData = append(expiredData, data)
		}
	}

	return expiredData, nil
}

// processExpiredData processes a single expired data item
func (srm *SimpleRetentionManager) processExpiredData(ctx context.Context, data *PersonalData) error {
	// Get applicable policy
	policy, err := srm.GetPolicyByType(ctx, data.DataType)
	if err != nil {
		return fmt.Errorf("failed to get retention policy: %w", err)
	}

	if policy == nil {
		return fmt.Errorf("no retention policy found for data type: %s", data.DataType)
	}

	switch policy.Action {
	case ActionDelete:
		return srm.deleteData(ctx, data)
	case ActionAnonymize:
		return srm.anonymize
	case ActionAnonymize:
		return srm.anonymizeData(ctx, data)
	case ActionArchive:
		return srm.archiveData(ctx, data)
	case ActionReview:
		return srm.markForReview(ctx, data)
	default:
		return fmt.Errorf("unknown retention action: %s", policy.Action)
	}
}

// deleteData deletes expired data
func (srm *SimpleRetentionManager) deleteData(ctx context.Context, data *PersonalData) error {
	if err := srm.dataStore.Delete(ctx, data.ID); err != nil {
		return fmt.Errorf("failed to delete data: %w", err)
	}

	logger.Info("Data deleted due to retention policy",
		"data_id", data.ID,
		"data_type", data.DataType,
		"subject_id", data.SubjectID)

	return nil
}

// anonymizeData anonymizes expired data
func (srm *SimpleRetentionManager) anonymizeData(ctx context.Context, data *PersonalData) error {
	// Anonymize the data
	anonymizedData, err := srm.anonymizer.Anonymize(data.Data)
	if err != nil {
		return fmt.Errorf("failed to anonymize data: %w", err)
	}

	// Update data record
	data.Data = anonymizedData
	data.IsAnonymized = true
	data.UpdatedAt = time.Now()
	if data.Metadata == nil {
		data.Metadata = make(map[string]interface{})
	}
	data.Metadata["anonymized_at"] = time.Now()
	data.Metadata["retention_action"] = "anonymized"

	if err := srm.dataStore.Update(ctx, data); err != nil {
		return fmt.Errorf("failed to update anonymized data: %w", err)
	}

	logger.Info("Data anonymized due to retention policy",
		"data_id", data.ID,
		"data_type", data.DataType,
		"subject_id", data.SubjectID)

	return nil
}

// archiveData archives expired data
func (srm *SimpleRetentionManager) archiveData(ctx context.Context, data *PersonalData) error {
	// Mark as archived
	if data.Metadata == nil {
		data.Metadata = make(map[string]interface{})
	}
	data.Metadata["archived"] = true
	data.Metadata["archived_at"] = time.Now()
	data.Metadata["retention_action"] = "archived"
	data.UpdatedAt = time.Now()

	if err := srm.dataStore.Update(ctx, data); err != nil {
		return fmt.Errorf("failed to archive data: %w", err)
	}

	logger.Info("Data archived due to retention policy",
		"data_id", data.ID,
		"data_type", data.DataType,
		"subject_id", data.SubjectID)

	return nil
}

// markForReview marks data for manual review
func (srm *SimpleRetentionManager) markForReview(ctx context.Context, data *PersonalData) error {
	if data.Metadata == nil {
		data.Metadata = make(map[string]interface{})
	}
	data.Metadata["requires_review"] = true
	data.Metadata["review_due_at"] = time.Now()
	data.Metadata["retention_action"] = "review"
	data.UpdatedAt = time.Now()

	if err := srm.dataStore.Update(ctx, data); err != nil {
		return fmt.Errorf("failed to mark data for review: %w", err)
	}

	logger.Info("Data marked for review due to retention policy",
		"data_id", data.ID,
		"data_type", data.DataType,
		"subject_id", data.SubjectID)

	return nil
}

// shouldProcessData checks if data should be processed by retention policy
func (srm *SimpleRetentionManager) shouldProcessData(data *PersonalData, policy *RetentionPolicy) bool {
	// Check category match
	if policy.Category != "" && data.Category != policy.Category {
		return false
	}

	// Check conditions
	for _, condition := range policy.Conditions {
		if !srm.evaluateCondition(data, condition) {
			return false
		}
	}

	return true
}

// evaluateCondition evaluates a retention condition
func (srm *SimpleRetentionManager) evaluateCondition(data *PersonalData, condition RetentionCondition) bool {
	var fieldValue interface{}

	switch condition.Field {
	case "data_type":
		fieldValue = data.DataType
	case "category":
		fieldValue = data.Category
	case "legal_basis":
		fieldValue = data.LegalBasis
	case "created_at":
		fieldValue = data.CreatedAt
	default:
		// Check in data fields
		if val, exists := data.Data[condition.Field]; exists {
			fieldValue = val
		} else {
			return false
		}
	}

	return srm.compareValues(fieldValue, condition.Operator, condition.Value)
}

// compareValues compares two values based on operator
func (srm *SimpleRetentionManager) compareValues(fieldValue interface{}, operator string, conditionValue interface{}) bool {
	switch operator {
	case "eq":
		return fieldValue == conditionValue
	case "ne":
		return fieldValue != conditionValue
	case "gt":
		return srm.compareNumeric(fieldValue, conditionValue) > 0
	case "lt":
		return srm.compareNumeric(fieldValue, conditionValue) < 0
	case "gte":
		return srm.compareNumeric(fieldValue, conditionValue) >= 0
	case "lte":
		return srm.compareNumeric(fieldValue, conditionValue) <= 0
	case "contains":
		if str, ok := fieldValue.(string); ok {
			if substr, ok := conditionValue.(string); ok {
				return strings.Contains(str, substr)
			}
		}
		return false
	default:
		return false
	}
}

// compareNumeric compares numeric values
func (srm *SimpleRetentionManager) compareNumeric(a, b interface{}) int {
	// This is a simplified implementation
	// In practice, you'd want more robust type handling
	switch av := a.(type) {
	case int:
		if bv, ok := b.(int); ok {
			if av > bv {
				return 1
			} else if av < bv {
				return -1
			}
			return 0
		}
	case float64:
		if bv, ok := b.(float64); ok {
			if av > bv {
				return 1
			} else if av < bv {
				return -1
			}
			return 0
		}
	case time.Time:
		if bv, ok := b.(time.Time); ok {
			if av.After(bv) {
				return 1
			} else if av.Before(bv) {
				return -1
			}
			return 0
		}
	}
	return 0
}
