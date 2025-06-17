pragma circom 2.0.0;

include "circomlib/circuits/poseidon.circom";

template HashCommitment() {
    signal input value;
    signal input nonce;
    signal output commitment;

    component poseidon = Poseidon(2);
    poseidon.inputs[0] <== value;
    poseidon.inputs[1] <== nonce;
    commitment <== poseidon.out;
}

template MultiHash(n) {
    signal input in[n];
    signal output out;

    component poseidon = Poseidon(n);
    for (var i = 0; i < n; i++) {
        poseidon.inputs[i] <== in[i];
    }
    out <== poseidon.out;
}