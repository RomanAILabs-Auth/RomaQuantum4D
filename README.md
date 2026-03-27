# RQ4D (RomaQuantum4D)

**RQ4D HyperEngine** — Go engine for quantum-style circuits in **Cl(4,0)** geometric algebra (16-component multivectors), no complex matrices.

**Repository:** [github.com/RomanAILabs-Auth/RomaQuantum4D](https://github.com/RomanAILabs-Auth/RomaQuantum4D)

## Featured demo (8-qubit manifold sweep)

Parallel Hadamard on eight lanes, then measurement (50/50 superposition per qubit):

```bash
go run ./cmd/rq4d examples/manifold_sweep.rq4d
```

Expected: green `Executing RQ4D HyperEngine...`, eight `MEASURE q[i]` lines, then `[Telemetry: Z-Axis = X.XXms]`.

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
| `H i` | Hadamard on qubit `i` (consecutive `H` lines batch in parallel) |
| `X i` | Pauli-X (bit flip) on `i` |
| `CNOT c t` | Conditional X on `t` when control `c` is $\|1\rangle$ |
| `MEASURE` | Print $P(\|0\rangle)$, $P(\|1\rangle)$ per qubit |

Lines starting with `#` are comments.

## Module

`romanailabs/rq4d`

## Copyright

RomanAILabs — Daniel Harding
