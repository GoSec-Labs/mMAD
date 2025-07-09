const snarkjs = require("snarkjs");
const fs = require("fs");
const path = require("path");

async function testReserveProof() {
    console.log("🔐 Testing Reserve Proof Generation...");
    
    // Use paths relative to project root (go up 2 levels from scripts/circom/)
    const wasmPath = "../../circuits/generated/ReserveProof_js/ReserveProof.wasm";
    const zkeyPath = "../../circuits/generated/keys/ReserveProof.zkey";
    
    // Check if files exist
    if (!fs.existsSync(wasmPath)) {
        console.log("❌ WASM file not found at:", wasmPath);
        return false;
    }
    
    if (!fs.existsSync(zkeyPath)) {
        console.log("❌ Zkey file not found at:", zkeyPath);
        return false;
    }
    
    const input = {
        minRequiredReserve: "110000",
        currentSupply: "100000", 
        timestamp: Math.floor(Date.now() / 1000).toString()
    };
    
    try {
        const { proof, publicSignals } = await snarkjs.groth16.fullProve(
            input,
            wasmPath,
            zkeyPath
        );
        
        console.log("✅ Reserve proof generated successfully!");
        console.log("📊 Public signals:", publicSignals);
        
        // Save proof for testing
        fs.writeFileSync("test-proof.json", JSON.stringify({
            proof,
            publicSignals
        }, null, 2));
        
        return true;
    } catch (error) {
        console.error("❌ Proof generation failed:", error.message);
        return false;
    }
}

async function testComplianceProof() {
    console.log("🔐 Testing Compliance Proof Generation...");
    
    const wasmPath = "../../circuits/generated/ComplianceCheck_js/ComplianceCheck.wasm";
    const zkeyPath = "../../circuits/generated/keys/ComplianceCheck.zkey";
    
    // Check if files exist
    if (!fs.existsSync(wasmPath)) {
        console.log("❌ WASM file not found at:", wasmPath);
        return false;
    }
    if (!fs.existsSync(zkeyPath)) {
        console.log("❌ Zkey file not found at:", zkeyPath);
        return false;
    }
    
    const input = {
        userHash: "12345678901234567890123456789012", // 32 byte hash
        riskScore: "5",
        timestamp: Math.floor(Date.now() / 1000).toString()
    };
    
    try {
        const { proof, publicSignals } = await snarkjs.groth16.fullProve(
            input,
            wasmPath,
            zkeyPath
        );
        
        console.log("✅ Compliance proof generated successfully!");
        console.log("📊 Public signals:", publicSignals);
        return true;
    } catch (error) {
        console.error("❌ Compliance proof generation failed:", error.message);
        return false;
    }
}

async function runAllTests() {
    console.log("🧪 Starting ZK Proof Tests...\n");
    
    const reserveResult = await testReserveProof();
    console.log("");
    const complianceResult = await testComplianceProof();
    
    console.log("\n📋 Test Results:");
    console.log(`Reserve Proof: ${reserveResult ? '✅ PASS' : '❌ FAIL'}`);
    console.log(`Compliance Proof: ${complianceResult ? '✅ PASS' : '❌ FAIL'}`);
    
    if (reserveResult && complianceResult) {
        console.log("\n🎉 All ZK proofs working correctly!");
        console.log("🚀 Ready for smart contract integration!");
    }
}

runAllTests().catch(console.error);
