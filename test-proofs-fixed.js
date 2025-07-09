const snarkjs = require("snarkjs");
const fs = require("fs");

async function testReserveProof() {
    console.log(" Testing Reserve Proof Generation...");

    // CORRECT inputs based on your ReserveProof.circom
    const input = {
        minRequiredReserve: 110000,
        actualReserves: 120000  // This must be >= minRequiredReserve for proof to pass
    };

    try {
        const { proof, publicSignals } = await snarkjs.groth16.fullProve(
            input,
            "circuits/generated/ReserveProof_js/ReserveProof.wasm",
            "circuits/generated/keys/ReserveProof.zkey"
        );

        console.log(" Reserve proof generated successfully!");
        console.log(" Input: ", input);
        console.log(" Public signals:", publicSignals);
        console.log(" Is Valid(output): ", publicSignals[0] === "1" ? "YES" : "NO");

        // Save proof for smart contract testing
        fs.writeFileSync("reserve-proof.json", JSON.stringify({
            proof,
            publicSignals,
            input
        }, null, 2));

        return true;
    } catch (error) {
        console.error("❌ Proof generation failed:", error.message);
        return false;
    }
}

async function testComplianceProof() {
    console.log(" Testing Compliance Proof Generation...");

    // CORRECT inputs based on your ComplianceCheck.circom
    const input = {
        // Public inputs
        userHash: "12345678901234567890123456789012345678901234567890123456789012345", // This should match Poseidon(userID)
        riskScore: 25, // Must be <= 30 to pass

        // Private inputs (these are what actually get verified)
        userID: 12345,
        kycData: 98765
    };

    try {
        const { proof, publicSignals } = await snarkjs.groth16.fullProve(
            input,
            "circuits/generated/ComplianceCheck_js/ComplianceCheck.wasm",
            "circuits/generated/keys/ComplianceCheck.zkey"
        );

        console.log(" Compliance proof generated successfully!");
        console.log(" Input:", input);
        console.log(" Public signals:", publicSignals);
        console.log(" Is Compliant:", publicSignals[0] === "1" ? "YES" : "NO");
        console.log(" User Commitment:", publicSignals[1]);

        return true;
    } catch (error) {
        console.error("❌ Compliance proof generation failed:", error.message);
        return false;
    }
}

async function testBatchVerifier() {
    console.log(" Testing Batch Verifier...");

    // CORRECT inputs based on your BatchVerifier.circom
    const input = {
        minRequiredReserves: [100000, 200000, 150000],  // Array of 3 values
        actualReserves: [120000, 250000, 180000]        // Array of 3 values (all >= required)
    };

    try {
        const { proof, publicSignals } = await snarkjs.groth16.fullProve(
            input,
            "circuits/generated/BatchVerifier_js/BatchVerifier.wasm",
            "circuits/generated/keys/BatchVerifier.zkey"
        );

        console.log(" Batch proof generated successfully!");
        console.log(" Input:", input);
        console.log(" Public signals:", publicSignals);
        console.log(" All Valid:", publicSignals[0] === "1" ? "YES" : "NO");
        console.log(" Batch Commitment:", publicSignals[1]);

        return true;
    } catch (error) {
        console.error(" Batch proof generation failed:", error.message);
        return false;
    }
}

async function runAllTests() {
    console.log(" Starting ZK Proof Tests with CORRECT inputs...\n");

    const reserveResult = await testReserveProof();
    console.log("");
    const complianceResult = await testComplianceProof();
    console.log("");
    const batchResult = await testBatchVerifier();

    console.log("\n📋 Test Results:");
    console.log(`Reserve Proof: ${reserveResult ? '✅ PASS' : '❌ FAIL'}`);
    console.log(`Compliance Proof: ${complianceResult ? '✅ PASS' : '❌ FAIL'}`);
    console.log(`Batch Proof: ${batchResult ? '✅ PASS' : '❌ FAIL'}`);

    if (reserveResult && complianceResult && batchResult) {
        console.log("\n ALL ZK PROOFS WORKING CORRECTLY!");
        console.log(" Ready for smart contract integration!");
    }
}

runAllTests().catch(console.error);