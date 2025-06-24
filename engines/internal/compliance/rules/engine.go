package rules

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/GoSec-Labs/mMAD/engines/internal/compliance"
	"github.com/GoSec-Labs/mMAD/engines/pkg/logger"
)

// RuleEngine interface for compliance rule evaluation
type RuleEngine interface {
	AddRule(ctx context.Context, rule *compliance.ComplianceRule) error
	UpdateRule(ctx context.Context, rule *compliance.ComplianceRule) error
	DeleteRule(ctx context.Context, ruleID string) error
	GetRule(ctx context.Context, ruleID string) (*compliance.ComplianceRule, error)
	ListRules(ctx context.Context, regulation compliance.RegulationType) ([]*compliance.ComplianceRule, error)
	EvaluateEvent(ctx context.Context, event *compliance.ComplianceEvent) ([]*compliance.ComplianceViolation, error)
	EvaluateAllRules(ctx context.Context, event *compliance.ComplianceEvent) ([]*compliance.ComplianceViolation, error)
}

// SimpleRuleEngine implements RuleEngine
type SimpleRuleEngine struct {
	storage RuleStorage
}

// RuleStorage defines the interface for rule storage
type RuleStorage interface {
	Store(ctx context.Context, rule *compliance.ComplianceRule) error
	Get(ctx context.Context, ruleID string) (*compliance.ComplianceRule, error)
	Update(ctx context.Context, rule *compliance.ComplianceRule) error
	Delete(ctx context.Context, ruleID string) error
	List(ctx context.Context, regulation compliance.RegulationType) ([]*compliance.ComplianceRule, error)
	ListAll(ctx context.Context) ([]*compliance.ComplianceRule, error)
	Close() error
}

// NewSimpleRuleEngine creates a new rule engine
func NewSimpleRuleEngine(storage RuleStorage) *SimpleRuleEngine {
	return &SimpleRuleEngine{
		storage: storage,
	}
}

// AddRule adds a new compliance rule
func (sre *SimpleRuleEngine) AddRule(ctx context.Context, rule *compliance.ComplianceRule) error {
	if rule.CreatedAt.IsZero() {
		rule.CreatedAt = time.Now()
	}
	rule.UpdatedAt = time.Now()

	if err := sre.storage.Store(ctx, rule); err != nil {
		return fmt.Errorf("failed to store rule: %w", err)
	}

	logger.Info("Compliance rule added",
		"rule_id", rule.ID,
		"name", rule.Name,
		"regulation", rule.Regulation)

	return nil
}

// UpdateRule updates an existing compliance rule
func (sre *SimpleRuleEngine) UpdateRule(ctx context.Context, rule *compliance.ComplianceRule) error {
	rule.UpdatedAt = time.Now()

	if err := sre.storage.Update(ctx, rule); err != nil {
		return fmt.Errorf("failed to update rule: %w", err)
	}

	logger.Info("Compliance rule updated", "rule_id", rule.ID)
	return nil
}

// DeleteRule deletes a compliance rule
func (sre *SimpleRuleEngine) DeleteRule(ctx context.Context, ruleID string) error {
	if err := sre.storage.Delete(ctx, ruleID); err != nil {
		return fmt.Errorf("failed to delete rule: %w", err)
	}

	logger.Info("Compliance rule deleted", "rule_id", ruleID)
	return nil
}

// GetRule retrieves a compliance rule by ID
func (sre *SimpleRuleEngine) GetRule(ctx context.Context, ruleID string) (*compliance.ComplianceRule, error) {
	return sre.storage.Get(ctx, ruleID)
}

// ListRules lists rules by regulation type
func (sre *SimpleRuleEngine) ListRules(ctx context.Context, regulation compliance.RegulationType) ([]*compliance.ComplianceRule, error) {
	return sre.storage.List(ctx, regulation)
}

// EvaluateEvent evaluates an event against applicable rules
func (sre *SimpleRuleEngine) EvaluateEvent(ctx context.Context, event *compliance.ComplianceEvent) ([]*compliance.ComplianceViolation, error) {
	// Get rules for the event's regulation
	rules, err := sre.storage.List(ctx, event.Regulation)
	if err != nil {
		return nil, fmt.Errorf("failed to get rules: %w", err)
	}

	var violations []*compliance.ComplianceViolation

	for _, rule := range rules {
		if !rule.Enabled {
			continue
		}

		// Check if rule applies to this event
		if sre.ruleApplies(rule, event) {
			// Evaluate rule conditions
			if sre.evaluateRule(rule, event) {
				violation := &compliance.ComplianceViolation{
					ID:          fmt.Sprintf("violation_%d", time.Now().UnixNano()),
					RuleID:      rule.ID,
					EventID:     event.ID,
					UserID:      event.UserID,
					EntityType:  event.EntityType,
					EntityID:    event.EntityID,
					Regulation:  rule.Regulation,
					Severity:    rule.Severity,
					Description: fmt.Sprintf("Rule violation: %s", rule.Description),
					Status:      compliance.ViolationOpen,
					DetectedAt:  time.Now(),
					Details: map[string]interface{}{
						"rule_name":     rule.Name,
						"rule_category": rule.Category,
						"event_action":  event.Action,
						"event_details": event.Details,
					},
					Metadata: map[string]interface{}{
						"detection_method": "rule_engine",
						"rule_version":     rule.UpdatedAt,
					},
				}

				violations = append(violations, violation)

				logger.Warn("Compliance rule violated",
					"rule_id", rule.ID,
					"event_id", event.ID,
					"user_id", event.UserID,
					"severity", rule.Severity)
			}
		}
	}

	return violations, nil
}

// EvaluateAllRules evaluates an event against all rules
func (sre *SimpleRuleEngine) EvaluateAllRules(ctx context.Context, event *compliance.ComplianceEvent) ([]*compliance.ComplianceViolation, error) {
	rules, err := sre.storage.ListAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get all rules: %w", err)
	}

	var violations []*compliance.ComplianceViolation

	for _, rule := range rules {
		if !rule.Enabled {
			continue
		}

		if sre.ruleApplies(rule, event) && sre.evaluateRule(rule, event) {
			violation := &compliance.ComplianceViolation{
				ID:          fmt.Sprintf("violation_%d", time.Now().UnixNano()),
				RuleID:      rule.ID,
				EventID:     event.ID,
				UserID:      event.UserID,
				EntityType:  event.EntityType,
				EntityID:    event.EntityID,
				Regulation:  rule.Regulation,
				Severity:    rule.Severity,
				Description: fmt.Sprintf("Rule violation: %s", rule.Description),
				Status:      compliance.ViolationOpen,
				DetectedAt:  time.Now(),
				Details: map[string]interface{}{
					"rule_name":     rule.Name,
					"rule_category": rule.Category,
					"event_action":  event.Action,
					"event_details": event.Details,
				},
				Metadata: map[string]interface{}{
					"detection_method": "rule_engine",
					"rule_version":     rule.UpdatedAt,
				},
			}

			violations = append(violations, violation)
		}
	}

	return violations, nil
}

// ruleApplies checks if a rule applies to an event
func (sre *SimpleRuleEngine) ruleApplies(rule *compliance.ComplianceRule, event *compliance.ComplianceEvent) bool {
	// Check regulation match
	if rule.Regulation != event.Regulation {
		return false
	}

	// Check category-specific logic
	switch rule.Category {
	case "data_access":
		return strings.Contains(event.Action, "access") || strings.Contains(event.Action, "read")
	case "data_modification":
		return strings.Contains(event.Action, "update") || strings.Contains(event.Action, "modify") || strings.Contains(event.Action, "delete")
	case "data_export":
		return strings.Contains(event.Action, "export") || strings.Contains(event.Action, "download")
	case "authentication":
		return strings.Contains(event.Action, "login")
	case "authorization":
		return strings.Contains(event.Action, "authorize") || strings.Contains(event.Action, "permission")
	case "data_retention":
		return strings.Contains(event.Action, "retention") || strings.Contains(event.Action, "expire")
	case "consent":
		return strings.Contains(event.Action, "consent")
	case "breach":
		return strings.Contains(event.Action, "breach") || strings.Contains(event.Action, "incident")
	default:
		return true // Apply to all events if no specific category
	}
}

// evaluateRule evaluates rule conditions against an event
func (sre *SimpleRuleEngine) evaluateRule(rule *compliance.ComplianceRule, event *compliance.ComplianceEvent) bool {
	if len(rule.Conditions) == 0 {
		return false // No conditions means no violation
	}

	// All conditions must be true for a violation
	for _, condition := range rule.Conditions {
		if !sre.evaluateCondition(condition, event) {
			return false
		}
	}

	return true
}

// evaluateCondition evaluates a single rule condition
func (sre *SimpleRuleEngine) evaluateCondition(condition compliance.RuleCondition, event *compliance.ComplianceEvent) bool {
	fieldValue := sre.getFieldValue(condition.Field, event)
	return sre.compareValues(fieldValue, condition.Operator, condition.Value)
}

// getFieldValue extracts field value from event
func (sre *SimpleRuleEngine) getFieldValue(field string, event *compliance.ComplianceEvent) interface{} {
	switch field {
	case "user_id":
		return event.UserID
	case "action":
		return event.Action
	case "entity_type":
		return event.EntityType
	case "entity_id":
		return event.EntityID
	case "risk":
		return event.Risk
	case "classification":
		return event.Classification
	case "timestamp":
		return event.Timestamp
	case "hour":
		return event.Timestamp.Hour()
	case "day_of_week":
		return int(event.Timestamp.Weekday())
	default:
		// Check in event details
		if val, exists := event.Details[field]; exists {
			return val
		}
		// Check in metadata
		if val, exists := event.Metadata[field]; exists {
			return val
		}
		return nil
	}
}

// compareValues compares field value with condition value using operator
func (sre *SimpleRuleEngine) compareValues(fieldValue interface{}, operator string, conditionValue interface{}) bool {
	if fieldValue == nil {
		return operator == "is_null"
	}

	switch operator {
	case "eq", "equals":
		return fieldValue == conditionValue
	case "ne", "not_equals":
		return fieldValue != conditionValue
	case "gt", "greater_than":
		return sre.compareNumeric(fieldValue, conditionValue) > 0
	case "lt", "less_than":
		return sre.compareNumeric(fieldValue, conditionValue) < 0
	case "gte", "greater_than_equals":
		return sre.compareNumeric(fieldValue, conditionValue) >= 0
	case "lte", "less_than_equals":
		return sre.compareNumeric(fieldValue, conditionValue) <= 0
	case "contains":
		return sre.stringContains(fieldValue, conditionValue)
	case "starts_with":
		return sre.stringStartsWith(fieldValue, conditionValue)
	case "ends_with":
		return sre.stringEndsWith(fieldValue, conditionValue)
	case "in":
		return sre.valueInList(fieldValue, conditionValue)
	case "not_in":
		return !sre.valueInList(fieldValue, conditionValue)
	case "regex":
		return sre.regexMatch(fieldValue, conditionValue)
	case "is_null":
		return fieldValue == nil
	case "is_not_null":
		return fieldValue != nil
	default:
		logger.Warn("Unknown operator in rule condition", "operator", operator)
		return false
	}
}

// compareNumeric compares numeric values
func (sre *SimpleRuleEngine) compareNumeric(a, b interface{}) int {
	aFloat := sre.toFloat64(a)
	bFloat := sre.toFloat64(b)

	if aFloat > bFloat {
		return 1
	} else if aFloat < bFloat {
		return -1
	}
	return 0
}

// toFloat64 converts various numeric types to float64
func (sre *SimpleRuleEngine) toFloat64(val interface{}) float64 {
	switch v := val.(type) {
	case int:
		return float64(v)
	case int32:
		return float64(v)
	case int64:
		return float64(v)
	case float32:
		return float64(v)
	case float64:
		return v
	case time.Time:
		return float64(v.Unix())
	default:
		return 0
	}
}

// stringContains checks if string contains substring
func (sre *SimpleRuleEngine) stringContains(fieldValue, conditionValue interface{}) bool {
	str := fmt.Sprintf("%v", fieldValue)
	substr := fmt.Sprintf("%v", conditionValue)
	return strings.Contains(strings.ToLower(str), strings.ToLower(substr))
}

// stringStartsWith checks if string starts with prefix
func (sre *SimpleRuleEngine) stringStartsWith(fieldValue, conditionValue interface{}) bool {
	str := fmt.Sprintf("%v", fieldValue)
	prefix := fmt.Sprintf("%v", conditionValue)
	return strings.HasPrefix(strings.ToLower(str), strings.ToLower(prefix))
}

// stringEndsWith checks if string ends with suffix
func (sre *SimpleRuleEngine) stringEndsWith(fieldValue, conditionValue interface{}) bool {
	str := fmt.Sprintf("%v", fieldValue)
	suffix := fmt.Sprintf("%v", conditionValue)
	return strings.HasSuffix(strings.ToLower(str), strings.ToLower(suffix))
}

// valueInList checks if value is in list
func (sre *SimpleRuleEngine) valueInList(fieldValue, conditionValue interface{}) bool {
	list, ok := conditionValue.([]interface{})
	if !ok {
		return false
	}

	for _, item := range list {
		if fieldValue == item {
			return true
		}
	}
	return false
}

// regexMatch checks if value matches regex pattern
func (sre *SimpleRuleEngine) regexMatch(fieldValue, conditionValue interface{}) bool {
	// Simplified regex matching - in production, use regexp package
	str := fmt.Sprintf("%v", fieldValue)
	pattern := fmt.Sprintf("%v", conditionValue)

	// Basic pattern matching
	return strings.Contains(str, pattern)
}
