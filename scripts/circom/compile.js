const { exec } = require('child_process');
const path = require('path');
const fs = require('fs');

async function compileCircuit(circuitName) {
    console.log(`🔨 Compiling ${circuitName}...`);

    const circuitPath = path.join(__dirname, `../../circuits/${circuitName}.circom`);
    const buildPath = path.join(__dirname, `../../circuits/generated`);

    // Check if circuit file exists
    if (!fs.existsSync(circuitPath)) {
        console.error(`❌ Circuit file not found: ${circuitPath}`);
        throw new Error(`Circuit file ${circuitName}.circom not found`);
    }

    // Ensure build directory exists
    if (!fs.existsSync(buildPath)) {
        console.log(`📁 Creating build directory: ${buildPath}`);
        fs.mkdirSync(buildPath, { recursive: true });
    }

    return new Promise((resolve, reject) => {
        // IMPORTANT: Add -l node_modules flag for circomlib includes
        const cmd = `circom "${circuitPath}" --r1cs --wasm --sym -o "${buildPath}" -l node_modules`;
        console.log(`🚀 Running: ${cmd}`);

        exec(cmd, (error, stdout, stderr) => {
            if (error) {
                console.error(`❌ Error compiling ${circuitName}:`, error.message);
                if (stderr) console.error('STDERR:', stderr);
                reject(error);
                return;
            }

            if (stderr) {
                console.log('STDERR:', stderr);
            }

            console.log(`✅ ${circuitName} compiled successfully!`);
            if (stdout) console.log(stdout);
            resolve();
        });
    });
}

async function main() {
    console.log("🎯 Starting circuit compilation...");

    try {
        // Check if circom is installed
        const { exec } = require('child_process');
        await new Promise((resolve, reject) => {
            exec('circom --version', (error, stdout, stderr) => {
                if (error) {
                    console.error('❌ Circom not found. Please install circom first.');
                    console.error('Install with: cargo install --git https://github.com/iden3/circom.git');
                    reject(error);
                    return;
                }
                console.log(`✅ Circom version: ${stdout.trim()}`);
                resolve();
            });
        });

        // List of circuits to compile
        const circuits = ["ReserveProof", "ComplianceCheck", "BatchVerifier"];

        for (const circuit of circuits) {
            await compileCircuit(circuit);
        }

        console.log("🎉 All circuits compiled successfully!");

        // List generated files
        const generatedPath = path.join(__dirname, '../../circuits/generated');
        const files = fs.readdirSync(generatedPath);
        console.log("\n📁 Generated files:");
        files.forEach(file => console.log(`   ${file}`));

    } catch (error) {
        console.error("💥 Compilation failed:", error.message);
        process.exit(1);
    }
}

if (require.main === module) {
    main();
}

module.exports = { compileCircuit };