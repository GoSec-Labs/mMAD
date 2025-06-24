package circuits

import (
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/std/algebra/emulated/sw_bn254"
	"github.com/consensys/gnark/std/signature/eddsa"
)

// SignatureCircuit verifies EdDSA signatures in zero-knowledge
type SignatureCircuit struct {
	// Public inputs
	PublicKey eddsa.PublicKey   `gnark:",public"`
	Message   frontend.Variable `gnark:",public"`

	// Private inputs
	Signature eddsa.Signature `gnark:",secret"`
}

func (circuit *SignatureCircuit) Define(api frontend.API) error {
	// Verify EdDSA signature
	curve, err := sw_bn254.NewCurve(api)
	if err != nil {
		return err
	}

	// Verify the signature
	circuit.Signature.Verify(curve, circuit.Message, circuit.PublicKey)

	return nil
}

// MultiSignatureCircuit verifies multiple signatures
type MultiSignatureCircuit struct {
	// Public inputs
	NumSignatures frontend.Variable   `gnark:",public"`
	PublicKeys    []eddsa.PublicKey   `gnark:",public"`
	Messages      []frontend.Variable `gnark:",public"`

	// Private inputs
	Signatures []eddsa.Signature `gnark:",secret"`
}

func (circuit *MultiSignatureCircuit) Define(api frontend.API) error {
	curve, err := sw_bn254.NewCurve(api)
	if err != nil {
		return err
	}

	// Verify each signature
	for i := 0; i < len(circuit.Signatures); i++ {
		if i < len(circuit.PublicKeys) && i < len(circuit.Messages) {
			circuit.Signatures[i].Verify(curve, circuit.Messages[i], circuit.PublicKeys[i])
		}
	}

	return nil
}

// ThresholdSignatureCircuit verifies threshold signatures (k-of-n)
type ThresholdSignatureCircuit struct {
	// Public inputs
	Threshold  frontend.Variable `gnark:",public"` // Minimum required signatures
	NumSigners frontend.Variable `gnark:",public"` // Total number of possible signers
	Message    frontend.Variable `gnark:",public"`
	PublicKeys []eddsa.PublicKey `gnark:",public"`

	// Private inputs
	Signatures []eddsa.Signature   `gnark:",secret"`
	SignerMask []frontend.Variable `gnark:",secret"` // 1 if signer participated, 0 otherwise
}

func (circuit *ThresholdSignatureCircuit) Define(api frontend.API) error {
	utils := NewCircuitUtils(api, DefaultCircuitConfig())
	curve, err := sw_bn254.NewCurve(api)
	if err != nil {
		return err
	}

	validSignatures := frontend.Variable(0)

	for i := 0; i < len(circuit.SignerMask); i++ {
		// Ensure mask is binary
		api.AssertIsBoolean(circuit.SignerMask[i])

		// If this signer participated, verify their signature
		if i < len(circuit.Signatures) && i < len(circuit.PublicKeys) {
			// Conditionally verify signature only if mask[i] == 1
			// This is a simplified approach - in practice you'd need more sophisticated conditional verification
			signerParticipated := circuit.SignerMask[i]

			// Count valid signatures
			validSignatures = api.Add(validSignatures, signerParticipated)

			// Verify signature (this would need conditional logic in real implementation)
			circuit.Signatures[i].Verify(curve, circuit.Message, circuit.PublicKeys[i])
		}
	}

	// Must meet threshold
	utils.AssertGreaterEqualThan(validSignatures, circuit.Threshold)

	return nil
}
