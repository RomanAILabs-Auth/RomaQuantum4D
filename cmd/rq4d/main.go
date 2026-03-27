// cmd/rq4d — RomaQuantum4D CLI (geometric simulation scale, traceable telemetry).
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/RomanAILabs-Auth/RomaQuantum4D/internal/parser"
	"github.com/RomanAILabs-Auth/RomaQuantum4D/internal/quantum"
)

func main() {
	truthMode := flag.Bool("truth-mode", false, "Strict execution: no H/X batching; one O(n) global pass per gate line")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [flags] <file.rq4d>\n", filepath.Base(os.Args[0]))
		flag.PrintDefaults()
	}
	flag.Parse()
	args := flag.Args()
	if len(args) < 1 {
		flag.Usage()
		os.Exit(2)
	}
	path := args[0]
	if fi, err := os.Stat(path); err != nil || fi.IsDir() {
		if err != nil && os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "rq4d: script not found: %s\n", path)
			os.Exit(2)
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "rq4d: cannot open script: %v\n", err)
			os.Exit(2)
		}
		fmt.Fprintf(os.Stderr, "rq4d: path is not a file: %s\n", path)
		os.Exit(2)
	}

	fmt.Println("Executing RQ4D (geometric simulation scale, Cl(4,0) register)...")
	if *truthMode {
		fmt.Println("[truth-mode] Field update: strict per-line global pass; H/X batching disabled.")
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
	elapsedSec := elapsed.Seconds()
	if elapsedSec < 1e-12 {
		elapsedSec = 1e-12
	}

	n := len(eng.Qubits)
	var avgPassNs int64
	if eng.Stats.GlobalPassCount > 0 {
		avgPassNs = eng.Stats.GlobalPassNanos / int64(eng.Stats.GlobalPassCount)
	}
	derivedOpsPerSec := float64(eng.Stats.TotalOps) / elapsedSec

	fmt.Println()
	fmt.Println("--- Geometric simulation scale (run summary) ---")
	fmt.Printf("Qubit count (lanes):        %d (16 float64 components each)\n", n)
	fmt.Printf("Elapsed wall time:          %v\n", elapsed)
	fmt.Printf("Total gates executed:       %d\n", eng.Stats.TotalOps)
	fmt.Printf("Global passes executed:     %d\n", eng.Stats.GlobalPassCount)
	fmt.Printf("Global pass aggregate time: %v\n", time.Duration(eng.Stats.GlobalPassNanos))
	fmt.Printf("Time per global pass (avg): %v\n", time.Duration(avgPassNs))
	fmt.Printf("Memory touched (estimate):  %d bytes (global-pass accounting)\n", eng.Stats.BytesTouched)
	fmt.Printf("Derived metric:             %.6g gate ops/s (TotalOps / wall time)\n", derivedOpsPerSec)
	fmt.Printf("Execution checksum (FNV1a): 0x%016x\n", eng.Stats.LastChecksum)
	fmt.Printf("Truth mode:                 %v\n", *truthMode)
	fmt.Println("State propagation completed.")
	fmt.Println("----------------------------------------------")
}
