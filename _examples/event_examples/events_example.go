package main

import (
	"fmt"
	"time"

	"github.com/GoSec-Labs/mMAD/engines/internal/events"
	"github.com/GoSec-Labs/mMAD/engines/internal/zkproof/types"
	"github.com/GoSec-Labs/mMAD/engines/pkg/logger"
)

func main() {
	// Initialize logger
	logger.Init(logger.Config{Level: "info"})

	fmt.Println("🎯 Event System Example")
	fmt.Println("======================")

	// Create event manager
	config := events.DefaultConfig()
	manager := events.NewManager(config)
	defer manager.Close()

	// Get emitter for our service
	emitter := manager.GetEmitter("proof-service")

	// Example 1: Basic event flow
	fmt.Println("\n📤 Emitting proof lifecycle events...")
	simulateProofLifecycle(emitter)

	// Example 2: Circuit compilation events
	fmt.Println("\n🔧 Emitting circuit events...")
	simulateCircuitCompilation(emitter)

	// Example 3: Error events
	fmt.Println("\n❌ Emitting error events...")
	simulateErrors(emitter)

	// Example 4: Custom handler
	fmt.Println("\n🔌 Adding custom handler...")
	addCustomHandler(manager)

	// Example 5: System events
	fmt.Println("\n🚀 Emitting system events...")
	simulateSystemEvents(emitter)

	// Wait for async processing
	time.Sleep(100 * time.Millisecond)

	// Show metrics
	fmt.Println("\n📊 Event Metrics:")
	showMetrics(manager)

	// Show health
	fmt.Println("\n🏥 System Health:")
	showHealth(manager)
}

func simulateProofLifecycle(emitter *events.EventEmitter) {
	proofID := "proof_12345"
	userID := "user_789"
	requestID := "req_456"

	// 1. Proof requested
	emitter.EmitProofRequested(proofID, types.ProofTypeBalance, userID, requestID)
	fmt.Printf("   ✅ Proof requested: %s\n", proofID)

	// 2. Proof generated (simulate processing time)
	time.Sleep(10 * time.Millisecond)
	emitter.EmitProofGenerated(proofID, types.ProofTypeBalance,
		5*time.Second, 1024, userID, requestID)
	fmt.Printf("   ✅ Proof generated: %s (5s, 1024 bytes)\n", proofID)

	// 3. Proof verified
	emitter.EmitProofVerified(proofID, types.ProofTypeBalance,
		50*time.Millisecond, userID, requestID)
	fmt.Printf("   ✅ Proof verified: %s (50ms)\n", proofID)
}

func simulateCircuitCompilation(emitter *events.EventEmitter) {
	emitter.EmitCircuitCompiled("balance_v1", "Balance Circuit",
		1000, 245*time.Millisecond)
	fmt.Printf("   ✅ Circuit compiled: balance_v1 (1000 constraints, 245ms)\n")

	emitter.EmitCircuitCompiled("solvency_v1", "Solvency Circuit",
		15000, 2*time.Second)
	fmt.Printf("   ✅ Circuit compiled: solvency_v1 (15000 constraints, 2s)\n")
}

func simulateErrors(emitter *events.EventEmitter) {
	// Simulate a few errors
	emitter.EmitProofFailed("proof_err_1", types.ProofTypeBalance,
		fmt.Errorf("invalid witness"), "user_123", "req_789")
	fmt.Printf("   ❌ Proof failed: proof_err_1\n")

	emitter.EmitError("circuit-compiler",
		fmt.Errorf("compilation timeout"), "user_123", "req_789")
	fmt.Printf("   ❌ System error: circuit-compiler\n")
}

func addCustomHandler(manager *events.Manager) {
	// Create a custom handler that counts successful proofs
	counter := &ProofCounterHandler{count: 0}

	manager.AddHandler(events.EventProofGenerated, counter)
	fmt.Printf("   🔌 Custom handler added\n")

	// Emit a few events to test the counter
	emitter := manager.GetEmitter("test-service")
	for i := 0; i < 3; i++ {
		emitter.EmitProofGenerated(fmt.Sprintf("proof_%d", i),
			types.ProofTypeBalance, time.Second, 512, "user", "req")
	}

	fmt.Printf("   📈 Custom handler counted: %d proofs\n", counter.GetCount())
}

func simulateSystemEvents(emitter *events.EventEmitter) {
	emitter.EmitSystemStarted("v1.0.0")
	fmt.Printf("   🚀 System started: v1.0.0\n")

	// Later...
	emitter.EmitSystemStopped("graceful shutdown")
	fmt.Printf("   🛑 System stopped: graceful shutdown\n")
}

func showMetrics(manager *events.Manager) {
	metrics := manager.GetMetrics()

	if busStats, ok := metrics["bus"].(map[string]interface{}); ok {
		fmt.Printf("   📊 Total handlers: %v\n", busStats["total_handlers"])
		fmt.Printf("   📊 Event types: %v\n", busStats["event_types"])
		fmt.Printf("   📊 Queued events: %v\n", busStats["queued_events"])
	}

	if eventMetrics, ok := metrics["events"].(map[string]interface{}); ok {
		if counts, ok := eventMetrics["event_counts"].(map[events.EventType]int64); ok {
			fmt.Printf("   📈 Event counts:\n")
			for eventType, count := range counts {
				fmt.Printf("      %s: %d\n", eventType, count)
			}
		}
	}
}

func showHealth(manager *events.Manager) {
	health := manager.Health()

	fmt.Printf("   🏥 Status: %s\n", health["status"])

	if warning, ok := health["warning"]; ok {
		fmt.Printf("   ⚠️  Warning: %s\n", warning)
	}

	if err, ok := health["error"]; ok {
		fmt.Printf("   ❌ Error: %s\n", err)
	}
}

// Custom handler example
type ProofCounterHandler struct {
	count int
}

func (h *ProofCounterHandler) Handle(event events.Event) error {
	h.count++
	return nil
}

func (h *ProofCounterHandler) CanHandle(eventType events.EventType) bool {
	return eventType == events.EventProofGenerated
}

func (h *ProofCounterHandler) GetCount() int {
	return h.count
}

// JSON output handler example
func createJSONOutputHandler() *events.JSONHandler {
	return events.NewJSONHandler(func(jsonString string) error {
		fmt.Printf("📄 JSON Event: %s\n", jsonString)
		return nil
	})
}
