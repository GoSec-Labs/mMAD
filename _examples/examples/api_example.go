package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/GoSec-Labs/mMAD/engines/internal/api/models"
	"github.com/GoSec-Labs/mMAD/engines/internal/zkproof/types"
)

const baseURL = "http://localhost:8080/api/v1"

func main() {
	fmt.Println("🌐 API Client Examples")
	fmt.Println("======================")

	// Wait for server to start
	time.Sleep(1 * time.Second)

	// Example 1: Health check
	fmt.Println("\n🏥 Health Check:")
	checkHealth()

	// Example 2: List circuits
	fmt.Println("\n📋 List Circuits:")
	listCircuits()

	// Example 3: Generate proof
	fmt.Println("\n🔐 Generate Proof:")
	generateProof()

	// Example 4: Verify proof
	fmt.Println("\n✅ Verify Proof:")
	verifyProof()

	// Example 5: Batch operations
	fmt.Println("\n📦 Batch Operations:")
	batchGenerate()

	// Example 6: System metrics
	fmt.Println("\n📊 System Metrics:")
	getMetrics()
}

func checkHealth() {
	resp, err := http.Get(baseURL + "/../health")
	if err != nil {
		fmt.Printf("   ❌ Error: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("   Status: %d\n", resp.StatusCode)
	fmt.Printf("   Response: %s\n", string(body))
}

func listCircuits() {
	resp, err := http.Get(baseURL + "/circuits")
	if err != nil {
		fmt.Printf("   ❌ Error: %v\n", err)
		return
	}
	defer resp.Body.Close()

	var response models.APIResponse
	json.NewDecoder(resp.Body).Decode(&response)

	fmt.Printf("   Success: %v\n", response.Success)
	if data, ok := response.Data.(map[string]interface{}); ok {
		if circuits, ok := data["circuits"].([]interface{}); ok {
			fmt.Printf("   Circuits found: %d\n", len(circuits))
			for i, circuit := range circuits {
				if c, ok := circuit.(map[string]interface{}); ok {
					fmt.Printf("   %d. %s (%s)\n", i+1, c["name"], c["id"])
				}
			}
		}
	}
}

func generateProof() {
	request := models.ProofRequest{
		CircuitID: "balance_v1",
		ProofType: types.ProofTypeBalance,
		PublicInputs: map[string]interface{}{
			"threshold": 1000,
			"user_id":   12345,
			"nonce":     1,
			"timestamp": time.Now().Unix(),
		},
		PrivateInputs: map[string]interface{}{
			"balance": 2500,
			"salt":    98765,
		},
		UserID:   "user_123",
		Priority: 5,
	}

	jsonData, _ := json.Marshal(request)
	resp, err := http.Post(baseURL+"/proofs/generate", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("   ❌ Error: %v\n", err)
		return
	}
	defer resp.Body.Close()

	var response models.APIResponse
	json.NewDecoder(resp.Body).Decode(&response)

	fmt.Printf("   Success: %v\n", response.Success)
	if response.Success {
		if data, ok := response.Data.(map[string]interface{}); ok {
			fmt.Printf("   Proof ID: %s\n", data["proof_id"])
			fmt.Printf("   Status: %s\n", data["status"])
			fmt.Printf("   Duration: %s\n", data["duration"])
		}
	} else if response.Error != nil {
		fmt.Printf("   Error: %s\n", response.Error.Message)
	}
}

func verifyProof() {
	request := models.VerifyRequest{
		ProofData: "mock_proof_data_here",
		PublicInputs: map[string]interface{}{
			"threshold": 1000,
			"user_id":   12345,
		},
		CircuitID: "balance_v1",
	}

	jsonData, _ := json.Marshal(request)
	resp, err := http.Post(baseURL+"/proofs/verify", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("   ❌ Error: %v\n", err)
		return
	}
	defer resp.Body.Close()

	var response models.APIResponse
	json.NewDecoder(resp.Body).Decode(&response)

	fmt.Printf("   Success: %v\n", response.Success)
	if response.Success {
		if data, ok := response.Data.(map[string]interface{}); ok {
			fmt.Printf("   Valid: %v\n", data["valid"])
			fmt.Printf("   Duration: %s\n", data["duration"])
		}
	}
}

func batchGenerate() {
	requests := []models.ProofRequest{
		{
			CircuitID: "balance_v1",
			ProofType: types.ProofTypeBalance,
			PublicInputs: map[string]interface{}{
				"threshold": 1000,
				"user_id":   12345,
			},
			PrivateInputs: map[string]interface{}{
				"balance": 2500,
				"salt":    98765,
			},
		},
		{
			CircuitID: "balance_v1",
			ProofType: types.ProofTypeBalance,
			PublicInputs: map[string]interface{}{
				"threshold": 500,
				"user_id":   67890,
			},
			PrivateInputs: map[string]interface{}{
				"balance": 1200,
				"salt":    54321,
			},
		},
	}

	batchRequest := models.BatchProofRequest{
		Requests: requests,
		Parallel: true,
	}

	jsonData, _ := json.Marshal(batchRequest)
	resp, err := http.Post(baseURL+"/batch/proofs", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("   ❌ Error: %v\n", err)
		return
	}
	defer resp.Body.Close()

	var response models.APIResponse
	json.NewDecoder(resp.Body).Decode(&response)

	fmt.Printf("   Success: %v\n", response.Success)
	if response.Success {
		if data, ok := response.Data.(map[string]interface{}); ok {
			fmt.Printf("   Batch ID: %s\n", data["batch_id"])
			fmt.Printf("   Total: %.0f\n", data["total"])
			fmt.Printf("   Completed: %.0f\n", data["completed"])
			fmt.Printf("   Failed: %.0f\n", data["failed"])
		}
	}
}

func getMetrics() {
	resp, err := http.Get(baseURL + "/system/metrics")
	if err != nil {
		fmt.Printf("   ❌ Error: %v\n", err)
		return
	}
	defer resp.Body.Close()

	var response models.APIResponse
	json.NewDecoder(resp.Body).Decode(&response)

	fmt.Printf("   Success: %v\n", response.Success)
	if response.Success {
		if data, ok := response.Data.(map[string]interface{}); ok {
			if proofStats, ok := data["proof_stats"].(map[string]interface{}); ok {
				fmt.Printf("   Total Proofs Generated: %.0f\n", proofStats["total_generated"])
				fmt.Printf("   Total Proofs Verified: %.0f\n", proofStats["total_verified"])
				fmt.Printf("   Average Generate Time: %s\n", proofStats["avg_generate_time"])
			}
			if systemStats, ok := data["system_stats"].(map[string]interface{}); ok {
				fmt.Printf("   Uptime: %s\n", systemStats["uptime"])
				fmt.Printf("   Memory Usage: %.0f bytes\n", systemStats["memory_usage"])
			}
		}
	}
}
