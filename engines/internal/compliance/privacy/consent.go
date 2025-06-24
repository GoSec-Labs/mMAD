package privacy

import (
	"context"
	"fmt"
	"time"

	"github.com/GoSec-Labs/mMAD/engines/pkg/logger"
)

// ConsentManager manages user consent
type ConsentManager interface {
    CreateConsent(ctx context.Context, consent *Consent) error
    GetConsent(ctx context.Context, consentID string) (*Consent, error)
    GetConsentsBySubject(ctx context.Context, subjectID string) ([]*Consent, error)
    UpdateConsent(ctx context.Context, consent *Consent) error
    WithdrawConsent(ctx context.Context, consentID string) error
    IsValidConsent(ctx context.Context, subjectID string, purpose ProcessingPurpose) (bool, error)
    GetConsentHistory(ctx context.Context, subjectID string) ([]*ConsentHistory, error)
}

// SimpleConsentManager implements ConsentManager
type SimpleConsentManager struct {
    storage ConsentStorage
}

// ConsentStorage defines the interface for consent storage
type ConsentStorage interface {
    Store(ctx context.Context, consent *Consent) error
    Get(ctx context.Context, consentID string) (*Consent, error)
    GetBySubject(ctx context.Context, subjectID string) ([]*Consent, error)
    Update(ctx context.Context, consent *Consent) error
    Delete(ctx context.Context, consentID string) error
    StoreHistory(ctx context.Context, history *ConsentHistory) error
    GetHistory(ctx context.Context, subjectID string) ([]*ConsentHistory, error)
    Close() error
}

// Consent represents user consent
type Consent struct {
    ID                string              `json:"id"`
    SubjectID         string              `json:"subject_id"`
    Purpose           ProcessingPurpose   `json:"purpose"`
    LegalBasis        LegalBasis          `json:"legal_basis"`
    DataCategories    []DataCategory      `json:"data_categories"`
    IsActive          bool                `json:"is_active"`
    GrantedAt         time.Time           `json:"granted_at"`
    WithdrawnAt       *time.Time          `json:"withdrawn_at,omitempty"`
    ExpiresAt         *time.Time          `json:"expires_at,omitempty"`
    ConsentMethod     ConsentMethod       `json:"consent_method"`
    ConsentText       string              `json:"consent_text"`
    ConsentVersion    string              `json:"consent_version"`
    IPAddress         string              `json:"ip_address,omitempty"`
    UserAgent         string              `json:"user_agent,omitempty"`
    Metadata          map[string]interface{} `json:"metadata"`
}

// ConsentMethod represents how consent was obtained
type ConsentMethod string

const (
    ConsentMethodOptIn        ConsentMethod = "opt_in"
    ConsentMethodOptOut       ConsentMethod = "opt_out"
    ConsentMethodImplied      ConsentMethod = "implied"
    ConsentMethodExpress     ConsentMethod = "express"
    ConsentMethodDigital     ConsentMethod = "digital"
    ConsentMethodPhysical    ConsentMethod = "physical"
)

// ConsentHistory represents the history of consent changes
type ConsentHistory struct {
    ID          string                 `json:"id"`
    ConsentID   string                 `json:"consent_id"`
    SubjectID   string                 `json:"subject_id"`
    Action      ConsentAction          `json:"action"`
    Timestamp   time.Time              `json:"timestamp"`
    Details     map[string]interface{} `json:"details"`
    IPAddress   string                 `json:"ip_address,omitempty"`
    UserAgent   string                 `json:"user_agent,omitempty"`
}

// ConsentAction represents an action taken on consent
type ConsentAction string

const (
    ConsentActionGranted    ConsentAction = "granted"
    ConsentActionWithdrawn  ConsentAction = "withdrawn"
    ConsentActionModified   ConsentAction = "modified"
    ConsentActionExpired    ConsentAction = "expired"
)

// NewSimpleConsentManager creates a new consent manager
func NewSimpleConsentManager(storage ConsentStorage) *SimpleConsentManager {
    return &SimpleConsentManager{
        storage: storage,
    }
}

// CreateConsent creates a new consent record
func (scm *SimpleConsentManager) CreateConsent(ctx context.Context, consent *Consent) error {
    // Set defaults
    if consent.GrantedAt.IsZero() {
        consent.GrantedAt = time.Now()
    }
    consent.IsActive = true
    
    // Store consent
    if err := scm.storage.Store(ctx, consent); err != nil {
        return fmt.Errorf("failed to store consent: %w", err)
    }
    
    // Record history
    history := &ConsentHistory{
        ID:        fmt.Sprintf("history_%d", time.Now().UnixNano()),
        ConsentID: consent.ID,
        SubjectID: consent.SubjectID,
        Action:    ConsentActionGranted,
        Timestamp: consent.GrantedAt,
        Details: map[string]interface{}{
            "purpose":           consent.Purpose,
            "legal_basis":       consent.LegalBasis,
            "data_categories":   consent.DataCategories,
            "consent_method":    consent.ConsentMethod,
            "consent_version":   consent.ConsentVersion,
        },
        IPAddress: consent.IPAddress,
        UserAgent: consent.UserAgent,
    }
    
    if err := scm.storage.StoreHistory(ctx, history); err != nil {
        logger.Error("Failed to store consent history", "error", err)
    }
    
    logger.Info("Consent created", 
        "consent_id", consent.ID,
        "subject_id", consent.SubjectID,
        "purpose", consent.Purpose)
    
    return nil
}

// GetConsent retrieves a consent record by ID
func (scm *SimpleConsentManager) GetConsent(ctx context.Context, consentID string) (*Consent, error) {
    consent, err := scm.storage.Get(ctx, consentID)
    if err != nil {
        return nil, fmt.Errorf("failed to get consent: %w", err)
    }
    
    // Check if expired
    if consent.ExpiresAt != nil && time.Now().After(*consent.ExpiresAt) {
        consent.IsActive = false
        scm.markExpired(ctx, consent)
    }
    
    return consent, nil
}

// GetConsentsBySubject retrieves all consents for a subject
func (scm *SimpleConsentManager) GetConsentsBySubject(ctx context.Context, subjectID string) ([]*Consent, error) {
    consents, err := scm.storage.GetBySubject(ctx, subjectID)
    if err != nil {
        return nil, fmt.Errorf("failed to get consents by subject: %w", err)
    }
    
    // Check for expired consents
    for _, consent := range consents {
        if consent.ExpiresAt != nil && time.Now().After(*consent.ExpiresAt) && consent.IsActive {
            consent.IsActive = false
            scm.markExpired(ctx, consent)
        }
    }
    
    return consents, nil
}

// UpdateConsent updates a consent record
func (scm *SimpleConsentManager) UpdateConsent(ctx context.Context, consent *Consent) error {
    if err := scm.storage.Update(ctx, consent); err != nil {
        return fmt.Errorf("failed to update consent: %w", err)
    }
    
    // Record history
    history := &ConsentHistory{
        ID:        fmt.Sprintf("history_%d", time.Now().UnixNano()),
        ConsentID: consent.ID,
        SubjectID: consent.SubjectID,
        Action:    ConsentActionModified,
        Timestamp: time.Now(),
        Details: map[string]interface{}{
            "purpose":           consent.Purpose,
            "legal_basis":       consent.LegalBasis,
            "data_categories":   consent.DataCategories,
            "is_active":         consent.IsActive,
        },
    }
    
    if err := scm.storage.StoreHistory(ctx, history); err != nil {
        logger.Error("Failed to store consent history", "error", err)
    }
    
    return nil
}

// WithdrawConsent withdraws consent
func (scm *SimpleConsentManager) WithdrawConsent(ctx context.Context, consentID string) error {
    consent, err := scm.GetConsent(ctx, consentID)
    if err != nil {
        return fmt.Errorf("failed to get consent for withdrawal: %w", err)
    }
    
    // Mark as withdrawn
    consent.IsActive = false
    now := time.Now()
    consent.WithdrawnAt = &now
    
    if err := scm.storage.Update(ctx, consent); err != nil {
        return fmt.Errorf("failed to update consent: %w", err)
    }
    
    // Record history
    history := &ConsentHistory{
        ID:        fmt.Sprintf("history_%d", time.Now().UnixNano()),
        ConsentID: consent.ID,
        SubjectID: consent.SubjectID,
        Action:    ConsentActionWithdrawn,
        Timestamp: now,
        Details: map[string]interface{}{
            "withdrawal_reason": "user_requested",
        },
    }
    
    if err := scm.storage.StoreHistory(ctx, history); err != nil {
        logger.Error("Failed to store consent history", "error", err)
    }
    
    logger.Info("Consent withdrawn", 
        "consent_id", consentID,
        "subject_id", consent.SubjectID)
    
    return nil
}

// IsValidConsent checks if there's valid consent for a purpose
func (scm *SimpleConsentManager) IsValidConsent(ctx context.Context, subjectID string, purpose ProcessingPurpose) (bool, error) {
    consents, err := scm.GetConsentsBySubject(ctx, subjectID)
    if err != nil {
        return false, fmt.Errorf("failed to get consents: %w", err)
    }
    
    for _, consent := range consents {
        if consent.Purpose == purpose && consent.IsActive {
            // Check if not expired
            if consent.ExpiresAt == nil || time.Now().Before(*consent.ExpiresAt) {
                return true, nil
            }
        }
    }
    
    return false, nil
}

// GetConsentHistory retrieves consent history for a subject
func (scm *SimpleConsentManager) GetConsentHistory(ctx context.Context, subjectID string) ([]*ConsentHistory, error) {
    return scm.storage.GetHistory(ctx, subjectID)
}

// markExpired marks a consent as expired
func (scm *SimpleConsentManager) markExpired(ctx context.Context, consent *Consent) {
    consent.IsActive = false
    
    if err := scm.storage.Update(ctx, consent); err != nil {
        logger.Error("Failed to mark consent as expired", "consent_id", consent.ID, "error", err)
        return
    }
    
    // Record expiry in history
    history := &ConsentHistory{
        ID:        fmt.Sprintf("history_%d", time.Now().UnixNano()),
        ConsentID: consent.ID,
        SubjectID: consent.SubjectID,
        Action:    ConsentActionExpired,
        Timestamp: time.Now(),
        Details: map[string]interface{}{
            "expired_at": consent.ExpiresAt,
        },
    }
    
    if err := scm.storage.StoreHistory(ctx, history); err != nil {
        logger.Error("Failed to store consent expiry history", "error", err)
    }
    
    logger.Info("Consent marked as expired", 
        "consent_id", consent.ID,
        "subject_id", consent.SubjectID)
}
