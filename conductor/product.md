# Product Definition: nForth Engine

## Executive Summary
The nForth Engine is a secure, zero-allocation Bytecode Virtual Machine designed to execute the Natural FORTH language. It acts as the actuation layer for AI Agents, replacing "ambient authority" interpreters with a constrained, capability-gated execution environment.

## Architectural Invariants
1. **Host-Guest Isolation:** Go Runtime (Host) controls the VM (Guest). Guest has zero OS access without explicit Capability Tokens.
2. **Zero-Allocation Hot Path:** The main execution loop (`vm.Run`) must generate zero garbage.
3. **Canonical State Enforcement (State Grounding):** Values must be consumed into local variables immediately via the `INTO` keyword (EBNF enforced).

## Language Grammar (v1.0)
- **Named State:** Every expression yielding a value must end with `INTO` or `->`.
- **Explicit Scope:** Locals declared upfront in function definitions.
- **Capability Gating:** `ADDRESS` blocks required for privileged operations.

## Core Modules
- **pkg/core/value**: 16-byte tagged union data representation.
- **pkg/vm**: Fixed-size stack/frame execution engine.
- **pkg/compiler**: EBNF-compliant parser with "Dangling Stack" validation.
- **pkg/security**: Capability-based access control for host function syscalls.
