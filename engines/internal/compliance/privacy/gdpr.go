package privacy

import (
	"context"
	"fmt"
	"time"

	"github.com/GoSec-Labs/mMAD/engines/pkg/logger"
)

// GDPRProcessor handles GDPR compliance
type GDPRProcessor struct {
	dataStore    DataStore
	encryptor    DataEncryptor
	anonymizer   DataAnonymizer
	consentMgr   ConsentManager
	retentionMgr RetentionManager
}

// DataStore defines the interface for data storage
type DataStore interface {
	Store(ctx context.Context, data *PersonalData) error
	Get(ctx context.Context, dataID string) (*PersonalData, error)
	GetBySubject(ctx context.Context, subjectID string) ([]*PersonalData, error)
	Update(ctx context.Context, data *PersonalData) error
	Delete(ctx context.Context, dataID string) error
	Search(ctx context.Context, query DataQuery) ([]*PersonalData, error)
	Close() error
}

// PersonalData represents personal data subject to GDPR
type PersonalData struct {
	ID            string                 `json:"id"`
	SubjectID     string                 `json:"subject_id"`
	DataType      string                 `json:"data_type"`
	Category      DataCategory           `json:"category"`
	Purpose       ProcessingPurpose      `json:"purpose"`
	LegalBasis    LegalBasis             `json:"legal_basis"`
	Data          map[string]interface{} `json:"data"`
	EncryptedData []byte                 `json:"encrypted_data,omitempty"`
	IsEncrypted   bool                   `json:"is_encrypted"`
	IsAnonymized  bool                   `json:"is_anonymized"`
	ConsentID     string                 `json:"consent_id"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
	ExpiresAt     *time.Time             `json:"expires_at,omitempty"`
	Metadata      map[string]interface{} `json:"metadata"`
}

// DataCategory represents categories of personal data
type DataCategory string

const (
	CategoryIdentifying DataCategory = "identifying"
	CategorySensitive   DataCategory = "sensitive"
	CategoryBehavioral  DataCategory = "behavioral"
	CategoryFinancial   DataCategory = "financial"
	CategoryHealth      DataCategory = "health"
	CategoryBiometric   DataCategory = "biometric"
)

// ProcessingPurpose represents the purpose of data processing
type ProcessingPurpose string

const (
	PurposeConsent            ProcessingPurpose = "consent"
	PurposeContract           ProcessingPurpose = "contract"
	PurposeLegalObligation    ProcessingPurpose = "legal_obligation"
	PurposeVitalInterests     ProcessingPurpose = "vital_interests"
	PurposePublicTask         ProcessingPurpose = "public_task"
	PurposeLegitimateInterest ProcessingPurpose = "legitimate_interest"
)

// LegalBasis represents the legal basis for processing
type LegalBasis string

const (
	BasisConsent            LegalBasis = "consent"
	BasisContract           LegalBasis = "contract"
	BasisLegalObligation    LegalBasis = "legal_obligation"
	BasisVitalInterests     LegalBasis = "vital_interests"
	BasisPublicTask         LegalBasis = "public_task"
	BasisLegitimateInterest LegalBasis = "legitimate_interest"
)

// DataQuery represents a query for personal data
type DataQuery struct {
	SubjectID string            `json:"subject_id,omitempty"`
	DataType  string            `json:"data_type,omitempty"`
	Category  DataCategory      `json:"category,omitempty"`
	Purpose   ProcessingPurpose `json:"purpose,omitempty"`
	StartDate *time.Time        `json:"start_date,omitempty"`
	EndDate   *time.Time        `json:"end_date,omitempty"`
	Limit     int               `json:"limit,omitempty"`
	Offset    int               `json:"offset,omitempty"`
}

// NewGDPRProcessor creates a new GDPR processor
func NewGDPRProcessor(
	dataStore DataStore,
	encryptor DataEncryptor,
	anonymizer DataAnonymizer,
	consentMgr ConsentManager,
	retentionMgr RetentionManager,
) *GDPRProcessor {
	return &GDPRProcessor{
		dataStore:    dataStore,
		encryptor:    encryptor,
		anonymizer:   anonymizer,
		consentMgr:   consentMgr,
		retentionMgr: retentionMgr,
	}
}

// ProcessDataSubjectRequest handles GDPR data subject requests
func (gp *GDPRProcessor) ProcessDataSubjectRequest(ctx context.Context, request *DataSubjectRequest) (*DataSubjectResponse, error) {
	switch request.Type {
	case RequestTypeAccess:
		return gp.handleAccessRequest(ctx, request)
	case RequestTypeRectification:
		return gp.handleRectificationRequest(ctx, request)
	case RequestTypeErasure:
		return gp.handleErasureRequest(ctx, request)
	case RequestTypePortability:
		return gp.handlePortabilityRequest(ctx, request)
	case RequestTypeRestriction:
		return gp.handleRestrictionRequest(ctx, request)
	case RequestTypeObjection:
		return gp.handleObjectionRequest(ctx, request)
	default:
		return nil, fmt.Errorf("unsupported request type: %s", request.Type)
	}
}

// handleAccessRequest handles right of access requests
func (gp *GDPRProcessor) handleAccessRequest(ctx context.Context, request *DataSubjectRequest) (*DataSubjectResponse, error) {
	// Get all data for the subject
	personalData, err := gp.dataStore.GetBySubject(ctx, request.SubjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get personal data: %w", err)
	}

	// Decrypt data if needed
	for _, data := range personalData {
		if data.IsEncrypted {
			decryptedData, err := gp.encryptor.Decrypt(data.EncryptedData)
			if err != nil {
				logger.Error("Failed to decrypt data for access request", "data_id", data.ID, "error", err)
				continue
			}
			data.Data = decryptedData
			data.IsEncrypted = false
		}
	}

	// Prepare response
	response := &DataSubjectResponse{
		ID:          fmt.Sprintf("dsr_%d", time.Now().Unix()),
		RequestID:   request.ID,
		Type:        request.Type,
		Status:      ResponseStatusCompleted,
		Data:        personalData,
		ProcessedAt: time.Now(),
	}

	logger.Info("Access request processed",
		"subject_id", request.SubjectID,
		"data_count", len(personalData))

	return response, nil
}

// handleErasureRequest handles right to erasure requests
func (gp *GDPRProcessor) handleErasureRequest(ctx context.Context, request *DataSubjectRequest) (*DataSubjectResponse, error) {
	// Get all data for the subject
	personalData, err := gp.dataStore.GetBySubject(ctx, request.SubjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get personal data: %w", err)
	}

	var erasedCount int
	var errors []string

	for _, data := range personalData {
		// Check if data can be erased
		if !gp.canEraseData(data) {
			errors = append(errors, fmt.Sprintf("cannot erase data %s due to legal obligations", data.ID))
			continue
		}

		// Delete the data
		if err := gp.dataStore.Delete(ctx, data.ID); err != nil {
			errors = append(errors, fmt.Sprintf("failed to delete data %s: %v", data.ID, err))
			continue
		}

		erasedCount++
	}

	status := ResponseStatusCompleted
	if len(errors) > 0 {
		status = ResponseStatusPartiallyCompleted
	}

	response := &DataSubjectResponse{
		ID:          fmt.Sprintf("dsr_%d", time.Now().Unix()),
		RequestID:   request.ID,
		Type:        request.Type,
		Status:      status,
		ProcessedAt: time.Now(),
		Metadata: map[string]interface{}{
			"erased_count": erasedCount,
			"errors":       errors,
		},
	}

	logger.Info("Erasure request processed",
		"subject_id", request.SubjectID,
		"erased_count", erasedCount,
		"errors", len(errors))

	return response, nil
}

// canEraseData checks if data can be erased
func (gp *GDPRProcessor) canEraseData(data *PersonalData) bool {
	// Check legal basis
	if data.LegalBasis == BasisLegalObligation {
		return false
	}

	// Check if consent is withdrawn
	if data.LegalBasis == BasisConsent {
		consent, err := gp.consentMgr.GetConsent(context.Background(), data.ConsentID)
		if err != nil || !consent.IsActive {
			return true
		}
	}

	// Check retention period
	if data.ExpiresAt != nil && time.Now().After(*data.ExpiresAt) {
		return true
	}

	return false
}

// handlePortabilityRequest handles data portability requests
func (gp *GDPRProcessor) handlePortabilityRequest(ctx context.Context, request *DataSubjectRequest) (*DataSubjectResponse, error) {
	// Get portable data (structured, commonly used formats)
	query := DataQuery{
		SubjectID: request.SubjectID,
	}

	personalData, err := gp.dataStore.Search(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to search personal data: %w", err)
	}

	// Filter for portable data
	var portableData []*PersonalData
	for _, data := range personalData {
		if gp.isDataPortable(data) {
			portableData = append(portableData, data)
		}
	}

	// Format data for portability
	exportData := gp.formatForExport(portableData)

	response := &DataSubjectResponse{
		ID:          fmt.Sprintf("dsr_%d", time.Now().Unix()),
		RequestID:   request.ID,
		Type:        request.Type,
		Status:      ResponseStatusCompleted,
		Data:        portableData,
		ExportData:  exportData,
		ProcessedAt: time.Now(),
	}

	logger.Info("Portability request processed",
		"subject_id", request.SubjectID,
		"portable_count", len(portableData))

	return response, nil
}

// isDataPortable checks if data is portable under GDPR
func (gp *GDPRProcessor) isDataPortable(data *PersonalData) bool {
	// Data must be processed based on consent or contract
	if data.LegalBasis != BasisConsent && data.LegalBasis != BasisContract {
		return false
	}

	// Data must be provided by the data subject
	portableCategories := []DataCategory{
		CategoryIdentifying,
		CategoryBehavioral,
		CategoryFinancial,
	}

	for _, category := range portableCategories {
		if data.Category == category {
			return true
		}
	}

	return false
}

// formatForExport formats data for export in standard formats
func (gp *GDPRProcessor) formatForExport(data []*PersonalData) map[string]interface{} {
	exportData := make(map[string]interface{})

	// Group by data type
	byType := make(map[string][]map[string]interface{})

	for _, item := range data {
		if item.IsEncrypted {
			// Decrypt for export
			if decryptedData, err := gp.encryptor.Decrypt(item.EncryptedData); err == nil {
				item.Data = decryptedData
			}
		}

		// Clean data for export
		cleanData := map[string]interface{}{
			"id":         item.ID,
			"created_at": item.CreatedAt,
			"updated_at": item.UpdatedAt,
			"data":       item.Data,
		}

		byType[item.DataType] = append(byType[item.DataType], cleanData)
	}

	exportData["data_by_type"] = byType
	exportData["export_timestamp"] = time.Now()
	exportData["format_version"] = "1.0"

	return exportData
}

// handleRectificationRequest handles rectification requests
func (gp *GDPRProcessor) handleRectificationRequest(ctx context.Context, request *DataSubjectRequest) (*DataSubjectResponse, error) {
	if request.UpdateData == nil {
		return nil, fmt.Errorf("rectification request must include update data")
	}

	var updatedCount int
	var errors []string

	for dataID, updates := range request.UpdateData {
		// Get existing data
		data, err := gp.dataStore.Get(ctx, dataID)
		if err != nil {
			errors = append(errors, fmt.Sprintf("failed to get data %s: %v", dataID, err))
			continue
		}

		// Verify subject ownership
		if data.SubjectID != request.SubjectID {
			errors = append(errors, fmt.Sprintf("data %s does not belong to subject", dataID))
			continue
		}

		// Apply updates
		for field, value := range updates {
			data.Data[field] = value
		}
		data.UpdatedAt = time.Now()

		// Re-encrypt if necessary
		if data.IsEncrypted {
			encryptedData, err := gp.encryptor.Encrypt(data.Data)
			if err != nil {
				errors = append(errors, fmt.Sprintf("failed to encrypt updated data %s: %v", dataID, err))
				continue
			}
			data.EncryptedData = encryptedData
		}

		// Update in store
		if err := gp.dataStore.Update(ctx, data); err != nil {
			errors = append(errors, fmt.Sprintf("failed to update data %s: %v", dataID, err))
			continue
		}

		updatedCount++
	}

	status := ResponseStatusCompleted
	if len(errors) > 0 {
		status = ResponseStatusPartiallyCompleted
	}

	response := &DataSubjectResponse{
		ID:          fmt.Sprintf("dsr_%d", time.Now().Unix()),
		RequestID:   request.ID,
		Type:        request.Type,
		Status:      status,
		ProcessedAt: time.Now(),
		Metadata: map[string]interface{}{
			"updated_count": updatedCount,
			"errors":        errors,
		},
	}

	logger.Info("Rectification request processed",
		"subject_id", request.SubjectID,
		"updated_count", updatedCount,
		"errors", len(errors))

	return response, nil
}

// handleRestrictionRequest handles processing restriction requests
func (gp *GDPRProcessor) handleRestrictionRequest(ctx context.Context, request *DataSubjectRequest) (*DataSubjectResponse, error) {
	personalData, err := gp.dataStore.GetBySubject(ctx, request.SubjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get personal data: %w", err)
	}

	var restrictedCount int
	var errors []string

	for _, data := range personalData {
		// Mark as restricted
		if data.Metadata == nil {
			data.Metadata = make(map[string]interface{})
		}
		data.Metadata["processing_restricted"] = true
		data.Metadata["restriction_date"] = time.Now()
		data.UpdatedAt = time.Now()

		if err := gp.dataStore.Update(ctx, data); err != nil {
			errors = append(errors, fmt.Sprintf("failed to restrict data %s: %v", data.ID, err))
			continue
		}

		restrictedCount++
	}

	status := ResponseStatusCompleted
	if len(errors) > 0 {
		status = ResponseStatusPartiallyCompleted
	}

	response := &DataSubjectResponse{
		ID:          fmt.Sprintf("dsr_%d", time.Now().Unix()),
		RequestID:   request.ID,
		Type:        request.Type,
		Status:      status,
		ProcessedAt: time.Now(),
		Metadata: map[string]interface{}{
			"restricted_count": restrictedCount,
			"errors":           errors,
		},
	}

	return response, nil
}

// handleObjectionRequest handles objection requests
func (gp *GDPRProcessor) handleObjectionRequest(ctx context.Context, request *DataSubjectRequest) (*DataSubjectResponse, error) {
	personalData, err := gp.dataStore.GetBySubject(ctx, request.SubjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get personal data: %w", err)
	}

	var objectedCount int
	var errors []string

	for _, data := range personalData {
		// Check if objection can be honored
		if data.LegalBasis == BasisLegitimateInterest {
			// Mark as objected
			if data.Metadata == nil {
				data.Metadata = make(map[string]interface{})
			}
			data.Metadata["processing_objected"] = true
			data.Metadata["objection_date"] = time.Now()
			data.UpdatedAt = time.Now()

			if err := gp.dataStore.Update(ctx, data); err != nil {
				errors = append(errors, fmt.Sprintf("failed to object data %s: %v", data.ID, err))
				continue
			}

			objectedCount++
		}
	}

	response := &DataSubjectResponse{
		ID:          fmt.Sprintf("dsr_%d", time.Now().Unix()),
		RequestID:   request.ID,
		Type:        request.Type,
		Status:      ResponseStatusCompleted,
		ProcessedAt: time.Now(),
		Metadata: map[string]interface{}{
			"objected_count": objectedCount,
			"errors":         errors,
		},
	}

	return response, nil
}

// DataSubjectRequest represents a GDPR data subject request
type DataSubjectRequest struct {
	ID         string                            `json:"id"`
	Type       RequestType                       `json:"type"`
	SubjectID  string                            `json:"subject_id"`
	UpdateData map[string]map[string]interface{} `json:"update_data,omitempty"`
	Reason     string                            `json:"reason,omitempty"`
	CreatedAt  time.Time                         `json:"created_at"`
	Metadata   map[string]interface{}            `json:"metadata"`
}

// RequestType represents the type of data subject request
type RequestType string

const (
	RequestTypeAccess        RequestType = "access"
	RequestTypeRectification RequestType = "rectification"
	RequestTypeErasure       RequestType = "erasure"
	RequestTypePortability   RequestType = "portability"
	RequestTypeRestriction   RequestType = "restriction"
	RequestTypeObjection     RequestType = "objection"
)

// DataSubjectResponse represents the response to a data subject request
type DataSubjectResponse struct {
	ID          string                 `json:"id"`
	RequestID   string                 `json:"request_id"`
	Type        RequestType            `json:"type"`
	Status      ResponseStatus         `json:"status"`
	Data        []*PersonalData        `json:"data,omitempty"`
	ExportData  map[string]interface{} `json:"export_data,omitempty"`
	ProcessedAt time.Time              `json:"processed_at"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// ResponseStatus represents the status of a data subject response
type ResponseStatus string

const (
	ResponseStatusPending            ResponseStatus = "pending"
	ResponseStatusCompleted          ResponseStatus = "completed"
	ResponseStatusPartiallyCompleted ResponseStatus = "partially_completed"
	ResponseStatusFailed             ResponseStatus = "failed"
)
