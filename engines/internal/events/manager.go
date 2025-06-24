package events

import "github.com/GoSec-Labs/mMAD/engines/pkg/logger"

// Manager manages the event system
type Manager struct {
	bus            *EventBus
	loggerHandler  *LoggerHandler
	metricsHandler *MetricsHandler
	alertHandler   *AlertHandler
}

// Config configures the event manager
type Config struct {
	BufferSize     int  `yaml:"buffer_size" json:"buffer_size"`
	EnableLogging  bool `yaml:"enable_logging" json:"enable_logging"`
	EnableMetrics  bool `yaml:"enable_metrics" json:"enable_metrics"`
	EnableAlerts   bool `yaml:"enable_alerts" json:"enable_alerts"`
	AlertThreshold int  `yaml:"alert_threshold" json:"alert_threshold"`
}

// DefaultConfig returns default event configuration
func DefaultConfig() *Config {
	return &Config{
		BufferSize:     1000,
		EnableLogging:  true,
		EnableMetrics:  true,
		EnableAlerts:   true,
		AlertThreshold: 10,
	}
}

// NewManager creates a new event manager
func NewManager(config *Config) *Manager {
	bus := NewEventBus(config.BufferSize)

	manager := &Manager{
		bus: bus,
	}

	// Setup handlers based on config
	if config.EnableLogging {
		manager.loggerHandler = NewLoggerHandler("info")
		bus.Subscribe(EventProofRequested, manager.loggerHandler)
		bus.Subscribe(EventProofGenerated, manager.loggerHandler)
		bus.Subscribe(EventProofVerified, manager.loggerHandler)
		bus.Subscribe(EventProofFailed, manager.loggerHandler)
		bus.Subscribe(EventCircuitCompiled, manager.loggerHandler)
		bus.Subscribe(EventError, manager.loggerHandler)
	}

	if config.EnableMetrics {
		manager.metricsHandler = NewMetricsHandler()
		// Subscribe to all event types for metrics
		eventTypes := []EventType{
			EventProofRequested, EventProofGenerated, EventProofVerified,
			EventProofFailed, EventCircuitCompiled, EventCircuitCached,
			EventSystemStarted, EventSystemStopped, EventError,
		}
		for _, eventType := range eventTypes {
			bus.Subscribe(eventType, manager.metricsHandler)
		}
	}

	if config.EnableAlerts {
		manager.alertHandler = NewAlertHandler(config.AlertThreshold)
		bus.Subscribe(EventProofFailed, manager.alertHandler)
		bus.Subscribe(EventError, manager.alertHandler)
		bus.Subscribe(EventProofGenerated, manager.alertHandler)
	}

	logger.Info("Event manager initialized",
		"buffer_size", config.BufferSize,
		"logging", config.EnableLogging,
		"metrics", config.EnableMetrics,
		"alerts", config.EnableAlerts)

	return manager
}

// GetBus returns the event bus
func (m *Manager) GetBus() *EventBus {
	return m.bus
}

// GetEmitter creates a new event emitter
func (m *Manager) GetEmitter(source string) *EventEmitter {
	return NewEventEmitter(m.bus, source)
}

// GetMetrics returns system metrics
func (m *Manager) GetMetrics() map[string]interface{} {
	metrics := make(map[string]interface{})

	// Bus stats
	metrics["bus"] = m.bus.GetStats()

	// Handler metrics
	if m.metricsHandler != nil {
		metrics["events"] = m.metricsHandler.GetMetrics()
	}

	return metrics
}

// AddHandler adds a custom event handler
func (m *Manager) AddHandler(eventType EventType, handler EventHandler) error {
	return m.bus.Subscribe(eventType, handler)
}

// RemoveHandler removes an event handler
func (m *Manager) RemoveHandler(eventType EventType, handler EventHandler) error {
	return m.bus.Unsubscribe(eventType, handler)
}

// Close shuts down the event manager
func (m *Manager) Close() error {
	logger.Info("Shutting down event manager")
	return m.bus.Close()
}

// Health checks the health of the event system
func (m *Manager) Health() map[string]interface{} {
	stats := m.bus.GetStats()

	health := map[string]interface{}{
		"status": "healthy",
		"stats":  stats,
	}

	// Check if buffer is getting full
	bufferSize := stats["buffer_size"].(int)
	queuedEvents := stats["queued_events"].(int)

	if queuedEvents > bufferSize*8/10 { // 80% full
		health["status"] = "warning"
		health["warning"] = "Event buffer is getting full"
	}

	if queuedEvents >= bufferSize {
		health["status"] = "critical"
		health["error"] = "Event buffer is full"
	}

	return health
}
