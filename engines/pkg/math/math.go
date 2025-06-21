package math

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
	"strconv"
	"strings"
)


var (
	Zero = big.NewInt(0)
	One  = big.NewInt(1)
	Two  = big.NewInt(2)
	Ten  = big.NewInt(10)

	DecimalPlaces = 18
	WeiPerEther   = new(big.Int).Exp(Ten, big.NewInt(18), nil)

	Secp256k1Prime, _ = new(big.Int).SetString("FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFEFFFFFC2F", 16)
	BN254Prime, _     = new(big.Int).SetString("21888242871839275222246405745257275088548364400416034343698204186575808495617", 10)

	MaxSafeInt = big.NewInt(9007199254740991) // 2^53 - 1
)

// ===== DECIMAL (FIXED-POINT) ARITHMETIC =====

type Decimal struct {
	value *big.Int // Internal value multiplied by 10^decimals
	scale int      // Number of decimal places
}

func NewDecimal(s string) (*Decimal, error) {
	parts := strings.Split(s, ".")
	if len(parts) > 2 {
		return nil, fmt.Errorf("invalid decimal format: %s", s)
	}

	integerPart := parts[0]
	var fractionalPart string
	if len(parts) == 2 {
		fractionalPart = parts[1]
	}

	if len(fractionalPart) > DecimalPlaces {
		fractionalPart = fractionalPart[:DecimalPlaces]
	} else {
		fractionalPart = fractionalPart + strings.Repeat("0", DecimalPlaces-len(fractionalPart))
	}

	combined := integerPart + fractionalPart
	value, ok := new(big.Int).SetString(combined, 10)
	if !ok {
		return nil, fmt.Errorf("invalid decimal format: %s", s)
	}

	return &Decimal{
		value: value,
		scale: DecimalPlaces,
	}, nil
}

func NewDecimalFromInt(i int64) *Decimal {
	value := new(big.Int).Mul(big.NewInt(i), WeiPerEther)
	return &Decimal{
		value: value,
		scale: DecimalPlaces,
	}
}

func NewDecimalFromBigInt(i *big.Int) *Decimal {
	value := new(big.Int).Mul(i, WeiPerEther)
	return &Decimal{
		value: new(big.Int).Set(value),
		scale: DecimalPlaces,
	}
}

func (d *Decimal) Add(other *Decimal) *Decimal {
	result := new(big.Int).Add(d.value, other.value)
	return &Decimal{
		value: result,
		scale: DecimalPlaces,
	}
}

func (d *Decimal) Sub(other *Decimal) *Decimal {
	result := new(big.Int).Sub(d.value, other.value)
	return &Decimal{
		value: result,
		scale: DecimalPlaces,
	}
}

func (d *Decimal) Mul(other *Decimal) *Decimal {
	result := new(big.Int).Mul(d.value, other.value)
	result.Div(result, WeiPerEther) // Adjust for double scaling
	return &Decimal{
		value: result,
		scale: DecimalPlaces,
	}
}

func (d *Decimal) Div(other *Decimal) *Decimal {
	if other.value.Cmp(Zero) == 0 {
		panic("division by zero")
	}

	// Multiply by scale factor before division to maintain precision
	result := new(big.Int).Mul(d.value, WeiPerEther)
	result.Div(result, other.value)

	return &Decimal{
		value: result,
		scale: DecimalPlaces,
	}
}

func (d *Decimal) String() string {
	str := d.value.String()

	negative := false
	if strings.HasPrefix(str, "-") {
		negative = true
		str = str[1:]
	}

	for len(str) <= DecimalPlaces {
		str = "0" + str
	}

	integerPart := str[:len(str)-DecimalPlaces]
	fractionalPart := str[len(str)-DecimalPlaces:]

	fractionalPart = strings.TrimRight(fractionalPart, "0")

	result := integerPart
	if fractionalPart != "" {
		result += "." + fractionalPart
	}

	if negative {
		result = "-" + result
	}

	return result
}

func (d *Decimal) ToFloat64() float64 {
	f, _ := strconv.ParseFloat(d.String(), 64)
	return f
}

func (d *Decimal) ToBigInt() *big.Int {
	return new(big.Int).Set(d.value)
}

func (d *Decimal) Cmp(other *Decimal) int {
	return d.value.Cmp(other.value)
}

func (d *Decimal) IsZero() bool {
	return d.value.Cmp(Zero) == 0
}

func (d *Decimal) IsPositive() bool {
	return d.value.Cmp(Zero) > 0
}

func (d *Decimal) IsNegative() bool {
	return d.value.Cmp(Zero) < 0
}

// ===== FINANCIAL MATH =====

func CalculateCompoundInterest(principal *Decimal, rate *Decimal, periods int64, periodsPerYear int64) *Decimal {
	// A = P(1 + r/n)^(nt)
	ratePerPeriod := rate.Div(NewDecimalFromInt(periodsPerYear))
	onePlusRate := ratePerPeriod.Add(NewDecimalFromInt(1))

	result := principal
	for i := int64(0); i < periods; i++ {
		result = result.Mul(onePlusRate)
	}

	return result
}

func CalculatePercentage(value *Decimal, percentage *Decimal) *Decimal {
	hundred := NewDecimalFromInt(100)
	return value.Mul(percentage).Div(hundred)
}

func CalculateYield(finalValue *Decimal, initialValue *Decimal) *Decimal {
	if initialValue.IsZero() {
		return NewDecimalFromInt(0)
	}

	gain := finalValue.Sub(initialValue)
	hundred := NewDecimalFromInt(100)
	return gain.Div(initialValue).Mul(hundred)
}

// ===== BIG INTEGER UTILITIES =====

func SafeAdd(a, b *big.Int) (*big.Int, error) {
	result := new(big.Int).Add(a, b)

	if result.Cmp(MaxSafeInt) > 0 {
		return nil, fmt.Errorf("integer overflow")
	}

	return result, nil
}

func SafeMul(a, b *big.Int) (*big.Int, error) {
	result := new(big.Int).Mul(a, b)

	if result.Cmp(MaxSafeInt) > 0 {
		return nil, fmt.Errorf("integer overflow")
	}

	return result, nil
}

func Gcd(a, b *big.Int) *big.Int {
	return new(big.Int).GCD(nil, nil, a, b)
}

func Lcm(a, b *big.Int) *big.Int {
	gcd := Gcd(a, b)
	result := new(big.Int).Mul(a, b)
	return result.Div(result, gcd)
}

func ModInverse(a, m *big.Int) *big.Int {
	return new(big.Int).ModInverse(a, m)
}

func ModPow(base, exp, mod *big.Int) *big.Int {
	return new(big.Int).Exp(base, exp, mod)
}

// ===== CRYPTOGRAPHIC OPERATIONS =====

func Hash(data []byte) []byte {
	hash := sha256.Sum256(data)
	return hash[:]
}

func HashString(s string) string {
	hash := Hash([]byte(s))
	return hex.EncodeToString(hash)
}

func CombineHashes(left, right []byte) []byte {
	combined := append(left, right...)
	return Hash(combined)
}

func GenerateRandomBigInt(max *big.Int) (*big.Int, error) {
	return rand.Int(rand.Reader, max)
}

func GenerateRandomBytes(length int) ([]byte, error) {
	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	return bytes, err
}

// ===== ZK PROOF MATH =====

type FieldElement struct {
	Value *big.Int
	Prime *big.Int
}

func NewFieldElement(value, prime *big.Int) *FieldElement {
	return &FieldElement{
		Value: new(big.Int).Mod(value, prime),
		Prime: new(big.Int).Set(prime),
	}
}

func (fe *FieldElement) Add(other *FieldElement) *FieldElement {
	result := new(big.Int).Add(fe.Value, other.Value)
	result.Mod(result, fe.Prime)
	return &FieldElement{
		Value: result,
		Prime: fe.Prime,
	}
}

func (fe *FieldElement) Mul(other *FieldElement) *FieldElement {
	result := new(big.Int).Mul(fe.Value, other.Value)
	result.Mod(result, fe.Prime)
	return &FieldElement{
		Value: result,
		Prime: fe.Prime,
	}
}

func (fe *FieldElement) Inverse() *FieldElement {
	inv := ModInverse(fe.Value, fe.Prime)
	return &FieldElement{
		Value: inv,
		Prime: fe.Prime,
	}
}

type Polynomial struct {
	Coefficients []*FieldElement
}

func NewPolynomial(coeffs []*FieldElement) *Polynomial {
	return &Polynomial{
		Coefficients: coeffs,
	}
}

func (p *Polynomial) Evaluate(x *FieldElement) *FieldElement {
	if len(p.Coefficients) == 0 {
		return NewFieldElement(Zero, x.Prime)
	}

	result := NewFieldElement(Zero, x.Prime)
	xPower := NewFieldElement(One, x.Prime)

	for _, coeff := range p.Coefficients {
		term := coeff.Mul(xPower)
		result = result.Add(term)
		xPower = xPower.Mul(x)
	}

	return result
}

// ===== UTILITY FUNCTIONS =====

func ParseBigInt(s string) (*big.Int, error) {
	var base int

	if strings.HasPrefix(s, "0x") || strings.HasPrefix(s, "0X") {
		base = 16
		s = s[2:]
	} else if strings.HasPrefix(s, "0b") || strings.HasPrefix(s, "0B") {
		base = 2
		s = s[2:]
	} else if strings.HasPrefix(s, "0o") || strings.HasPrefix(s, "0O") {
		base = 8
		s = s[2:]
	} else {
		base = 10
	}

	result, ok := new(big.Int).SetString(s, base)
	if !ok {
		return nil, fmt.Errorf("invalid number format: %s", s)
	}

	return result, nil
}

func ToHex(n *big.Int) string {
	return "0x" + n.Text(16)
}

func FromHex(s string) (*big.Int, error) {
	if strings.HasPrefix(s, "0x") || strings.HasPrefix(s, "0X") {
		s = s[2:]
	}

	result, ok := new(big.Int).SetString(s, 16)
	if !ok {
		return nil, fmt.Errorf("invalid hex format: %s", s)
	}

	return result, nil
}

func Min(a, b *big.Int) *big.Int {
	if a.Cmp(b) < 0 {
		return new(big.Int).Set(a)
	}
	return new(big.Int).Set(b)
}

func Max(a, b *big.Int) *big.Int {
	if a.Cmp(b) > 0 {
		return new(big.Int).Set(a)
	}
	return new(big.Int).Set(b)
}

func Abs(x *big.Int) *big.Int {
	result := new(big.Int).Set(x)
	return result.Abs(result)
}

// ===== PERCENTAGE AND RATIO CALCULATIONS =====

func CalculateRatio(numerator, denominator *Decimal) *Decimal {
	if denominator.IsZero() {
		return NewDecimalFromInt(0)
	}
	return numerator.Div(denominator)
}

func CalculateBPS(value *Decimal, basisPoints int64) *Decimal {
	bps := NewDecimalFromInt(basisPoints)
	tenThousand := NewDecimalFromInt(10000)
	return value.Mul(bps).Div(tenThousand)
}

// ===== VALIDATION FUNCTIONS =====

func IsValidDecimal(s string) bool {
	_, err := NewDecimal(s)
	return err == nil
}

func IsValidBigInt(s string) bool {
	_, err := ParseBigInt(s)
	return err == nil
}

func IsInRange(value, min, max *big.Int) bool {
	return value.Cmp(min) >= 0 && value.Cmp(max) <= 0
}
