pragma circom 2.0.0;

include "circomlib/circuits/poseidon.circom";
include "circomlib/circuits/comparators.circom";

template ComplianceCheck() {
    // Public inputs
    signal input userHash;
    signal input riskScore;
    
    // Private inputs
    signal input userID;
    signal input kycData;
    
    // Outputs
    signal output isCompliant;
    signal output userCommitment;
    
    // Verify user hash matches actual user ID
    component userHashCheck = Poseidon(1);
    userHashCheck.inputs[0] <== userID;
    userHash === userHashCheck.out;
    
    // Check risk score is within acceptable range (0-30)
    component riskCheck = LessEqThan(8);
    riskCheck.in[0] <== riskScore;
    riskCheck.in[1] <== 30;
    isCompliant <== riskCheck.out;
    
    // Create commitment to user data
    component commitment = Poseidon(2);
    commitment.inputs[0] <== userID;
    commitment.inputs[1] <== kycData;
    userCommitment <== commitment.out;
}

component main = ComplianceCheck();