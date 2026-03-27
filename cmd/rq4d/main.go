// cmd/rq4d — RomaQuantum4D CLI (geometric simulation scale, honest telemetry).
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"romanailabs/rq4d/internal/parser"
	"romanailabs/rq4d/internal/quantum"
)

func main() {
	truthMode := flag.Bool("truth-mode", false, "Honest path: sequential H/X (no parallel batches), global pass after every gate line")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [--truth-mode] <script.rq4d>\n", filepath.Base(os.Args[0]))
		flag.PrintDefaults()
	}
	flag.Parse()
	args := flag.Args()
	if len(args) < 1 {
		flag.Usage()
		os.Exit(2)
	}
	path := args[0]

	fmt.Println("Executing RQ4D (geometric simulation scale, Cl(4,0) register)...")
	if *truthMode {
		fmt.Println("[truth-mode] Parallel H/X batches disabled; global pass runs after each gate line.")
	}

	insts, err := parser.ParseFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "parse error: %v\n", err)
		os.Exit(1)
	}

	t0 := time.Now()
	eng, err := quantum.Run(insts, *truthMode)
	if err != nil {
		fmt.Fprintf(os.Stderr, "run error: %v\n", err)
		os.Exit(1)
	}
	elapsed := time.Since(t0)

	n := len(eng.Qubits)
	var avgPassNs int64
	if eng.Stats.GlobalPassCount > 0 {
		avgPassNs = eng.Stats.GlobalPassNanos / int64(eng.Stats.GlobalPassCount)
	}

	fmt.Println()
	fmt.Println("--- Geometric simulation scale (honest telemetry) ---")
	fmt.Printf("Manifold register size:     %d multivector lanes (16 floats each)\n", n)
	fmt.Printf("Wall-clock (script):        %v\n", elapsed)
	fmt.Printf("Total gate ops executed:    %d\n", eng.Stats.TotalOps)
	fmt.Printf("Global passes:              %d\n", eng.Stats.GlobalPassCount)
	fmt.Printf("Global pass aggregate time: %v\n", time.Duration(eng.Stats.GlobalPassNanos))
	fmt.Printf("Time per global pass (avg): %v\n", time.Duration(avgPassNs))
	fmt.Printf("Cumulative bytes touched:   %d (global pass accounting)\n", eng.Stats.BytesTouched)
	fmt.Printf("Execution checksum (FNV1a): 0x%016x\n", eng.Stats.LastChecksum)
	fmt.Printf("Truth mode:                 %v\n", *truthMode)
	fmt.Println("------------------------------------------------------")
}
