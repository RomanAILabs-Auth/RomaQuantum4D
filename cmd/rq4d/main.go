// main.go
// Copyright RomanAILabs - Daniel Harding

// RQ4D HyperEngine Phase 2 — executes .rq4d scripts (lexer + Cl(4,0) gates + Z-axis telemetry).
package main

import (
	"fmt"
	"os"
	"sync"
	"time"

	gamath "romanailabs/rq4d/internal/math"
	"romanailabs/rq4d/internal/parser"
	"romanailabs/rq4d/internal/quantum"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: rq4d <script.rq4d>")
		os.Exit(2)
	}
	path := os.Args[1]

	t0 := time.Now()
	instrs, err := parser.ParseFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "parse: %v\n", err)
		os.Exit(1)
	}

	const ansiGreen = "\033[32m"
	const ansiReset = "\033[0m"
	fmt.Printf("%sExecuting RQ4D HyperEngine...%s\n", ansiGreen, ansiReset)

	var qubits []gamath.Multivector
	for k := 0; k < len(instrs); {
		ins := instrs[k]
		switch ins.Op {
		case parser.OpAlloc:
			qubits = make([]gamath.Multivector, ins.N)
			var wg sync.WaitGroup
			for i := range qubits {
				wg.Add(1)
				go func(idx int) {
					defer wg.Done()
					qubits[idx] = quantum.QubitZero()
				}(i)
			}
			wg.Wait()
			k++
		case parser.OpH:
			end := k
			for end < len(instrs) && instrs[end].Op == parser.OpH {
				end++
			}
			if err := parallelHadamard(qubits, instrs[k:end]); err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			k = end
		case parser.OpX:
			end := k
			for end < len(instrs) && instrs[end].Op == parser.OpX {
				end++
			}
			if err := parallelPauliX(qubits, instrs[k:end]); err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			k = end
		case parser.OpCNOT:
			if !bounds(qubits, ins.Ctrl) || !bounds(qubits, ins.Target) {
				fmt.Fprintf(os.Stderr, "CNOT: invalid indices %d %d\n", ins.Ctrl, ins.Target)
				os.Exit(1)
			}
			quantum.CNOTGate(qubits, ins.Ctrl, ins.Target)
			k++
		case parser.OpMeasure:
			quantum.Measure(os.Stdout, qubits)
			k++
		default:
			k++
		}
	}

	elapsed := time.Since(t0)
	ms := float64(elapsed.Nanoseconds()) / 1e6
	fmt.Printf("[Telemetry: Z-Axis = %.2fms]\n", ms)
}

func bounds(q []gamath.Multivector, i int) bool {
	return len(q) > 0 && i >= 0 && i < len(q)
}

func parallelHadamard(q []gamath.Multivector, block []parser.Instruction) error {
	H := quantum.Hadamard()
	var wg sync.WaitGroup
	for _, ins := range block {
		if !bounds(q, ins.N) {
			return fmt.Errorf("H: invalid target %d", ins.N)
		}
		tgt := ins.N
		wg.Add(1)
		go func(t int) {
			defer wg.Done()
			q[t] = quantum.ApplyGate(q[t], H)
		}(tgt)
	}
	wg.Wait()
	return nil
}

func parallelPauliX(q []gamath.Multivector, block []parser.Instruction) error {
	X := quantum.PauliXBitFlip()
	var wg sync.WaitGroup
	for _, ins := range block {
		if !bounds(q, ins.N) {
			return fmt.Errorf("X: invalid target %d", ins.N)
		}
		tgt := ins.N
		wg.Add(1)
		go func(t int) {
			defer wg.Done()
			q[t] = quantum.ApplyGate(q[t], X)
		}(tgt)
	}
	wg.Wait()
	return nil
}
