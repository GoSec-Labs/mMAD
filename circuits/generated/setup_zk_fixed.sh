#!/bin/bash

echo "🚀 Setting up mMAD ZK circuits..."

# Navigate to circuits directory
cd circuits/generated/

# Create keys directory
mkdir -p keys

# Check if we have a working ceremony file, if not create one
if [ ! -f "../ceremony/powersOfTau28_hez_final_12.ptau" ]; then
    echo "📦 Generating local ceremony file for development..."
    snarkjs powersoftau new bn128 12 pot12_0000.ptau -v
    snarkjs powersoftau contribute pot12_0000.ptau pot12_0001.ptau --name="First contribution" -v -e="random entropy"
    snarkjs powersoftau prepare phase2 pot12_0001.ptau pot12_final.ptau -v
    CEREMONY_FILE="pot12_final.ptau"
else
    CEREMONY_FILE="../ceremony/powersOfTau28_hez_final_12.ptau"
fi

echo "🔐 Setting up ReserveProof circuit..."
snarkjs groth16 setup ReserveProof.r1cs $CEREMONY_FILE keys/ReserveProof_0000.zkey
snarkjs zkey contribute keys/ReserveProof_0000.zkey keys/ReserveProof.zkey --name="mMAD contribution" --entropy="mMAD random entropy 1"

echo "🔐 Setting up ComplianceCheck circuit..."
snarkjs groth16 setup ComplianceCheck.r1cs $CEREMONY_FILE keys/ComplianceCheck_0000.zkey
snarkjs zkey contribute keys/ComplianceCheck_0000.zkey keys/ComplianceCheck.zkey --name="mMAD contribution" --entropy="mMAD random entropy 2"

echo "🔐 Setting up BatchVerifier circuit..."
snarkjs groth16 setup BatchVerifier.r1cs $CEREMONY_FILE keys/BatchVerifier_0000.zkey
snarkjs zkey contribute keys/BatchVerifier_0000.zkey keys/BatchVerifier.zkey --name="mMAD contribution" --entropy="mMAD random entropy 3"

# Create contracts/generated directory
mkdir -p ../../contracts/generated

echo "📝 Generating Solidity verifiers..."
snarkjs zkey export solidityverifier keys/ReserveProof.zkey ../../contracts/generated/ReserveProofVerifier.sol
snarkjs zkey export solidityverifier keys/ComplianceCheck.zkey ../../contracts/generated/ComplianceCheckVerifier.sol
snarkjs zkey export solidityverifier keys/BatchVerifier.zkey ../../contracts/generated/BatchVerifierVerifier.sol

echo "🔑 Generating verification keys..."
snarkjs zkey export verificationkey keys/ReserveProof.zkey keys/ReserveProof_vkey.json
snarkjs zkey export verificationkey keys/ComplianceCheck.zkey keys/ComplianceCheck_vkey.json
snarkjs zkey export verificationkey keys/BatchVerifier.zkey keys/BatchVerifier_vkey.json

echo "✅ ZK setup complete!"
echo ""
echo "📁 Generated files:"
echo "   - Proving keys: circuits/generated/keys/*.zkey"
echo "   - Verification keys: circuits/generated/keys/*_vkey.json"
echo "   - Solidity verifiers: contracts/generated/*Verifier.sol"
echo ""
echo "🎯 Next steps:"
echo "   1. Create ZKReserveVerifier.sol wrapper contract"
echo "   2. Update MMadToken.sol to use the generated verifiers"
echo "   3. Test the complete integration"
