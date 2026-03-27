# RQ4D (RomaQuantum4D)

**RQ4D** — Go engine for **geometric simulation scale** quantum-style circuits in **Cl(4,0)** (16-component multivectors per lane), no complex matrices.

**Repository:** [github.com/RomanAILabs-Auth/RomaQuantum4D](https://github.com/RomanAILabs-Auth/RomaQuantum4D)

**Full documentation (install + user + programming + LLM pack):** [docs/RQ4D_MASTER_GUIDE.md](docs/RQ4D_MASTER_GUIDE.md)

## World-record style demo (PowerShell, any clone path)

From the repo root (or invoke by full path):

```powershell
pwsh -ExecutionPolicy Bypass -File .\scripts\RQ4D_World_Record.ps1
pwsh -File .\scripts\RQ4D_World_Record.ps1 -QubitCount 65536
pwsh -File .\scripts\RQ4D_World_Record.ps1 -QubitCount 131072 -GenerateOnly
```

Optional: `-MirrorDir "D:\Backups\RomanAILabs"` copies the generated `.rq4d` and this script there.  
`-EngineRoot` overrides auto-detected repo root. `-SkipBuild` uses an existing `rq4d` / `rq4d.exe` binary.

## Featured demo (8-qubit manifold sweep)

Parallel Hadamard on eight lanes, then measurement (50/50 superposition per qubit):

```bash
go run ./cmd/rq4d examples/manifold_sweep.rq4d
go run ./cmd/rq4d --truth-mode examples/manifold_sweep.rq4d
```

Expected: banner `Executing RQ4D (geometric simulation scale...)`, eight `MEASURE q[i]` lines, then the **honest telemetry** block (gate op count, global-pass timing, bytes touched, **FNV-1a checksum**).

Optional **`--truth-mode`**: disables parallel **H**/**X** batching and runs a full **O(n) global pass** after every gate line (stricter, slower).

## Other examples

```bash
go run ./cmd/rq4d examples/cnot_demo.rq4d
go run ./cmd/rq4d examples/parallel_h.rq4d
```

## Roma4D companion

`examples/spacetime_ui_v3.r4d` is a **native Roma4D** worldtube / `par for` UI sketch — run with **`r4d`** from a Roma4D checkout, not `rq4d`.

## `.rq4d` script ISA (RQ4D engine)

| Instruction | Meaning |
|-------------|---------|
| `ALLOC n` | n qubits in $\|0\rangle$ |
| `H i` | Hadamard on qubit `i` (consecutive `H` lines batch in parallel unless `--truth-mode`) |
| `X i` | Pauli-X (bit flip) on `i` |
| `CNOT c t` | Conditional X on `t` when control `c` is $\|1\rangle$ |
| `MEASURE` | Print $P(\|0\rangle)$, $P(\|1\rangle)$ per qubit |

Lines starting with `#` are comments.

## Module

`romanailabs/rq4d`

## Copyright

RomanAILabs — Daniel Harding
