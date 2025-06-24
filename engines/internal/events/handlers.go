package events

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/GoSec-Labs/mMAD/engines/pkg/logger"
)

// LoggerHandler logs all events
type LoggerHandler struct {
	logLevel string
}

// NewLoggerHandler creates a new logger handler
func NewLoggerHandler(logLevel string) *LoggerHandler {
	return &LoggerHandler{
		logLevel: logLevel,
	}
}

func (h *LoggerHandler) Handle(event Event) error {
	logger.Info("Event occurred",
		"type", event.Type,
		"id", event.ID,
		"source", event.Source,
		"timestamp", event.Timestamp,
		"user_id", event.UserID,
		"request_id", event.RequestID,
		"data", event.Data)

	return nil
}

func (h *LoggerHandler) CanHandle(eventType EventType) bool {
	return true // Handle all events
}

// MetricsHandler tracks event metrics
type MetricsHandler struct {
	eventCounts map[EventType]int64
	errorCounts map[EventType]int64
	lastSeen    map[EventType]time.Time
}

// NewMetricsHandler creates a new metrics handler
func NewMetricsHandler() *MetricsHandler {
	return &MetricsHandler{
		eventCounts: make(map[EventType]int64),
		errorCounts: make(map[EventType]int64),
		lastSeen:    make(map[EventType]time.Time),
	}
}

func (h *MetricsHandler) Handle(event Event) error {
	h.eventCounts[event.Type]++
	h.lastSeen[event.Type] = event.Timestamp

	// Track errors
	if event.Type == EventProofFailed || event.Type == EventError {
		h.errorCounts[event.Type]++
	}

	return nil
}

func (h *MetricsHandler) CanHandle(eventType EventType) bool {
	return true
}

// GetMetrics returns current metrics
func (h *MetricsHandler) GetMetrics() map[string]interface{} {
	return map[string]interface{}{
		"event_counts": h.eventCounts,
		"error_counts": h.errorCounts,
		"last_seen":    h.lastSeen,
	}
}

// AlertHandler sends alerts for critical events
type AlertHandler struct {
	alertThreshold int
	errorCount     int
}

// NewAlertHandler creates a new alert handler
func NewAlertHandler(threshold int) *AlertHandler {
	return &AlertHandler{
		alertThreshold: threshold,
	}
}

func (h *AlertHandler) Handle(event Event) error {
	switch event.Type {
	case EventProofFailed, EventError:
		h.errorCount++

		if h.errorCount >= h.alertThreshold {
			logger.Error("Alert: High error rate detected",
				"error_count", h.errorCount,
				"threshold", h.alertThreshold,
				"latest_event", event.ID)

			// In a real system, you'd send notifications here
			h.errorCount = 0 // Reset counter
		}

	case EventProofGenerated:
		// Reset error count on successful operations
		if h.errorCount > 0 {
			h.errorCount--
		}
	}

	return nil
}

func (h *AlertHandler) CanHandle(eventType EventType) bool {
	return eventType == EventProofFailed ||
		eventType == EventError ||
		eventType == EventProofGenerated
}

// JSONHandler outputs events as JSON
type JSONHandler struct {
	output func(string) error
}

// NewJSONHandler creates a new JSON handler
func NewJSONHandler(outputFunc func(string) error) *JSONHandler {
	return &JSONHandler{
		output: outputFunc,
	}
}

func (h *JSONHandler) Handle(event Event) error {
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	return h.output(string(data))
}

func (h *JSONHandler) CanHandle(eventType EventType) bool {
	return true
}
