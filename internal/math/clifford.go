// clifford.go
// Copyright RomanAILabs - Daniel Harding

// Package math implements Cl(4,0) multivectors (16 basis blades) and rotors.
// Layout mirrors Roma4D’s 4D story: four basis vectors e1..e4; blades indexed by
// bitmask (bit i => e_{i+1}). This is the Go-side twin to list[vec4] + rotor sweeps in .r4d.
package math

import (
	"math"
	"math/bits"
)

const (
	Dim = 16 // 2^4 blades for Cl(4,0)
)

// Multivector is a full Cl(4,0) element: scalar, 4 vectors, 6 bivectors, 4 trivectors, pseudoscalar.
// Index i corresponds to basis blade with mask i (e.g. 0=1, 1=e1, 2=e2, 3=e12, …, 15=e1234).
type Multivector struct {
	C [Dim]float64
}

// basisProduct returns the geometric product of basis blades a and b (as 4-bit masks).
func basisProduct(a, b uint8) (mask uint8, sign float64) {
	sign = 1.0
	r := a
	x := b
	for x != 0 {
		i := bits.TrailingZeros(uint(x))
		bit := uint8(1) << i
		lowOrEq := uint8((1 << (uint(i) + 1)) - 1)
		higher := r &^ lowOrEq
		if bits.OnesCount(uint(higher))%2 == 1 {
			sign = -sign
		}
		if r&bit != 0 {
			r &^= bit
			x &^= bit
		} else {
			r |= bit
			x &^= bit
		}
	}
	return r, sign
}

// GeometricProduct computes the full Clifford product AB in Cl(4,0) (+ + + + signature).
func GeometricProduct(a, b Multivector) (p Multivector) {
	for i := 0; i < Dim; i++ {
		if a.C[i] == 0 {
			continue
		}
		for j := 0; j < Dim; j++ {
			if b.C[j] == 0 {
				continue
			}
			k, s := basisProduct(uint8(i), uint8(j))
			p.C[k] += a.C[i] * b.C[j] * s
		}
	}
	return p
}

// Reverse returns the reversal ~M (sign (-1)^{k(k-1)/2} on grade k).
func Reverse(m Multivector) (r Multivector) {
	for i := 0; i < Dim; i++ {
		g := bits.OnesCount(uint(i))
		s := 1.0
		if g == 2 || g == 3 {
			s = -1.0
		}
		r.C[i] = m.C[i] * s
	}
	return r
}

// GradeProject returns the component living in the given grade (0..4).
func GradeProject(m Multivector, grade int) (g Multivector) {
	for i := 0; i < Dim; i++ {
		if bits.OnesCount(uint(i)) == grade {
			g.C[i] = m.C[i]
		}
	}
	return g
}

// NormSq is sum of squares of all components (Euclidean blade orthogonality).
func NormSq(m Multivector) float64 {
	var s float64
	for i := 0; i < Dim; i++ {
		s += m.C[i] * m.C[i]
	}
	return s
}

// Scale returns c * m.
func Scale(m Multivector, c float64) (o Multivector) {
	for i := 0; i < Dim; i++ {
		o.C[i] = m.C[i] * c
	}
	return o
}

// Normalize scales m to unit NormSq (no-op if zero).
func Normalize(m Multivector) Multivector {
	n := NormSq(m)
	if n < 1e-30 {
		return m
	}
	inv := 1.0 / math.Sqrt(n)
	return Scale(m, inv)
}

// Rotor is an even-grade element used to rotate via sandwiching: R M ~R.
// For operator-style gates (Hadamard on scalar |0⟩) the bridge may use left product R*M instead.
type Rotor struct {
	M Multivector
}

// Sandwich applies R * M * ~R (proper orthogonal action on the algebra).
func (r Rotor) Sandwich(m Multivector) Multivector {
	return GeometricProduct(GeometricProduct(r.M, m), Reverse(r.M))
}

// NewRotorFromBivectorPlane returns cos(θ/2) + sin(θ/2)*B for unit bivector B (single blade ±1).
func NewRotorFromBivectorPlane(bivectorIndex int, halfAngle float64) Rotor {
	var R Multivector
	c := math.Cos(halfAngle)
	s := math.Sin(halfAngle)
	R.C[0] = c
	if bivectorIndex >= 0 && bivectorIndex < Dim {
		R.C[bivectorIndex] = s
	}
	return Rotor{M: Normalize(R)}
}
