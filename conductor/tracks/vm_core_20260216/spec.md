# Specification: Phase 1 - VM Core Implementation

## Overview
Implement the foundational data structures and the high-performance execution loop for the nPython Virtual Machine. This phase establishes the "Zero-Allocation" hot path and the "Panic-based" safety model.

## Functional Requirements
- **Data Representation**: Implement `value.Value` as a 16-byte tagged union (Type + 64-bit Data).
- **Machine State**: Implement `vm.Machine` with fixed-size arrays:
    - `Stack`: `[128]Value`
    - `Frames`: `[32]Frame` (with `[16]Value` for locals).
- **Execution Engine**: Implement `vm.Run()` using a **Big Switch** dispatch mechanism.
- **Stack Operations**: Implement internal `push` and `pop` using **Panic-based** bounds checking, with a `recover` safety net in `Run()`.
- **String Handling**: Implement a **Byte Buffer Arena** in the `Machine`. Strings on the stack are packed `(offset, length)` integers.
- **Opcodes**: Implement initial core opcodes:
    - `OP_HALT` (0x00)
    - `OP_PUSH_C` (0x02) - Push Constant
    - `OP_POP_L` (0x04) - INTO (Pop to Local)
    - `OP_ADD` (0x10) - Integer Addition

## Non-Functional Requirements
- **Zero Allocations**: The `vm.Run` loop must produce 0 heap allocations for pure arithmetic/stack operations.
- **Performance**: Throughput must exceed 1,000,000 ops/sec.
- **Safety**: Stack overflows/underflows must be caught and returned as `error`, not crash the host process.

## Acceptance Criteria
- [ ] `Machine` can execute a manually assembled sequence: `PUSH_C(1) PUSH_C(2) ADD POP_L(0) HALT`.
- [ ] `Machine.Reset()` correctly zeroes out state for pool reuse.
- [ ] 100% test coverage for `pkg/vm` and `pkg/core/value`.
- [ ] Benchmarks confirm 0 allocs/op for the execution loop.
