// bridge.go
// Copyright RomanAILabs - Daniel Harding

// Package quantum maps geometric qubit states into Cl(4,0) and applies gates as multivector products.
// Roma4D alignment: gates behave like rotor sweeps (p = p * rot in .r4d); here we use Cl(4,0)
// products on multivector “particles” without complex matrices or density matrices.
package quantum

import (
	"fmt"
	"io"
	stdmath "math"
	"sync"

	gamath "romanailabs/rq4d/internal/math"
)

const (
	bladeScalar      = 0
	bladeE1          = 1
	bladeE2          = 2
	bladeE12         = 3
	sqrt1d2          = 0.7071067811865476 // 1/sqrt(2)
)

// QubitZero returns |0⟩ as scalar 1 (all other blades zero), per spec.
func QubitZero() gamath.Multivector {
	var m gamath.Multivector
	m.C[bladeScalar] = 1.0
	return m
}

// QubitOne returns |1⟩ as vector e1 — orthogonal partner to |0⟩ in this GA encoding.
func QubitOne() gamath.Multivector {
	var m gamath.Multivector
	m.C[bladeE1] = 1.0
	return m
}

// Hadamard returns a Cl(4,0) operator as a Rotor wrapper: left-multiplying |0⟩ yields (|0⟩+|1⟩)/√2
// in this encoding (scalar + e1), i.e. geometric superposition without complex amplitudes.
func Hadamard() gamath.Rotor {
	var H gamath.Multivector
	H.C[bladeScalar] = sqrt1d2
	H.C[bladeE1] = sqrt1d2
	return gamath.Rotor{M: H}
}

// PauliX is a π rotation in the e1∧e2 plane (sandwich swaps e1↔e2 in that subspace).
func PauliX() gamath.Rotor {
	var R gamath.Multivector
	R.C[bladeE12] = 1.0
	return gamath.Rotor{M: gamath.Normalize(R)}
}

// PauliXBitFlip is computational NOT in the scalar/e1 encoding: e1 * 1 = e1, e1 * e1 = 1.
// CNOT and script opcode X use this (not the e12 sandwich rotor).
func PauliXBitFlip() gamath.Rotor {
	var R gamath.Multivector
	R.C[bladeE1] = 1.0
	return gamath.Rotor{M: R}
}

// ApplyGate applies the gate to the state using geometric product order: GeometricProduct(rotor.M, state).
// The result is normalized for a stable “probability mass” analogue (unit norm in R^16).
func ApplyGate(state gamath.Multivector, gate gamath.Rotor) gamath.Multivector {
	out := gamath.GeometricProduct(gate.M, state)
	return gamath.Normalize(out)
}

// ApplyGateSandwich applies R * state * ~R (useful for Pauli-X on vector-encoded states).
func ApplyGateSandwich(state gamath.Multivector, gate gamath.Rotor) gamath.Multivector {
	out := gate.Sandwich(state)
	return gamath.Normalize(out)
}

// BlochPhase returns a cheap real “phase proxy” atan2(scalar, e1) for telemetry (not complex arg).
func BlochPhase(m gamath.Multivector) float64 {
	return stdmath.Atan2(m.C[bladeE1], m.C[bladeScalar])
}

// ctrlOneEps treats the control as |1⟩ only when P(|1⟩) is essentially 1 (avoids spurious flips on |+⟩).
const ctrlOneEps = 1e-6

// ProbComputationalOne returns P(|1⟩) ≈ e1 blade energy / total norm² (single-qubit readout proxy).
func ProbComputationalOne(m gamath.Multivector) float64 {
	n := gamath.NormSq(m)
	if n < 1e-18 {
		return 0
	}
	return (m.C[bladeE1] * m.C[bladeE1]) / n
}

// CNOTGate applies a Pauli-X bit-flip to the target iff the control multivector is (approximately) |1⟩.
// Independent multivectors cannot model full entanglement; this is the conditional-rotor stepping stone.
func CNOTGate(qubits []gamath.Multivector, control, target int) {
	if len(qubits) == 0 || control == target {
		return
	}
	if control < 0 || target < 0 || control >= len(qubits) || target >= len(qubits) {
		return
	}
	p1 := ProbComputationalOne(qubits[control])
	if p1 > 1.0-ctrlOneEps {
		qubits[target] = ApplyGate(qubits[target], PauliXBitFlip())
	}
}

// measSlot holds parallel measurement aggregation output.
type measSlot struct {
	p0, p1, other float64
}

// Measure writes per-qubit computational-basis probabilities (scalar = |0⟩, e1 = |1⟩); other blades summed as leakage.
// Uses a WaitGroup to mirror par-for style aggregation across lanes.
func Measure(w io.Writer, qubits []gamath.Multivector) {
	n := len(qubits)
	if n == 0 {
		fmt.Fprintln(w, "MEASURE: (no qubits)")
		return
	}
	slots := make([]measSlot, n)
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			m := qubits[idx]
			ns := gamath.NormSq(m)
			if ns < 1e-18 {
				return
			}
			inv := 1.0 / ns
			p0 := m.C[bladeScalar] * m.C[bladeScalar] * inv
			p1 := m.C[bladeE1] * m.C[bladeE1] * inv
			other := 1.0 - p0 - p1
			if other < 0 {
				other = 0
			}
			slots[idx] = measSlot{p0: p0, p1: p1, other: other}
		}(i)
	}
	wg.Wait()
	for i := 0; i < n; i++ {
		s := slots[i]
		fmt.Fprintf(w, "MEASURE q[%d]  P(|0>)=%.6f  P(|1>)=%.6f  P(other)=%.6f\n", i, s.p0, s.p1, s.other)
	}
}
