package api

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/GoSec-Labs/mMAD/engines/internal/api/handlers"
	"github.com/GoSec-Labs/mMAD/engines/internal/circuits"
	"github.com/GoSec-Labs/mMAD/engines/internal/events"
	"github.com/GoSec-Labs/mMAD/engines/pkg/config"
	"github.com/GoSec-Labs/mMAD/engines/pkg/logger"
	"github.com/gin-gonic/gin"
)

// Server represents the API server
type Server struct {
	config  *config.Config
	handler *handlers.Handler
	server  *http.Server
	events  *events.Manager
}

// New creates a new API server
func New(
	cfg *config.Config,
	prover *config.ZKProofConfig,
	circuits *circuits.CircuitRegistry,
	events *events.Manager,
) *Server {
	// Set Gin mode
	if cfg.API.Debug {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	// Create handler
	handler := handlers.New(cfg, prover, circuits, events)

	// Create router
	router := gin.New()
	SetupRoutes(router, handler)

	// Create HTTP server
	server := &http.Server{
		Addr:           fmt.Sprintf(":%d", cfg.API.Port),
		Handler:        router,
		ReadTimeout:    time.Duration(cfg.API.ReadTimeout) * time.Second,
		WriteTimeout:   time.Duration(cfg.API.WriteTimeout) * time.Second,
		IdleTimeout:    time.Duration(cfg.API.IdleTimeout) * time.Second,
		MaxHeaderBytes: cfg.API.MaxHeaderBytes,
	}

	return &Server{
		config:  cfg,
		handler: handler,
		server:  server,
		events:  events,
	}
}

// Start starts the API server
func (s *Server) Start(ctx context.Context) error {
	logger.Info("Starting API server",
		"port", s.config.API.Port,
		"debug", s.config.API.Debug)

	// Emit system started event
	emitter := s.events.GetEmitter("api-server")
	emitter.EmitSystemStarted("1.0.0")

	// Start server
	go func() {
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Failed to start server", "error", err)
		}
	}()

	logger.Info("API server started", "address", s.server.Addr)
	return nil
}

// Stop gracefully stops the API server
func (s *Server) Stop(ctx context.Context) error {
	logger.Info("Stopping API server")

	// Emit system stopped event
	emitter := s.events.GetEmitter("api-server")
	emitter.EmitSystemStopped("graceful shutdown")

	// Graceful shutdown
	return s.server.Shutdown(ctx)
}

// GetAddr returns the server address
func (s *Server) GetAddr() string {
	return s.server.Addr
}
