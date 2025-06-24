package handlers

import (
	"net/http"
	"time"

	"github.com/GoSec-Labs/mMAD/engines/internal/api/models"
	"github.com/gin-gonic/gin"
)

// Health returns service health status
func (h *Handler) Health(c *gin.Context) {
	health := HealthStatus{
		Status:    "healthy",
		Timestamp: time.Now(),
		Version:   "1.0.0",
		Uptime:    time.Since(h.startTime),
		Services: map[string]string{
			"proof_engine": "healthy",
			"circuits":     "healthy",
			"events":       "healthy",
			"config":       "healthy",
		},
	}

	// Check individual service health
	if !h.checkProofEngineHealth() {
		health.Services["proof_engine"] = "unhealthy"
		health.Status = "degraded"
	}

	if !h.checkCircuitsHealth() {
		health.Services["circuits"] = "unhealthy"
		health.Status = "degraded"
	}

	if !h.checkEventsHealth() {
		health.Services["events"] = "unhealthy"
		health.Status = "degraded"
	}

	statusCode := http.StatusOK
	if health.Status != "healthy" {
		statusCode = http.StatusServiceUnavailable
	}

	c.JSON(statusCode, models.APIResponse{
		Success:   true,
		Data:      health,
		RequestID: getRequestID(c),
		Timestamp: time.Now(),
	})
}

// Ready returns readiness status
func (h *Handler) Ready(c *gin.Context) {
	ready := true
	services := make(map[string]bool)

	// Check if all services are ready
	services["proof_engine"] = h.checkProofEngineHealth()
	services["circuits"] = h.checkCircuitsHealth()
	services["events"] = h.checkEventsHealth()

	for _, serviceReady := range services {
		if !serviceReady {
			ready = false
			break
		}
	}

	statusCode := http.StatusOK
	if !ready {
		statusCode = http.StatusServiceUnavailable
	}

	c.JSON(statusCode, models.APIResponse{
		Success: ready,
		Data: map[string]interface{}{
			"ready":    ready,
			"services": services,
		},
		RequestID: getRequestID(c),
		Timestamp: time.Now(),
	})
}

// Helper functions
func (h *Handler) checkProofEngineHealth() bool {
	// Check if proof engine is responsive
	return h.prover != nil
}

func (h *Handler) checkCircuitsHealth() bool {
	// Check if circuits are available
	return h.circuits != nil
}

func (h *Handler) checkEventsHealth() bool {
	// Check event system health
	if h.events == nil {
		return false
	}

	healthData := h.events.Health()
	status, ok := healthData["status"].(string)
	return ok && status != "critical"
}

func getRequestID(c *gin.Context) string {
	if id, exists := c.Get("request_id"); exists {
		return id.(string)
	}
	return ""
}
