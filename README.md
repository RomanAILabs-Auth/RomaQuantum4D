# RQ4D (RomaQuantum4D)

**RQ4D** — Go **quantum lattice simulator**: complex amplitudes on a 3D grid, Trotter-style steps, optional **tensor-network–style** bond truncation (`--backend=tn`, `--chi`). Mean-field and TN paths are deterministic aside from explicit measurement sampling.

**Repository:** [github.com/RomanAILabs-Auth/RomaQuantum4D](https://github.com/RomanAILabs-Auth/RomaQuantum4D)

**Documentation:** [docs/RQ4D_MASTER_GUIDE.md](docs/RQ4D_MASTER_GUIDE.md) (may still describe the legacy geometric CLI in places; the binary below is the lattice engine.)

## Build, run, install

From the repository root (module `github.com/RomanAILabs-Auth/RomaQuantum4D`):

```bash
go build -o rq4d ./cmd/rq4d          # Unix/macOS
go build -o rq4d.exe ./cmd/rq4d      # Windows
go install ./cmd/rq4d                # installs rq4d to $GOPATH/bin or $GOBIN
```

### Lattice run (default)

```bash
go run ./cmd/rq4d
# or with options:
go run ./cmd/rq4d -lx 8 -ly 8 -lz 8 -steps 30 -backend tn -chi 4 -measure -seed 7
```

Flags include `-lx`, `-ly`, `-lz`, `-dim` (2, 4, or 8), `-dt`, `-steps`, `-j`, `-hz`, `-hx`, `-workers`, `-backend` (`meanfield` | `tn` | `cpu`), `-chi` (1…32 for `tn`), `-measure`, `-collapse`, `-seed`.

### Legacy `.rq4d` script examples

Files under `examples/*.rq4d` targeted the **previous** Cl(4,0) geometric script runner. The current `rq4d` binary does **not** parse those scripts; keep them as reference or remove locally.

### Large-scale demo script

`scripts/RQ4D_World_Record.ps1` was written for the legacy script-based CLI. It is **not** wired to the lattice flags yet; use `go run ./cmd/rq4d` with explicit `-lx/-ly/-lz` for large sweeps (mind memory).

## Roma4D companion

`examples/spacetime_ui_v3.r4d` is **Roma4D** source — run with **`r4d`** from a Roma4D checkout, not `rq4d`.

## Module

`github.com/RomanAILabs-Auth/RomaQuantum4D`

## Copyright

RomanAILabs — Daniel Harding
