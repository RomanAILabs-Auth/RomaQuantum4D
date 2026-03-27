# RQ4D

**RQ4D HyperEngine** — Go reference engine for quantum-style circuits expressed in **Cl(4,0)** geometric algebra (16-dimensional multivectors), without complex matrices.

## Run

```bash
go run ./cmd/rq4d examples/cnot_demo.rq4d
go run ./cmd/rq4d examples/parallel_h.rq4d
```

## `.rq4d` script format

| Instruction | Meaning |
|-------------|---------|
| `ALLOC n` | n qubits in $\|0\rangle$ |
| `H i` | Hadamard on qubit `i` (consecutive `H` lines run in parallel) |
| `X i` | Pauli-X (bit flip) on `i` |
| `CNOT c t` | Conditional X on `t` when control `c` is $\|1\rangle$ (product-state model) |
| `MEASURE` | Print $P(\|0\rangle)$, $P(\|1\rangle)$ per qubit |

Lines starting with `#` are comments.

## Module

`romanailabs/rq4d`

## Copyright

RomanAILabs — Daniel Harding
