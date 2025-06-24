package storage

import (
	"context"
	"fmt"
	"sync"

	"github.com/GoSec-Labs/mMAD/engines/internal/compliance"
	"github.com/GoSec-Labs/mMAD/engines/internal/compliance/audit"
	"github.com/GoSec-Labs/mMAD/engines/internal/compliance/privacy"
)

// MemoryComplianceStorage implements all storage interfaces in memory
type MemoryComplianceStorage struct {
	// Compliance data
	rules      map[string]*compliance.ComplianceRule
	violations map[string]*compliance.ComplianceViolation
	reports    map[string]*compliance.ComplianceReport

	// Audit data
	auditEvents map[string]*audit.AuditEvent

	// Privacy data
	personalData      map[string]*privacy.PersonalData
	consents          map[string]*privacy.Consent
	consentHistory    map[string][]*privacy.ConsentHistory
	retentionPolicies map[string]*privacy.RetentionPolicy

	mu sync.RWMutex
}

// NewMemoryComplianceStorage creates a new memory-based storage
func NewMemoryComplianceStorage() *MemoryComplianceStorage {
	return &MemoryComplianceStorage{
		rules:             make(map[string]*compliance.ComplianceRule),
		violations:        make(map[string]*compliance.ComplianceViolation),
		reports:           make(map[string]*compliance.ComplianceReport),
		auditEvents:       make(map[string]*audit.AuditEvent),
		personalData:      make(map[string]*privacy.PersonalData),
		consents:          make(map[string]*privacy.Consent),
		consentHistory:    make(map[string][]*privacy.ConsentHistory),
		retentionPolicies: make(map[string]*privacy.RetentionPolicy),
	}
}

// Rule Storage Implementation
func (mcs *MemoryComplianceStorage) Store(ctx context.Context, rule *compliance.ComplianceRule) error {
	mcs.mu.Lock()
	defer mcs.mu.Unlock()

	ruleCopy := *rule
	mcs.rules[rule.ID] = &ruleCopy
	return nil
}

func (mcs *MemoryComplianceStorage) Get(ctx context.Context, ruleID string) (*compliance.ComplianceRule, error) {
	mcs.mu.RLock()
	defer mcs.mu.RUnlock()

	rule, exists := mcs.rules[ruleID]
	if !exists {
		return nil, fmt.Errorf("rule not found: %s", ruleID)
	}

	ruleCopy := *rule
	return &ruleCopy, nil
}

func (mcs *MemoryComplianceStorage) Update(ctx context.Context, rule *compliance.ComplianceRule) error {
	mcs.mu.Lock()
	defer mcs.mu.Unlock()

	if _, exists := mcs.rules[rule.ID]; !exists {
		return fmt.Errorf("rule not found: %s", rule.ID)
	}

	ruleCopy := *rule
	mcs.rules[rule.ID] = &ruleCopy
	return nil
}

func (mcs *MemoryComplianceStorage) Delete(ctx context.Context, ruleID string) error {
	mcs.mu.Lock()
	defer mcs.mu.Unlock()

	delete(mcs.rules, ruleID)
	return nil
}

func (mcs *MemoryComplianceStorage) List(ctx context.Context, regulation compliance.RegulationType) ([]*compliance.ComplianceRule, error) {
	mcs.mu.RLock()
	defer mcs.mu.RUnlock()

	var rules []*compliance.ComplianceRule
	for _, rule := range mcs.rules {
		if rule.Regulation == regulation {
			ruleCopy := *rule
			rules = append(rules, &ruleCopy)
		}
	}

	return rules, nil
}

func (mcs *MemoryComplianceStorage) ListAll(ctx context.Context) ([]*compliance.ComplianceRule, error) {
	mcs.mu.RLock()
	defer mcs.mu.RUnlock()

	var rules []*compliance.ComplianceRule
	for _, rule := range mcs.rules {
		ruleCopy := *rule
		rules = append(rules, &ruleCopy)
	}

	return rules, nil
}

// Violation Storage Implementation
func (mcs *MemoryComplianceStorage) StoreViolation(ctx context.Context, violation *compliance.ComplianceViolation) error {
	mcs.mu.Lock()
	defer mcs.mu.Unlock()

	violationCopy := *violation
	mcs.violations[violation.ID] = &violationCopy
	return nil
}

func (mcs *MemoryComplianceStorage) GetViolation(ctx context.Context, violationID string) (*compliance.ComplianceViolation, error) {
	mcs.mu.RLock()
	defer mcs.mu.RUnlock()

	violation, exists := mcs.violations[violationID]
	if !exists {
		return nil, fmt.Errorf("violation not found: %s", violationID)
	}

	violationCopy := *violation
	return &violationCopy, nil
}

func (mcs *MemoryComplianceStorage) UpdateViolation(ctx context.Context, violation *compliance.ComplianceViolation) error {
	mcs.mu.Lock()
	defer mcs.mu.Unlock()

	if _, exists := mcs.violations[violation.ID]; !exists {
		return fmt.Errorf("violation not found: %s", violation.ID)
	}

	violationCopy := *violation
	mcs.violations[violation.ID] = &violationCopy
	return nil
}

func (mcs *MemoryComplianceStorage) ListViolations(ctx context.Context, filters compliance.ViolationFilters) ([]*compliance.ComplianceViolation, error) {
	mcs.mu.RLock()
	defer mcs.mu.RUnlock()

	var violations []*compliance.ComplianceViolation

	for _, violation := range mcs.violations {
		if mcs.matchesViolationFilters(violation, filters) {
			violationCopy := *violation
			violations = append(violations, &violationCopy)
		}
	}

	// Apply pagination
	if filters.Offset > 0 && filters.Offset < len(violations) {
		violations = violations[filters.Offset:]
	}

	if filters.Limit > 0 && filters.Limit < len(violations) {
		violations = violations[:filters.Limit]
	}

	return violations, nil
}

// Audit Storage Implementation
func (mcs *MemoryComplianceStorage) StoreAudit(ctx context.Context, event *audit.AuditEvent) error {
	mcs.mu.Lock()
	defer mcs.mu.Unlock()

	eventCopy := *event
	mcs.auditEvents[event.ID] = &eventCopy
	return nil
}

func (mcs *MemoryComplianceStorage) GetAudit(ctx context.Context, eventID string) (*audit.AuditEvent, error) {
	mcs.mu.RLock()
	defer mcs.mu.RUnlock()

	event, exists := mcs.auditEvents[eventID]
	if !exists {
		return nil, fmt.Errorf("audit event not found: %s", eventID)
	}

	eventCopy := *event
	return &eventCopy, nil
}

func (mcs *MemoryComplianceStorage) ListAudit(ctx context.Context, filters audit.AuditFilters) ([]*audit.AuditEvent, error) {
	mcs.mu.RLock()
	defer mcs.mu.RUnlock()

	var events []*audit.AuditEvent

	for _, event := range mcs.auditEvents {
		if mcs.matchesAuditFilters(event, filters) {
			eventCopy := *event
			events = append(events, &eventCopy)
		}
	}

	// Apply pagination
	if filters.Offset > 0 && filters.Offset < len(events) {
		events = events[filters.Offset:]
	}

	if filters.Limit > 0 && filters.Limit < len(events) {
		events = events[:filters.Limit]
	}

	return events, nil
}

func (mcs *MemoryComplianceStorage) GetChain(ctx context.Context, startHash string, limit int) ([]*audit.AuditEvent, error) {
	mcs.mu.RLock()
	defer mcs.mu.RUnlock()

	var events []*audit.AuditEvent

	// Simple implementation - in production you'd want proper chain traversal
	for _, event := range mcs.auditEvents {
		if event.PreviousHash == startHash || event.Hash == startHash {
			eventCopy := *event
			events = append(events, &eventCopy)
			if len(events) >= limit {
				break
			}
		}
	}

	return events, nil
}

// Personal Data Storage Implementation
func (mcs *MemoryComplianceStorage) StorePersonalData(ctx context.Context, data *privacy.PersonalData) error {
	mcs.mu.Lock()
	defer mcs.mu.Unlock()

	dataCopy := *data
	mcs.personalData[data.ID] = &dataCopy
	return nil
}

func (mcs *MemoryComplianceStorage) GetPersonalData(ctx context.Context, dataID string) (*privacy.PersonalData, error) {
	mcs.mu.RLock()
	defer mcs.mu.RUnlock()

	data, exists := mcs.personalData[dataID]
	if !exists {
		return nil, fmt.Errorf("personal data not found: %s", dataID)
	}

	dataCopy := *data
	return &dataCopy, nil
}

func (mcs *MemoryComplianceStorage) GetBySubject(ctx context.Context, subjectID string) ([]*privacy.PersonalData, error) {
	mcs.mu.RLock()
	defer mcs.mu.RUnlock()

	var dataList []*privacy.PersonalData
	for _, data := range mcs.personalData {
		if data.SubjectID == subjectID {
			dataCopy := *data
			dataList = append(dataList, &dataCopy)
		}
	}

	return dataList, nil
}

func (mcs *MemoryComplianceStorage) UpdatePersonalData(ctx context.Context, data *privacy.PersonalData) error {
	mcs.mu.Lock()
	defer mcs.mu.Unlock()

	if _, exists := mcs.personalData[data.ID]; !exists {
		return fmt.Errorf("personal data not found: %s", data.ID)
	}

	dataCopy := *data
	mcs.personalData[data.ID] = &dataCopy
	return nil
}

func (mcs *MemoryComplianceStorage) DeletePersonalData(ctx context.Context, dataID string) error {
	mcs.mu.Lock()
	defer mcs.mu.Unlock()

	delete(mcs.personalData, dataID)
	return nil
}

func (mcs *MemoryComplianceStorage) SearchPersonalData(ctx context.Context, query privacy.DataQuery) ([]*privacy.PersonalData, error) {
	mcs.mu.RLock()
	defer mcs.mu.RUnlock()

	var results []*privacy.PersonalData

	for _, data := range mcs.personalData {
		if mcs.matchesDataQuery(data, query) {
			dataCopy := *data
			results = append(results, &dataCopy)
		}
	}

	// Apply pagination
	if query.Offset > 0 && query.Offset < len(results) {
		results = results[query.Offset:]
	}

	if query.Limit > 0 && query.Limit < len(results) {
		results = results[:query.Limit]
	}

	return results, nil
}

// Consent Storage Implementation
func (mcs *MemoryComplianceStorage) StoreConsent(ctx context.Context, consent *privacy.Consent) error {
	mcs.mu.Lock()
	defer mcs.mu.Unlock()

	consentCopy := *consent
	mcs.consents[consent.ID] = &consentCopy
	return nil
}

func (mcs *MemoryComplianceStorage) GetConsent(ctx context.Context, consentID string) (*privacy.Consent, error) {
	mcs.mu.RLock()
	defer mcs.mu.RUnlock()

	consent, exists := mcs.consents[consentID]
	if !exists {
		return nil, fmt.Errorf("consent not found: %s", consentID)
	}

	consentCopy := *consent
	return &consentCopy, nil
}

func (mcs *MemoryComplianceStorage) GetConsentsBySubject(ctx context.Context, subjectID string) ([]*privacy.Consent, error) {
	mcs.mu.RLock()
	defer mcs.mu.RUnlock()

	var consents []*privacy.Consent
	for _, consent := range mcs.consents {
		if consent.SubjectID == subjectID {
			consentCopy := *consent
			consents = append(consents, &consentCopy)
		}
	}

	return consents, nil
}

func (mcs *MemoryComplianceStorage) UpdateConsent(ctx context.Context, consent *privacy.Consent) error {
	mcs.mu.Lock()
	defer mcs.mu.Unlock()

	if _, exists := mcs.consents[consent.ID]; !exists {
		return fmt.Errorf("consent not found: %s", consent.ID)
	}

	consentCopy := *consent
	mcs.consents[consent.ID] = &consentCopy
	return nil
}

func (mcs *MemoryComplianceStorage) DeleteConsent(ctx context.Context, consentID string) error {
	mcs.mu.Lock()
	defer mcs.mu.Unlock()

	delete(mcs.consents, consentID)
	return nil
}

func (mcs *MemoryComplianceStorage) StoreHistory(ctx context.Context, history *privacy.ConsentHistory) error {
	mcs.mu.Lock()
	defer mcs.mu.Unlock()

	historyCopy := *history
	mcs.consentHistory[history.SubjectID] = append(mcs.consentHistory[history.SubjectID], &historyCopy)
	return nil
}

func (mcs *MemoryComplianceStorage) GetHistory(ctx context.Context, subjectID string) ([]*privacy.ConsentHistory, error) {
	mcs.mu.RLock()
	defer mcs.mu.RUnlock()

	history := mcs.consentHistory[subjectID]
	if history == nil {
		return []*privacy.ConsentHistory{}, nil
	}

	// Return copies
	var result []*privacy.ConsentHistory
	for _, h := range history {
		historyCopy := *h
		result = append(result, &historyCopy)
	}

	return result, nil
}

// Close closes the storage
func (mcs *MemoryComplianceStorage) Close() error {
	mcs.mu.Lock()
	defer mcs.mu.Unlock()

	// Clear all data
	mcs.rules = make(map[string]*compliance.ComplianceRule)
	mcs.violations = make(map[string]*compliance.ComplianceViolation)
	mcs.reports = make(map[string]*compliance.ComplianceReport)
	mcs.auditEvents = make(map[string]*audit.AuditEvent)
	mcs.personalData = make(map[string]*privacy.PersonalData)
	mcs.consents = make(map[string]*privacy.Consent)
	mcs.consentHistory = make(map[string][]*privacy.ConsentHistory)
	mcs.retentionPolicies = make(map[string]*privacy.RetentionPolicy)

	return nil
}

// Helper methods for filtering

func (mcs *MemoryComplianceStorage) matchesViolationFilters(violation *compliance.ComplianceViolation, filters compliance.ViolationFilters) bool {
	if filters.Regulation != "" && violation.Regulation != filters.Regulation {
		return false
	}

	if filters.Severity != "" && violation.Severity != filters.Severity {
		return false
	}

	if filters.Status != "" && violation.Status != filters.Status {
		return false
	}

	if filters.UserID != "" && violation.UserID != filters.UserID {
		return false
	}

	if filters.StartDate != nil && violation.DetectedAt.Before(*filters.StartDate) {
		return false
	}

	if filters.EndDate != nil && violation.DetectedAt.After(*filters.EndDate) {
		return false
	}

	return true
}

func (mcs *MemoryComplianceStorage) matchesAuditFilters(event *audit.AuditEvent, filters audit.AuditFilters) bool {
	if filters.UserID != "" && event.UserID != filters.UserID {
		return false
	}

	if filters.Action != "" && event.Action != filters.Action {
		return false
	}

	if filters.Resource != "" && event.Resource != filters.Resource {
		return false
	}

	if filters.Risk != "" && event.Risk != filters.Risk {
		return false
	}

	if filters.Classification != "" && event.Classification != filters.Classification {
		return false
	}

	if filters.StartDate != nil && event.Timestamp.Before(*filters.StartDate) {
		return false
	}

	if filters.EndDate != nil && event.Timestamp.After(*filters.EndDate) {
		return false
	}

	return true
}

func (mcs *MemoryComplianceStorage) matchesDataQuery(data *privacy.PersonalData, query privacy.DataQuery) bool {
	if query.SubjectID != "" && data.SubjectID != query.SubjectID {
		return false
	}

	if query.DataType != "" && data.DataType != query.DataType {
		return false
	}

	if query.Category != "" && data.Category != query.Category {
		return false
	}

	if query.Purpose != "" && data.Purpose != query.Purpose {
		return false
	}

	if query.StartDate != nil && data.CreatedAt.Before(*query.StartDate) {
		return false
	}

	if query.EndDate != nil && data.CreatedAt.After(*query.EndDate) {
		return false
	}

	return true
}
