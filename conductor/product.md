# Product Definition: nPython Engine

## Executive Summary
The nPython Engine is a secure, zero-allocation Bytecode Virtual Machine designed to execute a subset of Python. It acts as the actuation layer for AI Agents, replacing "ambient authority" interpreters with a constrained, capability-gated execution environment.

## Architectural Invariants
1. **Host-Guest Isolation:** Go Runtime (Host) controls the VM (Guest). Guest has zero OS access without explicit Capability Tokens.
2. **Zero-Allocation Hot Path:** The main execution loop (`vm.Run`) must generate zero garbage.
3. **Capability Gating:** All privileged operations (I/O, Network) require an active security scope.

## Core Modules
- **pkg/vm**: Fixed-size stack/frame execution engine.
- **pkg/compiler/python**: Python AST to nPython Bytecode translator.
- **pkg/security**: Capability-based access control.
- **pkg/stdlib**: Sandboxed syscalls.
