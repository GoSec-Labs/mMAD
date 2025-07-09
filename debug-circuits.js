const snarkjs = require("snarkjs");
const fs = require("fs");

async function debugCircuitInputs() {
    console.log("🔍 Debugging Circuit Input Requirements...");

    // Try ReserveProof with minimal inputs
    console.log("\n--- Testing ReserveProof with different input structures ---");

    const testInputs = [
        // Test 1: Single values
        { minRequiredReserve: 110000, currentSupply: 100000, timestamp: 1234567890 },

        // Test 2: String values
        { minRequiredReserve: "110000", currentSupply: "100000", timestamp: "1234567890" },

        // Test 3: Just one field
        { minRequiredReserve: 110000 },

        // Test 4: Different field names (common variations)
        { minRequired: 110000, supply: 100000, time: 1234567890 },
        { reserve: 110000, totalSupply: 100000, blockTime: 1234567890 }
    ];

    for (let i = 0; i < testInputs.length; i++) {
        console.log(`\nTest ${i + 1}:`, testInputs[i]);
        try {
            const { proof, publicSignals } = await snarkjs.groth16.fullProve(
                testInputs[i],
                "circuits/generated/ReserveProof_js/ReserveProof.wasm",
                "circuits/generated/keys/ReserveProof.zkey"
            );
            console.log("✅ SUCCESS! This input structure works:");
            console.log("Input:", testInputs[i]);
            console.log("Public signals:", publicSignals);
            return testInputs[i];
        } catch (error) {
            console.log(`❌ Failed: ${error.message}`);
        }
    }

    console.log("\n❌ None of the test inputs worked. Need to check the circuit file.");
    return null;
}

debugCircuitInputs().catch(console.error);
