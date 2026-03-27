# RQ4D Master Guide

**RomanAILabs — Daniel Harding**

Single reference for **install**, **daily use**, **programming / extending** the engine, and **LLM-assisted development** of [RomaQuantum4D](https://github.com/RomanAILabs-Auth/RomaQuantum4D) (RQ4D HyperEngine).

---

## Table of contents

1. [Install guide](#1-install-guide)  
2. [User guide](#2-user-guide)  
3. [Programming guide](#3-programming-guide)  
4. [LLM briefing pack](#4-llm-briefing-pack-advanced)  
5. [Quick reference](#5-quick-reference)

---

## 1. Install guide

### 1.1 Prerequisites

| Requirement | Notes |
|-------------|--------|
| **Go** | **1.22+** (`go version`). [https://go.dev/dl/](https://go.dev/dl/) |
| **Git** | To clone the repository. |
| **PowerShell** | **5.1+** or **PowerShell 7+** for `scripts/RQ4D_World_Record.ps1` (optional). |

RQ4D is **pure Go** — no CGO, no external quantum SDKs, no Python runtime for the core engine.

### 1.2 Clone

```bash
git clone https://github.com/RomanAILabs-Auth/RomaQuantum4D.git
cd RomaQuantum4D
```

Your folder name may be `RQ4D` or `RomaQuantum4D` depending on how you cloned; the **module root** is the directory that contains `go.mod`.

### 1.3 Verify toolchain

```bash
go version
go build -o rq4d ./cmd/rq4d        # Unix / macOS binary name
go build -o rq4d.exe ./cmd/rq4d    # Windows
```

On Windows, `rq4d.exe` is gitignored at repo root when named `/rq4d.exe` — local builds are fine.

### 1.4 First run (smoke test)

```bash
go run ./cmd/rq4d examples/manifold_sweep.rq4d
```

You should see a green **Executing RQ4D HyperEngine...** line (ANSI), eight measurement lines, and **`[Telemetry: Z-Axis = X.XXms]`**.

### 1.5 Optional: world-record style demo script

From repo root:

```powershell
pwsh -ExecutionPolicy Bypass -File .\scripts\RQ4D_World_Record.ps1
```

This builds the binary, generates a large `.rq4d` under `examples/`, and runs it. Large `-QubitCount` values create huge scripts and may stress memory and goroutine counts — start with defaults, then scale up.

### 1.6 Troubleshooting

| Symptom | Likely cause | Action |
|---------|----------------|--------|
| `go: cannot find module` | Not at module root | `cd` to folder containing `go.mod`. |
| Parse error on script | Wrong path or bad opcode | Check spelling; see [§5](#5-quick-reference). |
| Green banner shows as escape codes | Terminal without ANSI | Use Windows Terminal, or ignore raw `\033[32m`. |
| Slow / OOM on huge scripts | Millions of `H` lines batched as one parallel wave | Reduce qubit count or use `-GenerateOnly` on the PS1 script. |

### 1.7 Relation to **Roma4D** (`.r4d`)

- **RQ4D** runs **`.rq4d` text scripts** with the **RQ4D instruction set** (ALLOC, H, X, CNOT, MEASURE).  
- **Roma4D** (`r4d`) compiles **`.r4d` source language** (different syntax: `def`, `vec4`, `rotor`, `par for`, etc.).  
- Files like `examples/spacetime_ui_v3.r4d` in this repo are **Roma4D-style samples** for narrative alignment; execute them with **`r4d`**, not **`rq4d`**.

---

## 2. User guide

### 2.1 Command-line interface

```text
rq4d <path-to-script.rq4d>
```

- **One argument**: path to a script file.  
- **No argument**: prints `usage: rq4d <script.rq4d>` and exits with code **2**.  
- **Parse failure**: message on stderr, exit **1**.

Execution order is **strictly sequential** by instruction type, with **internal parallel batches** where noted below.

### 2.2 Script format (`.rq4d` for RQ4D engine)

- **One instruction per line** (whitespace-separated tokens).  
- **Case-insensitive** opcodes (`H`, `h`, `H` are equivalent).  
- **Comments**: lines starting with `#` are ignored.  
- **Blank lines** ignored.

### 2.3 Instruction set

| Opcode | Syntax | Meaning |
|--------|--------|---------|
| **ALLOC** | `ALLOC n` | Allocate `n` qubits, each initialized to **\|0⟩** (scalar `1` in Cl(4,0) encoding). Parallel zero-init. |
| **H** | `H i` | Hadamard on qubit index `i`. **Consecutive `H` lines** are executed in **one parallel batch** (goroutines). |
| **X** | `X i` | Computational **NOT** (bit flip via left multiply by `e₁`). Consecutive `X` lines are parallel-batched. |
| **CNOT** | `CNOT c t` | If control `c` is **definitely \|1⟩** (probability ≈ 1 on the `e₁` blade), apply **X** on target `t`. See §3.4. |
| **MEASURE** | `MEASURE` | Print per-qubit **P(\|0⟩)**, **P(\|1⟩)**, **P(other)** from multivector energy; parallel reduction across indices. |

Indices are **0-based**. **`ALLOC` must appear before** gates that use qubits.

### 2.4 Built-in examples (paths relative to repo root)

| File | Intent |
|------|--------|
| `examples/manifold_sweep.rq4d` | 8 qubits, parallel **H**, **MEASURE** — flagship demo. |
| `examples/parallel_h.rq4d` | 4-way parallel **H** batch. |
| `examples/cnot_demo.rq4d` | **X** on control, **CNOT**, both in **\|1⟩** (product state). |
| `examples/spacetime_ui_v3.r4d` | Roma4D **worldtube** demo (run with **`r4d`**, not `rq4d`). |

### 2.5 Output semantics

- After successful parse, the engine prints **`Executing RQ4D HyperEngine...`** in green (ANSI).  
- **MEASURE** lines look like:  
  `MEASURE q[k]  P(|0>)=...  P(|1>)=...  P(other)=...`  
- After **all** instructions, wall-clock telemetry:  
  **`[Telemetry: Z-Axis = X.XXms]`**  
  This covers **parse + full circuit**, not a single gate.

### 2.6 PowerShell: `scripts/RQ4D_World_Record.ps1`

**Universal paths** — resolves engine root from the script’s location (`scripts/` → parent = repo root).

| Parameter | Role |
|-----------|------|
| `-QubitCount` | Default `65536`; max `131072`. |
| `-EngineRoot` | Override repo root. |
| `-OutScriptName` | Generated file name under `examples/`. |
| `-MirrorDir` | Optional copy of script + `.rq4d` elsewhere. |
| `-GenerateOnly` | Write `.rq4d` only; do not run. |
| `-SkipBuild` | Assume binary already built. |

---

## 3. Programming guide

### 3.1 Repository layout

```text
cmd/rq4d/main.go          # CLI, instruction dispatch, parallel batches, telemetry
internal/math/clifford.go # Cl(4,0): Multivector, GeometricProduct, Rotor, Normalize, Reverse
internal/quantum/bridge.go# Qubit encoding, Hadamard, Pauli-X variants, CNOT, Measure
internal/parser/lexer.go # .rq4d line parser → []Instruction
examples/*.rq4d           # Sample circuits (and some Roma4D-style .r4d companions)
scripts/*.ps1             # Automation / large manifold generation
go.mod                    # module romanailabs/rq4d
```

### 3.2 Geometric model (non-negotiable design choices)

- **Algebra**: **Cl(4,0)** — Euclidean signature `(+,+,+,+)` on four basis vectors **e₁…e₄**.  
- **Storage**: `Multivector` holds **16** `float64` coefficients; basis blade index = **4-bit bitmask** (bit `i` ⟺ **e_{i+1}** present).  
- **No complex matrices**, no Hilbert-space tensor product for multi-qubit entanglement in the current **product-state** simulator. Each “qubit” is an independent **16D multivector**.  
- **Geometric product** is the full Clifford product on basis blades (see `basisProduct` + double loop in `GeometricProduct`).

### 3.3 Qubit encoding

| State | Multivector |
|-------|-------------|
| **\|0⟩** | Scalar blade **1.0** (index 0), rest 0. |
| **\|1⟩** | **e₁** blade (index 1) normalized usage via gates. |
| **Hadamard** | Left multiply by operator with scalar and **e₁** each **1/√2** (superposition in this real slice). |
| **X (script / CNOT)** | **PauliXBitFlip**: left multiply by **e₁** (swap **\|0⟩↔|1⟩** in this encoding). |

### 3.4 CNOT semantics (current engine)

- **Control “is \|1⟩”** only if **P(\|1⟩) = e₁² / ‖M‖² > 1 − ε** (`ctrlOneEps` in `bridge.go`).  
- **\|+⟩**-style controls (**≈50/50**) do **not** flip the target (avoids bogus classical randomness).  
- True **entanglement** (correlated multi-qubit state in one GA object) is **not** implemented; CNOT is a **conditional classical flip** on the product representation.

### 3.5 Parallelism

- **ALLOC**: goroutine per slot for `QubitZero()`.  
- **Runs of `H` or `X`**: one `WaitGroup` wave per contiguous block of the same opcode.  
- **MEASURE**: goroutine per qubit for probability extraction, then ordered print.  
- Extremely large **H** blocks ⇒ **O(n) goroutines** in one wave — may limit practical **n** on client hardware.

### 3.6 Extending the engine

1. **New opcode**  
   - Add `Op*` constant and `parseLine` branch in `internal/parser/lexer.go`.  
   - Extend `Instruction` fields if needed.  
   - Handle in `cmd/rq4d/main.go` (consider batching like `H`/`X` if embarrassingly parallel).

2. **New gate**  
   - Implement in `internal/quantum/bridge.go` using **`gamath.GeometricProduct`** / **`Rotor`**.  
   - Keep **`romanailabs/rq4d/internal/math`** import alias **`gamath`** to avoid clashing with **`math`**.

3. **Joint quantum state**  
   - Would require a **new representation** (e.g. dedicated tensor or GA in higher dimension), **not** a drop-in patch to the current per-qubit `[]Multivector`.

### 3.7 Copyright / file headers

Go sources in this project use:

```text
// <filename>.go
// Copyright RomanAILabs - Daniel Harding
```

Preserve this when adding files.

---

## 4. LLM briefing pack (advanced)

Use this section as **system or user context** when asking an LLM to modify RQ4D.

### 4.1 One-paragraph project truth

> **RomaQuantum4D (RQ4D)** is a **Go 1.22** module **`romanailabs/rq4d`** that executes line-oriented **`.rq4d` scripts** with opcodes **ALLOC, H, X, CNOT, MEASURE**. Quantum-style behavior is simulated with **Cl(4,0) multivectors** (16 `float64`s per qubit) and the **geometric product** — **not** with complex unitary matrices or density matrices. The CLI is **`cmd/rq4d`**. Multi-qubit **entanglement is not** modeled; CNOT is a **conditional bit-flip** when the control is **definitely \|1⟩** in the scalar/**e₁** encoding.

### 4.2 Hard rules for generated changes

1. **Do not** replace the GA core with `complex128` matrices or NumPy-style statevectors unless the user explicitly requests a **new subsystem** and accepts a **breaking redesign**.  
2. **Do not** import standard **`math`** as **`math`** in `internal/quantum` if it shadows **`romanailabs/rq4d/internal/math`** — use **`gamath`** / **`stdmath`**.  
3. **Preserve** telemetry line format: **`[Telemetry: Z-Axis = %.2fms]`** unless the user asks to change it.  
4. **Parser** must remain **line-based**, **fail with clear errors** (file:line), and **ignore `#`**.  
5. **Batch parallel `H` / `X`** for consecutive lines — do not serialize those without reason.  
6. **`.r4d` Roma4D language** files in `examples/` are **not** parsed by `rq4d`; do not assume they load in the Go engine.

### 4.3 File map (where to edit what)

| Task | Primary files |
|------|----------------|
| New script opcode | `internal/parser/lexer.go`, `cmd/rq4d/main.go` |
| Gate / measure math | `internal/quantum/bridge.go` |
| GA product / rotor | `internal/math/clifford.go` |
| UX / banner / flags | `cmd/rq4d/main.go` |
| Large demo generation | `scripts/RQ4D_World_Record.ps1` |
| User-facing docs | `README.md`, **`docs/RQ4D_MASTER_GUIDE.md`** |

### 4.4 Suggested prompts (copy-paste)

**Refactor**

> In `romanailabs/rq4d`, extract instruction execution from `main` into an `internal/engine` package without changing observable output or the `.rq4d` ISA.

**Opcode**

> Add opcode `Y [target]` that applies a documented Cl(4,0) operator (left multiply) to qubit `target`, batch consecutive `Y` lines like `H`, and update `docs/RQ4D_MASTER_GUIDE.md`.

**Performance**

> Reduce goroutine fan-out for large `H` blocks by processing in chunks of 1024 with a worker pool; keep results identical to the current engine within float tolerance.

**Tests**

> Add `internal/math` table-driven tests for `GeometricProduct` on basis blades and golden values for `e12*e12 == -1` in the scalar blade.

### 4.5 Hallucination guardrails

- There is **no** built-in GPU, **no** Qiskit, **no** automatic Roma4D compiler bridge in this repo.  
- **Repository URL**: **https://github.com/RomanAILabs-Auth/RomaQuantum4D**  
- Module path: **`romanailabs/rq4d`**

---

## 5. Quick reference

### 5.1 ISA (RQ4D `.rq4d` engine)

```text
ALLOC n
H i
X i
CNOT c t
MEASURE
```

### 5.2 Commands

```bash
go run ./cmd/rq4d examples/manifold_sweep.rq4d
go build -o rq4d ./cmd/rq4d && ./rq4d examples/cnot_demo.rq4d
```

```powershell
pwsh -File .\scripts\RQ4D_World_Record.ps1 -QubitCount 8192
```

### 5.3 Links

- **GitHub**: [RomanAILabs-Auth/RomaQuantum4D](https://github.com/RomanAILabs-Auth/RomaQuantum4D)  
- **Go**: [https://go.dev/doc/](https://go.dev/doc/)

---

*End of RQ4D Master Guide — RomanAILabs / Daniel Harding.*
