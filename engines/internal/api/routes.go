package api

import (
	"github.com/gin-gonic/gin"
)

// SetupRoutes configures all API routes
func SetupRoutes(r *gin.Engine, h *handlers.Handler) {
	// Middleware
	r.Use(middleware.CORS())
	r.Use(middleware.RequestID())
	r.Use(middleware.Logger())
	r.Use(middleware.Recovery())

	// Health check
	r.GET("/health", h.Health)
	r.GET("/ready", h.Ready)

	// API v1 group
	v1 := r.Group("/api/v1")
	v1.Use(middleware.RateLimit())
	{
		// Proof endpoints
		proofs := v1.Group("/proofs")
		{
			proofs.POST("/generate", h.GenerateProof)
			proofs.POST("/verify", h.VerifyProof)
			proofs.GET("/:id", h.GetProof)
			proofs.GET("/:id/status", h.GetProofStatus)
			proofs.DELETE("/:id", h.DeleteProof)
		}

		// Circuit endpoints
		circuits := v1.Group("/circuits")
		{
			circuits.GET("/", h.ListCircuits)
			circuits.GET("/:id", h.GetCircuit)
			circuits.POST("/:id/compile", h.CompileCircuit)
			circuits.GET("/:id/info", h.GetCircuitInfo)
		}

		// Batch operations
		batch := v1.Group("/batch")
		{
			batch.POST("/proofs", h.BatchGenerateProofs)
			batch.POST("/verify", h.BatchVerifyProofs)
			batch.GET("/:id", h.GetBatchStatus)
		}

		// System endpoints
		system := v1.Group("/system")
		{
			system.GET("/metrics", h.GetMetrics)
			system.GET("/events", h.GetEvents)
			system.GET("/config", h.GetConfig)
			system.POST("/config/reload", h.ReloadConfig)
		}
	}
}
