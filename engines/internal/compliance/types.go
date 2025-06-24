package compliance

import (
	"context"
	"time"
)

// ComplianceLevel represents the level of compliance required
type ComplianceLevel string

const (
	ComplianceLevelBasic    ComplianceLevel = "basic"
	ComplianceLevelStandard ComplianceLevel = "standard"
	ComplianceLevelStrict   ComplianceLevel = "strict"
	ComplianceLevelMaximum  ComplianceLevel = "maximum"
)

// RegulationType represents different regulatory frameworks
type RegulationType string

const (
	RegulationGDPR  RegulationType = "gdpr"
	RegulationCCPA  RegulationType = "ccpa"
	RegulationHIPAA RegulationType = "hipaa"
	RegulationSOX   RegulationType = "sox"
	RegulationPCI   RegulationType = "pci"
	RegulationKYC   RegulationType = "kyc"
	RegulationAML   RegulationType = "aml"
)

// DataClassification represents the sensitivity level of data
type DataClassification string

const (
	DataPublic       DataClassification = "public"
	DataInternal     DataClassification = "internal"
	DataConfidential DataClassification = "confidential"
	DataRestricted   DataClassification = "restricted"
	DataTopSecret    DataClassification = "top_secret"
)

// ComplianceStatus represents the compliance status of an entity
type ComplianceStatus string

const (
	StatusCompliant    ComplianceStatus = "compliant"
	StatusNonCompliant ComplianceStatus = "non_compliant"
	StatusPending      ComplianceStatus = "pending"
	StatusUnknown      ComplianceStatus = "unknown"
)

// ComplianceEvent represents a compliance-related event
type ComplianceEvent struct {
	ID             string                 `json:"id"`
	Type           string                 `json:"type"`
	EntityType     string                 `json:"entity_type"`
	EntityID       string                 `json:"entity_id"`
	UserID         string                 `json:"user_id"`
	Action         string                 `json:"action"`
	Timestamp      time.Time              `json:"timestamp"`
	Regulation     RegulationType         `json:"regulation"`
	Classification DataClassification     `json:"classification"`
	Details        map[string]interface{} `json:"details"`
	Risk           RiskLevel              `json:"risk"`
	Status         ComplianceStatus       `json:"status"`
	Metadata       map[string]interface{} `json:"metadata"`
}

// RiskLevel represents the risk level of an action or event
type RiskLevel string

const (
	RiskLow      RiskLevel = "low"
	RiskMedium   RiskLevel = "medium"
	RiskHigh     RiskLevel = "high"
	RiskCritical RiskLevel = "critical"
)

// ComplianceRule represents a compliance rule
type ComplianceRule struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Regulation  RegulationType         `json:"regulation"`
	Category    string                 `json:"category"`
	Severity    RiskLevel              `json:"severity"`
	Conditions  []RuleCondition        `json:"conditions"`
	Actions     []RuleAction           `json:"actions"`
	Enabled     bool                   `json:"enabled"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// RuleCondition represents a condition for a compliance rule
type RuleCondition struct {
	Field    string      `json:"field"`
	Operator string      `json:"operator"`
	Value    interface{} `json:"value"`
}

// RuleAction represents an action to take when a rule is triggered
type RuleAction struct {
	Type       string                 `json:"type"`
	Parameters map[string]interface{} `json:"parameters"`
}

// ComplianceViolation represents a compliance violation
type ComplianceViolation struct {
	ID          string                 `json:"id"`
	RuleID      string                 `json:"rule_id"`
	EventID     string                 `json:"event_id"`
	UserID      string                 `json:"user_id"`
	EntityType  string                 `json:"entity_type"`
	EntityID    string                 `json:"entity_id"`
	Regulation  RegulationType         `json:"regulation"`
	Severity    RiskLevel              `json:"severity"`
	Description string                 `json:"description"`
	Status      ViolationStatus        `json:"status"`
	DetectedAt  time.Time              `json:"detected_at"`
	ResolvedAt  *time.Time             `json:"resolved_at,omitempty"`
	Details     map[string]interface{} `json:"details"`
	Resolution  string                 `json:"resolution,omitempty"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// ViolationStatus represents the status of a compliance violation
type ViolationStatus string

const (
	ViolationOpen     ViolationStatus = "open"
	ViolationResolved ViolationStatus = "resolved"
	ViolationIgnored  ViolationStatus = "ignored"
)

// ComplianceReport represents a compliance report
type ComplianceReport struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Regulation  RegulationType         `json:"regulation"`
	Period      ReportPeriod           `json:"period"`
	Status      ReportStatus           `json:"status"`
	GeneratedAt time.Time              `json:"generated_at"`
	Data        map[string]interface{} `json:"data"`
	Violations  []ComplianceViolation  `json:"violations"`
	Metrics     ReportMetrics          `json:"metrics"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// ReportPeriod represents the time period for a report
type ReportPeriod struct {
	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`
}

// ReportStatus represents the status of a report
type ReportStatus string

const (
	ReportGenerating ReportStatus = "generating"
	ReportCompleted  ReportStatus = "completed"
	ReportFailed     ReportStatus = "failed"
)

// ReportMetrics represents metrics in a compliance report
type ReportMetrics struct {
	TotalEvents      int64                  `json:"total_events"`
	TotalViolations  int64                  `json:"total_violations"`
	ViolationsByType map[string]int64       `json:"violations_by_type"`
	RiskDistribution map[RiskLevel]int64    `json:"risk_distribution"`
	ComplianceScore  float64                `json:"compliance_score"`
	Trends           map[string]interface{} `json:"trends"`
}

// ComplianceEngine defines the main compliance interface
type ComplianceEngine interface {
	// Event processing
	ProcessEvent(ctx context.Context, event *ComplianceEvent) error

	// Rule management
	AddRule(ctx context.Context, rule *ComplianceRule) error
	UpdateRule(ctx context.Context, rule *ComplianceRule) error
	DeleteRule(ctx context.Context, ruleID string) error
	GetRule(ctx context.Context, ruleID string) (*ComplianceRule, error)
	ListRules(ctx context.Context, regulation RegulationType) ([]*ComplianceRule, error)

	// Violation management
	GetViolations(ctx context.Context, filters ViolationFilters) ([]*ComplianceViolation, error)
	ResolveViolation(ctx context.Context, violationID string, resolution string) error

	// Reporting
	GenerateReport(ctx context.Context, reportType string, regulation RegulationType, period ReportPeriod) (*ComplianceReport, error)
	GetReport(ctx context.Context, reportID string) (*ComplianceReport, error)

	// Compliance checking
	CheckCompliance(ctx context.Context, entityType string, entityID string) (ComplianceStatus, error)

	// Health check
	HealthCheck(ctx context.Context) error
}

// ViolationFilters represents filters for querying violations
type ViolationFilters struct {
	Regulation RegulationType  `json:"regulation,omitempty"`
	Severity   RiskLevel       `json:"severity,omitempty"`
	Status     ViolationStatus `json:"status,omitempty"`
	UserID     string          `json:"user_id,omitempty"`
	StartDate  *time.Time      `json:"start_date,omitempty"`
	EndDate    *time.Time      `json:"end_date,omitempty"`
	Limit      int             `json:"limit,omitempty"`
	Offset     int             `json:"offset,omitempty"`
}
