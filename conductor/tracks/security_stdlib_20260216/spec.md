# Specification: Phase 3 - Security & The Standard Library

## Overview
Implement the security gateway and sandboxed standard libraries for the nPython engine. This phase enables safe interaction with the host system through capability-gated syscalls.

## Functional Requirements
- **OP_ADDRESS & OP_EXIT_ADDR VM Logic**: 
    - Implement a scope stack in `vm.Machine`.
    - `OP_ADDRESS` validates capability tokens via a host-provided gatekeeper and pushes to the stack.
    - `OP_EXIT_ADDR` pops the current scope (LIFO).
- **Dynamic Host Function Registry**:
    - Allow the host to register Go functions (`func(*Machine) error`) at runtime.
    - Use O(1) index-based lookup for syscalls.
- **OP_SYSCALL Implementation**:
    - Validate that the required scope for a syscall is present in the cumulative scope hierarchy.
- **FS-ENV Sandbox**:
    - Root jailing (relative to workspace).
    - Separate `fs:read` and `fs:write` capabilities.
    - 5MB file size limit for writes.
- **HTTP-ENV Sandbox**:
    - Strict domain allowlist (provided in token).
    - Block localhost and internal IP ranges.

## Acceptance Criteria
- [ ] `OP_ADDRESS` correctly manages the scope stack in the VM.
- [ ] `OP_SYSCALL` fails if the required scope is missing.
- [ ] `FS-ENV` rejects paths that attempt to traverse outside the root.
- [ ] `HTTP-ENV` only allows requests to authorized domains.
- [ ] Integration: A complete agent script can `FETCH` from an authorized URL and `WRITE` to a local file.
