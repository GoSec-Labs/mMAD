pragma circom 2.0.0;

// Re-export from circomlib for convenience
include "circomlib/circuits/comparators.circom";

template LessThan(n) {
    assert(n <= 252);
    signal input in[2];
    signal output out;

    component lt = LessEqThan(n);
    lt.in[0] <== in[0];
    lt.in[1] <== in[1] + 1;
    out <== lt.out;
}

template LessEqThan(n) {
    assert(n <= 252);
    signal input in[2];
    signal output out;

    component lt = ComparatorLT(n);
    lt.in[0] <== in[0];
    lt.in[1] <== in[1] + 1;
    out <== lt.out;
}

template ComparatorLT(n) {
    assert(n <= 252);
    signal input in[2];
    signal output out;

    component num2bits = Num2Bits(n+1);
    num2bits.in <== in[0] + (1<<n) - in[1];
    out <== 1 - num2bits.out[n];
}

template GreaterThan(n) {
    signal input in[2];
    signal output out;

    component lt = LessThan(n);
    lt.in[0] <== in[1];
    lt.in[1] <== in[0];
    out <== lt.out;
}

template GreaterEqThan(n) {
    signal input in[2];
    signal output out;

    component lt = LessEqThan(n);
    lt.in[0] <== in[1];
    lt.in[1] <== in[0];
    out <== lt.out;
}

template IsEqual() {
    signal input in[2];
    signal output out;

    component eq = IsZero();
    eq.in <== in[0] - in[1];
    out <== eq.out;
}

template IsZero() {
    signal input in;
    signal output out;

    signal inv;
    inv <-- in != 0 ? 1 / in : 0;
    out <== -in * inv + 1;
    in * out === 0;
}

// You can add custom comparators here if needed
template CustomGreaterThan(n) {
    signal input in[2];
    signal output out;
    
    component gt = GreaterThan(n);
    gt.in[0] <== in[0];
    gt.in[1] <== in[1];
    out <== gt.out;
}