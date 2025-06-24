package circuits

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/GoSec-Labs/mMAD/engines/pkg/logger"
	"github.com/consensys/gnark/backend/groth16"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/r1cs"
)

// CircuitCompiler compiles and manages circuits
type CircuitCompiler struct {
	config  *CircuitConfig
	cache   *CircuitCache
	metrics *CompilerMetrics
	mu      sync.RWMutex
}

// CompiledCircuit represents a compiled circuit with its artifacts
type CompiledCircuit struct {
	ID               string
	Name             string
	Circuit          frontend.Circuit
	ConstraintSystem frontend.CompiledConstraintSystem
	ProvingKey       groth16.ProvingKey
	VerifyingKey     groth16.VerifyingKey
	CompileTime      time.Duration
	NumConstraints   int
	PublicInputs     []string
	PrivateInputs    []string
	Hash             string
	CreatedAt        time.Time
}

// CircuitCache caches compiled circuits
type CircuitCache struct {
	circuits map[string]*CompiledCircuit
	mu       sync.RWMutex
	maxSize  int
}

// CompilerMetrics tracks compilation metrics
type CompilerMetrics struct {
	CompilationsTotal  int64
	CompilationsFailed int64
	CacheHits          int64
	CacheMisses        int64
	AverageCompileTime time.Duration
	TotalConstraints   int64
}

// NewCircuitCompiler creates a new circuit compiler
func NewCircuitCompiler(config *CircuitConfig) *CircuitCompiler {
	return &CircuitCompiler{
		config: config,
		cache: &CircuitCache{
			circuits: make(map[string]*CompiledCircuit),
			maxSize:  100, // Cache up to 100 circuits
		},
		metrics: &CompilerMetrics{},
	}
}

// CompileCircuit compiles a circuit and caches the result
func (c *CircuitCompiler) CompileCircuit(ctx context.Context, circuitID string, circuit frontend.Circuit) (*CompiledCircuit, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	start := time.Now()
	logger.Info("Compiling circuit", "circuit_id", circuitID)

	// Check cache first
	if cached := c.cache.Get(circuitID); cached != nil {
		c.metrics.CacheHits++
		logger.Info("Circuit cache hit", "circuit_id", circuitID)
		return cached, nil
	}
	c.metrics.CacheMisses++

	// Compile constraint system
	cs, err := frontend.Compile(c.config.FieldModulus, r1cs.NewBuilder, circuit)
	if err != nil {
		c.metrics.CompilationsFailed++
		return nil, fmt.Errorf("failed to compile constraint system: %w", err)
	}

	// Setup proving and verifying keys
	pk, vk, err := groth16.Setup(cs)
	if err != nil {
		c.metrics.CompilationsFailed++
		return nil, fmt.Errorf("failed to setup keys: %w", err)
	}

	compileTime := time.Since(start)

	// Create compiled circuit
	compiled := &CompiledCircuit{
		ID:               circuitID,
		Circuit:          circuit,
		ConstraintSystem: cs,
		ProvingKey:       pk,
		VerifyingKey:     vk,
		CompileTime:      compileTime,
		NumConstraints:   cs.GetNbConstraints(),
		Hash:             c.computeCircuitHash(circuit),
		CreatedAt:        time.Now(),
	}

	// Extract input information
	compiled.PublicInputs, compiled.PrivateInputs = c.extractInputInfo(circuit)

	// Cache the result
	c.cache.Put(circuitID, compiled)

	// Update metrics
	c.metrics.CompilationsTotal++
	c.metrics.TotalConstraints += int64(compiled.NumConstraints)
	c.updateAverageCompileTime(compileTime)

	logger.Info("Circuit compiled successfully",
		"circuit_id", circuitID,
		"constraints", compiled.NumConstraints,
		"compile_time", compileTime)

	return compiled, nil
}

// GetCircuit retrieves a compiled circuit from cache
func (c *CircuitCompiler) GetCircuit(circuitID string) (*CompiledCircuit, error) {
	circuit := c.cache.Get(circuitID)
	if circuit == nil {
		return nil, fmt.Errorf("circuit not found: %s", circuitID)
	}
	return circuit, nil
}

// ListCircuits returns all cached circuits
func (c *CircuitCompiler) ListCircuits() []*CompiledCircuit {
	return c.cache.List()
}

// GetMetrics returns compiler metrics
func (c *CircuitCompiler) GetMetrics() *CompilerMetrics {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Return a copy to avoid race conditions
	return &CompilerMetrics{
		CompilationsTotal:  c.metrics.CompilationsTotal,
		CompilationsFailed: c.metrics.CompilationsFailed,
		CacheHits:          c.metrics.CacheHits,
		CacheMisses:        c.metrics.CacheMisses,
		AverageCompileTime: c.metrics.AverageCompileTime,
		TotalConstraints:   c.metrics.TotalConstraints,
	}
}

// Private methods

func (c *CircuitCompiler) computeCircuitHash(circuit frontend.Circuit) string {
	// In a real implementation, this would compute a hash of the circuit structure
	return fmt.Sprintf("hash_%p", circuit)
}

func (c *CircuitCompiler) extractInputInfo(circuit frontend.Circuit) ([]string, []string) {
	// In a real implementation, this would use reflection to extract input field names
	// For now, return placeholders
	return []string{"public_input_1", "public_input_2"}, []string{"private_input_1", "private_input_2"}
}

func (c *CircuitCompiler) updateAverageCompileTime(newTime time.Duration) {
	if c.metrics.CompilationsTotal == 1 {
		c.metrics.AverageCompileTime = newTime
	} else {
		// Running average
		total := float64(c.metrics.CompilationsTotal)
		current := float64(c.metrics.AverageCompileTime)
		new := float64(newTime)

		avg := (current*(total-1) + new) / total
		c.metrics.AverageCompileTime = time.Duration(avg)
	}
}

// CircuitCache methods

func (cc *CircuitCache) Get(circuitID string) *CompiledCircuit {
	cc.mu.RLock()
	defer cc.mu.RUnlock()

	return cc.circuits[circuitID]
}

func (cc *CircuitCache) Put(circuitID string, circuit *CompiledCircuit) {
	cc.mu.Lock()
	defer cc.mu.Unlock()

	// Check if cache is full
	if len(cc.circuits) >= cc.maxSize {
		// Simple LRU: remove oldest circuit
		var oldestID string
		var oldestTime time.Time

		for id, c := range cc.circuits {
			if oldestID == "" || c.CreatedAt.Before(oldestTime) {
				oldestID = id
				oldestTime = c.CreatedAt
			}
		}

		if oldestID != "" {
			delete(cc.circuits, oldestID)
		}
	}

	cc.circuits[circuitID] = circuit
}

func (cc *CircuitCache) List() []*CompiledCircuit {
	cc.mu.RLock()
	defer cc.mu.RUnlock()

	circuits := make([]*CompiledCircuit, 0, len(cc.circuits))
	for _, circuit := range cc.circuits {
		circuits = append(circuits, circuit)
	}

	return circuits
}

func (cc *CircuitCache) Clear() {
	cc.mu.Lock()
	defer cc.mu.Unlock()

	cc.circuits = make(map[string]*CompiledCircuit)
}

func (cc *CircuitCache) Size() int {
	cc.mu.RLock()
	defer cc.mu.RUnlock()

	return len(cc.circuits)
}
