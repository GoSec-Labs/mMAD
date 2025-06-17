pragma circom 2.0.0;

include "circomlib/circuits/poseidon.circom";
include "circomlib/circuits/comparators.circom";

template BatchVerifier() {
    // Simple batch verifier for 3 reserve proofs
    signal input minRequiredReserves[3];
    signal input actualReserves[3];
    
    signal output allValid;
    signal output batchCommitment;
    
    // Verify each reserve proof
    component checks[3];
    signal validFlags[3];
    
    for (var i = 0; i < 3; i++) {
        checks[i] = GreaterEqThan(64);
        checks[i].in[0] <== actualReserves[i];
        checks[i].in[1] <== minRequiredReserves[i];
        validFlags[i] <== checks[i].out;
    }
    
    // Check all proofs are valid (sum should equal 3)
    signal sum;
    sum <== validFlags[0] + validFlags[1] + validFlags[2];
    
    component allValidCheck = IsEqual();
    allValidCheck.in[0] <== sum;
    allValidCheck.in[1] <== 3;
    allValid <== allValidCheck.out;
    
    // Create batch commitment
    component batchHash = Poseidon(3);
    batchHash.inputs[0] <== actualReserves[0];
    batchHash.inputs[1] <== actualReserves[1];
    batchHash.inputs[2] <== actualReserves[2];
    batchCommitment <== batchHash.out;
}

component main = BatchVerifier();