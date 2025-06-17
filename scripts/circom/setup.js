const snarkjs = require("snarkjs");
const fs = require("fs");
const path = require("path");

async function setup(circuitName) {
    console.log(`🔧 Setting up ${circuitName}...`);

    const buildPath = path.join(__dirname, `../../circuits/generated`);
    const ceremonyPath = path.join(__dirname, `../../circuits/ceremony`);

    const r1csPath = path.join(buildPath, `${circuitName}.r1cs`);
    const ptauPath = path.join(ceremonyPath, "powersOfTau28_hez_final_15.ptau");

    // Check if required files exist
    if (!fs.existsSync(r1csPath)) {
        console.error(`❌ R1CS file not found: ${r1csPath}`);
        console.error("   Run 'npm run circuit:compile' first");
        throw new Error(`R1CS file for ${circuitName} not found`);
    }

    if (!fs.existsSync(ptauPath)) {
        console.error(`❌ Powers of Tau file not found: ${ptauPath}`);
        console.error("   Please download it with:");
        console.error("   cd circuits/ceremony");
        console.error("   curl -o powersOfTau28_hez_final_15.ptau https://hermez.s3-eu-west-1.amazonaws.com/powersOfTau28_hez_final_15.ptau");
        throw new Error("Powers of Tau file not found");
    }

    try {
        // Generate initial zkey
        console.log("🔑 Generating initial zkey...");
        const zkeyPath = path.join(buildPath, `${circuitName}_0000.zkey`);
        await snarkjs.zKey.newZKey(r1csPath, ptauPath, zkeyPath);
        console.log(`✅ Initial zkey generated: ${zkeyPath}`);

        // Contribute to ceremony (dummy contribution for development)
        console.log("🎭 Contributing to ceremony...");
        const finalZkeyPath = path.join(buildPath, `${circuitName}_final.zkey`);
        await snarkjs.zKey.contribute(
            zkeyPath,
            finalZkeyPath,
            "mMAD Development",
            "random entropy for development " + Math.random().toString()
        );
        console.log(`✅ Final zkey generated: ${finalZkeyPath}`);

        // Export verification key
        console.log("📋 Exporting verification key...");
        const vkeyPath = path.join(buildPath, `${circuitName}_verification_key.json`);
        const vKey = await snarkjs.zKey.exportVerificationKey(finalZkeyPath);
        fs.writeFileSync(vkeyPath, JSON.stringify(vKey, null, 2));
        console.log(`✅ Verification key exported: ${vkeyPath}`);

        // Generate Solidity verifier
        console.log("📜 Generating Solidity verifier...");
        const solidityVerifier = await snarkjs.zKey.exportSolidityVerifier(finalZkeyPath);
        const verifierPath = path.join(buildPath, `${circuitName}Verifier.sol`);
        fs.writeFileSync(verifierPath, solidityVerifier);
        console.log(`✅ Solidity verifier generated: ${verifierPath}`);

        // Clean up intermediate files
        if (fs.existsSync(zkeyPath)) {
            fs.unlinkSync(zkeyPath);
            console.log("🧹 Cleaned up intermediate zkey file");
        }

        console.log(`🎉 ${circuitName} setup complete!`);

    } catch (error) {
        console.error(`❌ Setup failed for ${circuitName}:`, error.message);
        throw error;
    }
}

async function main() {
    console.log("🎯 Starting circuit setup...");

    try {
        // Check if snarkjs is available
        console.log(`✅ Using snarkjs version: ${require('snarkjs/package.json').version}`);

        const circuits = ["ReserveProof", "ComplianceCheck", "BatchVerifier"];

        for (const circuit of circuits) {
            await setup(circuit);
            console.log(""); // Empty line for readability
        }

        console.log("🎉 All circuits setup complete!");

        // List generated files
        const generatedPath = path.join(__dirname, '../../circuits/generated');
        console.log("\n📁 Generated files:");
        const files = fs.readdirSync(generatedPath);
        files.forEach(file => {
            const stats = fs.statSync(path.join(generatedPath, file));
            const size = (stats.size / 1024).toFixed(1);
            console.log(`   ${file} (${size} KB)`);
        });

    } catch (error) {
        console.error("💥 Setup failed:", error.message);
        process.exit(1);
    }
}

if (require.main === module) {
    main();
}

module.exports = { setup };