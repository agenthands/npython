# Implementation Plan: Phase 3 - Security & The Standard Library

## Phase 1: VM Security Logic
- [x] Task: Implement `ScopeStack` and `Gatekeeper` in `pkg/vm/machine.go`.
- [x] Task: Implement `OP_ADDRESS` and `OP_EXIT_ADDR` in `vm.Run()`.
- [x] Task: Implement the dynamic `HostFunctionRegistry` in `pkg/vm/machine.go`.
- [x] Task: Implement `OP_SYSCALL` in `vm.Run()` with scope validation.

## Phase 2: Standard Library Sandboxes
- [x] Task: Implement `pkg/stdlib/fs.go` (FS-ENV sandbox).
    - [x] Implement Root Jailing, Read/Write split, and Size Limits.
- [x] Task: Implement `pkg/stdlib/http.go` (HTTP-ENV sandbox).
    - [x] Implement Domain Allowlist and Localhost blocking.

## Phase 3: Integration & Verification
- [x] Task: Write `tests/security_test.go` to verify capability gating and sandbox violations.
- [x] Task: Conductor - User Manual Verification 'Phase 3: Security & Stdlib' (Protocol in workflow.md).
