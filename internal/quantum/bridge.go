// bridge.go — RomaQuantum4D geometric simulator (Cl(4,0), global coupling).
// Copyright RomanAILabs — Daniel Harding

package quantum

import (
	"encoding/binary"
	"fmt"
	"hash/fnv"
	"math"
	"runtime"
	"sync"
	"time"

	gamath "github.com/RomanAILabs-Auth/RomaQuantum4D/internal/math"
	"github.com/RomanAILabs-Auth/RomaQuantum4D/internal/parser"
)

const (
	rippleDecay    = 0.07
	rippleStrength = 1e-5
	longRangeScale = 1e-7
	interfereBlend = 0.012
	gradeDecay     = 8e-5
	phaseCoupling  = 1e-4
	coherenceGate  = 0.9997
)

// GlobalSystem is shared manifold state; every gate and pass reads/writes it.
type GlobalSystem struct {
	GlobalPhase   float64
	EnergyField   []float64
	Coherence     float64
	SystemEntropy float64
}

// Stats records verifiable work metrics (traceable, not hype).
type Stats struct {
	TotalOps        uint64
	GlobalPassCount uint64
	GlobalPassNanos int64
	BytesTouched    uint64
	LastChecksum    uint64
}

// Engine holds the full register plus global coupling.
type Engine struct {
	Qubits    []gamath.Multivector
	Global    *GlobalSystem
	TruthMode bool
	Stats     Stats
	// TraceHash is a rolling mix of script opcodes and operands in file order (deterministic).
	TraceHash uint64
}

// NewEngine allocates n qubits in |0⟩ and global fields of length n.
func NewEngine(n int) *Engine {
	if n < 1 {
		n = 1
	}
	g := &GlobalSystem{
		EnergyField: make([]float64, n),
		Coherence:   1.0,
	}
	q := make([]gamath.Multivector, n)
	for i := range q {
		q[i].C[0] = 1.0
		g.EnergyField[i] = 1.0
	}
	return &Engine{Qubits: q, Global: g}
}

func probOne(m gamath.Multivector) float64 {
	n := gamath.NormSq(m)
	if n < 1e-30 {
		return 0
	}
	return (m.C[1] * m.C[1]) / n
}

// mixTraceInstr folds this script line into TraceHash (order-sensitive, reproducible).
func (e *Engine) mixTraceInstr(ins parser.Instruction) {
	var b [25]byte
	b[0] = byte(ins.Op)
	binary.LittleEndian.PutUint64(b[1:9], uint64(ins.N))
	binary.LittleEndian.PutUint64(b[9:17], uint64(ins.Ctrl))
	binary.LittleEndian.PutUint64(b[17:25], uint64(ins.Target))
	x := fnv.New64a()
	_, _ = x.Write(b[:])
	block := x.Sum64()
	e.TraceHash ^= block
	e.TraceHash *= 1099511628211
}

func applyHadamard(m *gamath.Multivector) {
	rt2 := math.Sqrt2
	a0, a1 := m.C[0], m.C[1]
	m.C[0] = (a0 + a1) / rt2
	m.C[1] = (a0 - a1) / rt2
}

func applyPauliX(m *gamath.Multivector) {
	m.C[0], m.C[1] = m.C[1], m.C[0]
}

func (e *Engine) touchGlobalFromGate(qidx int) {
	g := e.Global
	g.GlobalPhase = math.Mod(g.GlobalPhase+phaseCoupling*float64(qidx+1), 2*math.Pi)
	if qidx >= 0 && qidx < len(g.EnergyField) {
		g.EnergyField[qidx] = gamath.NormSq(e.Qubits[qidx]) / float64(gamath.Dim)
	}
	g.Coherence *= coherenceGate
	if g.Coherence < 1e-6 {
		g.Coherence = 1e-6
	}
}

// deterministicSpread is a reproducible [0,1) pseudo-variate from global + local state (no math/rand).
func (e *Engine) deterministicSpread(control, target int) float64 {
	h := fnv.New64a()
	var buf [8]byte
	binary.LittleEndian.PutUint64(buf[:], e.TraceHash)
	_, _ = h.Write(buf[:])
	binary.LittleEndian.PutUint64(buf[:], math.Float64bits(e.Global.GlobalPhase))
	_, _ = h.Write(buf[:])
	binary.LittleEndian.PutUint64(buf[:], math.Float64bits(e.Global.Coherence))
	_, _ = h.Write(buf[:])
	m := e.Qubits[control]
	for k := 0; k < gamath.Dim; k++ {
		binary.LittleEndian.PutUint64(buf[:], math.Float64bits(m.C[k]))
		_, _ = h.Write(buf[:])
	}
	binary.LittleEndian.PutUint64(buf[:], uint64(control))
	_, _ = h.Write(buf[:])
	binary.LittleEndian.PutUint64(buf[:], uint64(target))
	_, _ = h.Write(buf[:])
	u := h.Sum64()
	return float64(u%1_000_000) / 1e6
}

// cnotShouldFlipTarget blends local amplitude, global phase/coherence/energy field, and a deterministic spread.
// It is not driven by P(|1⟩) > ½ alone.
func (e *Engine) cnotShouldFlipTarget(control, target int) bool {
	p1 := probOne(e.Qubits[control])
	g := e.Global
	var ef float64
	if control >= 0 && control < len(g.EnergyField) {
		ef = g.EnergyField[control]
	}
	phaseWindow := 0.5 + 0.5*math.Sin(g.GlobalPhase+float64(control)*0.17+float64(target)*0.09)
	coh := g.Coherence
	if coh < 1e-9 {
		coh = 1e-9
	}
	score := p1*0.38 + ef*0.28 + phaseWindow*0.22 + coh*0.12*(0.5+0.5*p1)
	spread := e.deterministicSpread(control, target)
	score += (spread - 0.5) * 0.14 * coh
	return score > 0.5
}

// applyCNOTLocal conditional flip (field-blended) + full-register ripple (O(n)).
func (e *Engine) applyCNOTLocal(control, target int) {
	n := len(e.Qubits)
	if control < 0 || control >= n || target < 0 || target >= n {
		return
	}
	p1 := probOne(e.Qubits[control])
	if e.cnotShouldFlipTarget(control, target) {
		applyPauliX(&e.Qubits[target])
	}
	ctrl := e.Qubits[control]
	coher := e.Global.Coherence
	if coher < 1e-9 {
		coher = 1e-9
	}
	invN := 1.0 / float64(n)
	for j := 0; j < n; j++ {
		d := j - control
		if d < 0 {
			d = -d
		}
		w := p1 * math.Exp(-rippleDecay*float64(d)) * coher
		tail := p1 * coher * longRangeScale * invN
		for k := 0; k < gamath.Dim; k++ {
			coupling := (w + tail) * rippleStrength
			e.Qubits[j].C[k] += coupling * ctrl.C[k]
		}
	}
	e.Global.SystemEntropy += p1 * 1e-6
}

func (e *Engine) applyGlobalPass() {
	start := time.Now()
	g := e.Global
	n := len(e.Qubits)
	var accPhase float64

	for i := 0; i < n; i++ {
		m := &e.Qubits[i]
		var norm float64
		for k := 0; k < gamath.Dim; k++ {
			v := m.C[k]
			norm += v * v
		}
		if norm > 1e-30 {
			inv := 1.0 / math.Sqrt(norm)
			for k := 0; k < gamath.Dim; k++ {
				m.C[k] *= inv
			}
		}

		b := g.Coherence * interfereBlend
		a0, a1 := m.C[0], m.C[1]
		m.C[0] = a0*(1-b) + a1*b
		m.C[1] = a1*(1-b) + a0*b

		for k := 2; k < gamath.Dim; k++ {
			m.C[k] *= (1.0 - gradeDecay)
		}

		*m = gamath.Normalize(*m)
		ns := gamath.NormSq(*m)
		g.EnergyField[i] = ns / float64(gamath.Dim)
		accPhase += ns * float64(i+1) * 1e-6
	}

	g.GlobalPhase = math.Mod(g.GlobalPhase+accPhase, 2*math.Pi)
	g.Coherence = g.Coherence*(1.0-1e-5) + 1e-5

	e.Stats.GlobalPassCount++
	e.Stats.GlobalPassNanos += time.Since(start).Nanoseconds()
	bytesPer := uint64(gamath.Dim * 8)
	e.Stats.BytesTouched += uint64(n) * bytesPer * 2
}

// ChecksumFNV includes register, globals, execution-order trace, and counters.
func (e *Engine) ChecksumFNV() uint64 {
	h := fnv.New64a()
	var buf [8]byte
	writeF := func(f float64) {
		binary.LittleEndian.PutUint64(buf[:], math.Float64bits(f))
		_, _ = h.Write(buf[:])
	}
	writeU64 := func(v uint64) {
		binary.LittleEndian.PutUint64(buf[:], v)
		_, _ = h.Write(buf[:])
	}

	g := e.Global
	writeF(g.GlobalPhase)
	writeF(g.Coherence)
	writeF(g.SystemEntropy)
	for _, ef := range g.EnergyField {
		writeF(ef)
	}
	for i := range e.Qubits {
		for k := 0; k < gamath.Dim; k++ {
			writeF(e.Qubits[i].C[k])
		}
	}
	writeU64(e.TraceHash)
	writeU64(e.Stats.TotalOps)
	writeU64(e.Stats.GlobalPassCount)
	return h.Sum64()
}

func parallelApplyH(q []gamath.Multivector, indices []int) {
	parallelGate(q, indices, applyHadamard)
}

func parallelApplyX(q []gamath.Multivector, indices []int) {
	parallelGate(q, indices, func(m *gamath.Multivector) { applyPauliX(m) })
}

func parallelGate(q []gamath.Multivector, indices []int, fn func(*gamath.Multivector)) {
	if len(indices) == 0 {
		return
	}
	w := runtime.GOMAXPROCS(0)
	if w < 1 {
		w = 1
	}
	if len(indices) <= w {
		for _, idx := range indices {
			if idx >= 0 && idx < len(q) {
				fn(&q[idx])
			}
		}
		return
	}
	chunk := (len(indices) + w - 1) / w
	var wg sync.WaitGroup
	for c := 0; c < w; c++ {
		lo := c * chunk
		if lo >= len(indices) {
			break
		}
		hi := lo + chunk
		if hi > len(indices) {
			hi = len(indices)
		}
		wg.Add(1)
		go func(lo, hi int) {
			defer wg.Done()
			for _, idx := range indices[lo:hi] {
				if idx >= 0 && idx < len(q) {
					fn(&q[idx])
				}
			}
		}(lo, hi)
	}
	wg.Wait()
}

func (e *Engine) applySequentialH(idx int) {
	if idx >= 0 && idx < len(e.Qubits) {
		applyHadamard(&e.Qubits[idx])
	}
}

func (e *Engine) applySequentialX(idx int) {
	if idx >= 0 && idx < len(e.Qubits) {
		applyPauliX(&e.Qubits[idx])
	}
}

func (e *Engine) flushHBatch(indices []int) {
	if len(indices) == 0 {
		return
	}
	if e.TruthMode {
		for _, idx := range indices {
			e.applySequentialH(idx)
			e.touchGlobalFromGate(idx)
			e.Stats.TotalOps++
			e.applyGlobalPass()
		}
		return
	}
	parallelApplyH(e.Qubits, indices)
	for _, idx := range indices {
		e.touchGlobalFromGate(idx)
		e.Stats.TotalOps++
	}
	e.applyGlobalPass()
}

func (e *Engine) flushXBatch(indices []int) {
	if len(indices) == 0 {
		return
	}
	if e.TruthMode {
		for _, idx := range indices {
			e.applySequentialX(idx)
			e.touchGlobalFromGate(idx)
			e.Stats.TotalOps++
			e.applyGlobalPass()
		}
		return
	}
	parallelApplyX(e.Qubits, indices)
	for _, idx := range indices {
		e.touchGlobalFromGate(idx)
		e.Stats.TotalOps++
	}
	e.applyGlobalPass()
}

// MeasureAll prints per-qubit geometric probabilities; couples each lane to globals.
func (e *Engine) MeasureAll() {
	g := e.Global
	n := len(e.Qubits)
	for i := 0; i < n; i++ {
		m := gamath.Normalize(e.Qubits[i])
		e.Qubits[i] = m
		p0 := m.C[0] * m.C[0]
		p1 := m.C[1] * m.C[1]
		fmt.Printf("MEASURE q[%d]: P(|0>)=%.4f P(|1>)=%.4f\n", i, p0, p1)
		g.SystemEntropy += p0 * p1
		e.touchGlobalFromGate(i)
	}
	g.Coherence *= 0.999
	e.applyGlobalPass()
}

// Run executes instructions: first must be ALLOC; respects truth-mode batching.
func Run(instructions []parser.Instruction, truthMode bool) (*Engine, error) {
	var n int
	var started bool
	var hBatch, xBatch []int
	flush := func(e *Engine) {
		e.flushHBatch(hBatch)
		hBatch = hBatch[:0]
		e.flushXBatch(xBatch)
		xBatch = xBatch[:0]
	}

	var eng *Engine
	for _, ins := range instructions {
		switch ins.Op {
		case parser.OpAlloc:
			if started {
				return nil, fmt.Errorf("duplicate ALLOC")
			}
			n = ins.N
			eng = NewEngine(n)
			eng.TruthMode = truthMode
			eng.mixTraceInstr(ins)
			eng.applyGlobalPass()
			started = true
		case parser.OpH:
			if eng == nil {
				return nil, fmt.Errorf("H before ALLOC")
			}
			if ins.N >= n {
				return nil, fmt.Errorf("H: target %d out of range (n=%d)", ins.N, n)
			}
			eng.mixTraceInstr(ins)
			if len(xBatch) > 0 {
				eng.flushXBatch(xBatch)
				xBatch = xBatch[:0]
			}
			hBatch = append(hBatch, ins.N)
		case parser.OpX:
			if eng == nil {
				return nil, fmt.Errorf("X before ALLOC")
			}
			if ins.N >= n {
				return nil, fmt.Errorf("X: target %d out of range (n=%d)", ins.N, n)
			}
			eng.mixTraceInstr(ins)
			if len(hBatch) > 0 {
				eng.flushHBatch(hBatch)
				hBatch = hBatch[:0]
			}
			xBatch = append(xBatch, ins.N)
		case parser.OpCNOT:
			if eng == nil {
				return nil, fmt.Errorf("CNOT before ALLOC")
			}
			flush(eng)
			eng.mixTraceInstr(ins)
			if ins.Ctrl >= n || ins.Target >= n {
				return nil, fmt.Errorf("CNOT: indices out of range")
			}
			eng.applyCNOTLocal(ins.Ctrl, ins.Target)
			eng.touchGlobalFromGate(ins.Ctrl)
			eng.touchGlobalFromGate(ins.Target)
			eng.Stats.TotalOps++
			eng.applyGlobalPass()
		case parser.OpMeasure:
			if eng == nil {
				return nil, fmt.Errorf("MEASURE before ALLOC")
			}
			flush(eng)
			eng.mixTraceInstr(ins)
			eng.MeasureAll()
			eng.Stats.TotalOps++
		}
	}
	if eng == nil {
		return nil, fmt.Errorf("missing ALLOC")
	}
	flush(eng)
	eng.Stats.LastChecksum = eng.ChecksumFNV()
	return eng, nil
}
