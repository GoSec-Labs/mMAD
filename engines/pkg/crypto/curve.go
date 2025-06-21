package crypto

import (
	"math/big"

	"github.com/consensys/gnark-crypto/ecc/bn254"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
)

// Point represents a point on elliptic curve
type Point struct {
	X *big.Int
	Y *big.Int
}

// CurveManager handles elliptic curve operations
type CurveManager struct {
	curve *bn254.G1Affine
}

func NewCurveManager() *CurveManager {
	return &CurveManager{}
}

// GenerateRandomPoint generates a random point on BN254 curve
func (cm *CurveManager) GenerateRandomPoint() (*Point, error) {
	var scalar fr.Element
	scalar.SetRandom()

	var point bn254.G1Affine
	scalarBigInt := scalar.BigInt(new(big.Int))
	point.ScalarMultiplication(&point, scalarBigInt)

	return &Point{
		X: point.X.BigInt(new(big.Int)),
		Y: point.Y.BigInt(new(big.Int)),
	}, nil
}

// AddPoints adds two points on the curve
func (cm *CurveManager) AddPoints(p1, p2 *Point) (*Point, error) {
	var point1, point2, result bn254.G1Affine

	point1.X.SetBigInt(p1.X)
	point1.Y.SetBigInt(p1.Y)
	point2.X.SetBigInt(p2.X)
	point2.Y.SetBigInt(p2.Y)

	result.Add(&point1, &point2)

	return &Point{
		X: result.X.BigInt(new(big.Int)),
		Y: result.Y.BigInt(new(big.Int)),
	}, nil
}

// ScalarMultiply multiplies a point by a scalar
func (cm *CurveManager) ScalarMultiply(point *Point, scalar *big.Int) (*Point, error) {
	var p bn254.G1Affine
	var s fr.Element

	p.X.SetBigInt(point.X)
	p.Y.SetBigInt(point.Y)
	s.SetBigInt(scalar)

	var result bn254.G1Affine
	scalarBigInt := s.BigInt(new(big.Int))
	result.ScalarMultiplication(&p, scalarBigInt)

	return &Point{
		X: result.X.BigInt(new(big.Int)),
		Y: result.Y.BigInt(new(big.Int)),
	}, nil
}

// IsOnCurve checks if a point is on the BN254 curve
func (cm *CurveManager) IsOnCurve(point *Point) bool {
	var p bn254.G1Affine
	p.X.SetBigInt(point.X)
	p.Y.SetBigInt(point.Y)
	return p.IsOnCurve()
}

// FieldElement represents an element in the finite field
type FieldElement struct {
	value *fr.Element
}

// NewFieldElement creates a new field element
func NewFieldElement(value *big.Int) *FieldElement {
	var elem fr.Element
	elem.SetBigInt(value)
	return &FieldElement{value: &elem}
}

// Add adds two field elements
func (fe *FieldElement) Add(other *FieldElement) *FieldElement {
	var result fr.Element
	result.Add(fe.value, other.value)
	return &FieldElement{value: &result}
}

// Multiply multiplies two field elements
func (fe *FieldElement) Multiply(other *FieldElement) *FieldElement {
	var result fr.Element
	result.Mul(fe.value, other.value)
	return &FieldElement{value: &result}
}

// ToBigInt converts field element to big.Int
func (fe *FieldElement) ToBigInt() *big.Int {
	return fe.value.BigInt(new(big.Int))
}
