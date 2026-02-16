# Implementation Plan: Phase 1 - VM Core Implementation

## Phase 1: Core Data Structures
- [x] Task: Implement `pkg/core/value/value.go` (Value type, Tagged Union).
    - [x] Define `Type` enum (Int, Bool, String).
    - [x] Implement `Value` struct.
- [x] Task: Implement `pkg/vm/machine.go` (Machine and Frame structs).
    - [x] Define constants: `StackDepth=128`, `MaxFrames=32`, `MaxLocals=16`.
    - [x] Implement `Machine.Reset()`.

## Phase 2: Stack & Memory Logic
- [x] Task: Implement panic-based stack helpers in `pkg/vm/machine.go` (internal `push`/`pop`).
- [x] Task: Implement Byte Buffer Arena packing in `pkg/core/value/value.go`.
    - [x] `PackString(offset, length)` and `UnpackString(data, arena)`.

## Phase 3: The Execution Loop
- [x] Task: Define Opcode constants in `pkg/vm/ops.go`.
- [x] Task: Implement `vm.Run()` in `pkg/vm/machine.go`.
    - [x] Add `defer recover()` safety net.
    - [x] Implement `switch` statement for `OP_HALT`, `OP_PUSH_C`, `OP_ADD`, `OP_POP_L`.
    - [x] Implement local variable caching for `IP`, `SP`, `FP`.

## Phase 4: Verification & Benchmarking
- [x] Task: Write unit tests in `tests/vm_test.go` for core opcodes.
- [x] Task: Write benchmarks in `tests/vm_bench_test.go` to verify zero-allocation and throughput.
- [x] Task: Conductor - User Manual Verification 'Phase 1: VM Core' (Protocol in workflow.md).
