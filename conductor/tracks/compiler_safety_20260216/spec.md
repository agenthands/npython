# Specification: Phase 2 - Compiler & Safety Pipeline

## Overview
Implement a high-performance, zero-allocation compiler for Natural FORTH (nFORTH). The compiler acts as the primary safety gate, enforcing "State Grounding" (the INTO rule) and "Scoped Authority" (capability-based access) at compile-time to prevent LLM hallucinations and security violations.

## Functional Requirements
- **Zero-Allocation Lexer**: 
    - Cursor-based scanner operating on `[]byte` source.
    - Support for `<ENV-GATE>` sugar (TokenSugarGate).
    - Support for `->` as `TokenInto`.
    - Support for "Noise Words" (e.g., `THE`, `WITH`, `USING`) which are identified but skipped by the parser.
    - Support for single-line comments (`\ `).
- **EBNF Parser & Validator**:
    - Implement the AST based on the nFORTH v1.0 EBNF.
    - **The "INTO" Enforcer**: Track `VirtualStackDepth`. Every statement must result in `depth == 0`.
    - Raise `ErrFloatingState` with contextual hints on any stack drift.
- **Cumulative Scope Validation**:
    - Maintain a scope stack for `ADDRESS` blocks.
    - Validate that restricted words (e.g., `FETCH`) are only used within their required capability scopes.
- **Bytecode Emitter**:
    - Convert the validated AST into a `[]uint32` instruction stream and a Constant Pool.

## Non-Functional Requirements
- **Zero Allocations (Lexer)**: The `Lexer.Next()` call must produce 0 heap allocations.
- **Path-Dependent Truth**: Stop parsing and halt on the first validation error to prevent desynchronized error cascades.

## Acceptance Criteria
- [ ] Lexer correctly identifies `<HTTP-GATE>` and `->` with zero allocations.
- [ ] Parser rejects `1 2 ADD` with a "Dangling Stack" error.
- [ ] Parser accepts `1 2 ADD INTO result`.
- [ ] Parser rejects `FETCH` outside of an `ADDRESS HTTP-ENV` block.
- [ ] Integration: Source code successfully compiles to bytecode executable by Phase 1 VM.
