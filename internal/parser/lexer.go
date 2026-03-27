// lexer.go
// Copyright RomanAILabs - Daniel Harding

// Package parser implements a line-oriented lexer for .rq4d scripts (ALLOC, H, X, CNOT, MEASURE).
package parser

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Op is a script opcode.
type Op byte

const (
	OpAlloc Op = 1 + iota
	OpH
	OpX
	OpCNOT
	OpMeasure
)

// Instruction is one decoded line.
type Instruction struct {
	Op     Op
	N      int // ALLOC count or single target for H/X
	Ctrl   int // CNOT control
	Target int // CNOT target
}

// ParseFile reads path line-by-line, skips blanks and # comments, returns instructions.
func ParseFile(path string) ([]Instruction, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var out []Instruction
	sc := bufio.NewScanner(f)
	const maxLine = 512 * 1024
	buf := make([]byte, 0, 64*1024)
	sc.Buffer(buf, maxLine)
	lineNo := 0
	for sc.Scan() {
		lineNo++
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		ins, err := parseLine(line)
		if err != nil {
			return nil, fmt.Errorf("%s:%d: %w", path, lineNo, err)
		}
		out = append(out, ins)
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func parseLine(line string) (Instruction, error) {
	fields := strings.Fields(line)
	if len(fields) == 0 {
		return Instruction{}, fmt.Errorf("empty instruction")
	}
	op := strings.ToUpper(fields[0])
	switch op {
	case "ALLOC":
		if len(fields) != 2 {
			return Instruction{}, fmt.Errorf("ALLOC: want ALLOC [n]")
		}
		n, err := strconv.Atoi(fields[1])
		if err != nil || n < 1 || n > 1<<20 {
			return Instruction{}, fmt.Errorf("ALLOC: invalid n")
		}
		return Instruction{Op: OpAlloc, N: n}, nil
	case "H":
		if len(fields) != 2 {
			return Instruction{}, fmt.Errorf("H: want H [target]")
		}
		t, err := strconv.Atoi(fields[1])
		if err != nil || t < 0 {
			return Instruction{}, fmt.Errorf("H: invalid target")
		}
		return Instruction{Op: OpH, N: t}, nil
	case "X":
		if len(fields) != 2 {
			return Instruction{}, fmt.Errorf("X: want X [target]")
		}
		t, err := strconv.Atoi(fields[1])
		if err != nil || t < 0 {
			return Instruction{}, fmt.Errorf("X: invalid target")
		}
		return Instruction{Op: OpX, N: t}, nil
	case "CNOT":
		if len(fields) != 3 {
			return Instruction{}, fmt.Errorf("CNOT: want CNOT [control] [target]")
		}
		c, err := strconv.Atoi(fields[1])
		if err != nil || c < 0 {
			return Instruction{}, fmt.Errorf("CNOT: invalid control")
		}
		t, err := strconv.Atoi(fields[2])
		if err != nil || t < 0 {
			return Instruction{}, fmt.Errorf("CNOT: invalid target")
		}
		return Instruction{Op: OpCNOT, Ctrl: c, Target: t}, nil
	case "MEASURE":
		if len(fields) != 1 {
			return Instruction{}, fmt.Errorf("MEASURE: takes no arguments")
		}
		return Instruction{Op: OpMeasure}, nil
	default:
		return Instruction{}, fmt.Errorf("unknown opcode %q", op)
	}
}
