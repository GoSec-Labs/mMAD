pragma circom 2.0.0;

template ReserveProof() {
    signal input minRequiredReserve;
    signal input actualReserves;
    signal output isValid;
    
    component geq = GreaterEqThan(64);
    geq.in[0] <== actualReserves;
    geq.in[1] <== minRequiredReserve;
    isValid <== geq.out;
}

template GreaterEqThan(n) {
    assert(n <= 252);
    signal input in[2];
    signal output out;
    
    component lt = LessThan(n);
    lt.in[0] <== in[1];
    lt.in[1] <== in[0] + 1;
    out <== 1 - lt.out;
}

template LessThan(n) {
    assert(n <= 252);
    signal input in[2];
    signal output out;

    component num2bits = Num2Bits(n+1);
    num2bits.in <== in[0] + (1<<n) - in[1];
    out <== 1 - num2bits.out[n];
}

template Num2Bits(n) {
    signal input in;
    signal output out[n];
    var lc1=0;

    for (var i = 0; i<n; i++) {
        out[i] <-- (in >> i) & 1;
        out[i] * (out[i] -1 ) === 0;
        lc1 += out[i] * 2**i;
    }
    lc1 === in;
}

component main = ReserveProof();
