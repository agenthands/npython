# Implementation Plan: Phase 2 - Compiler & Safety Pipeline

## Phase 1: Zero-Allocation Lexer
- [x] Task: Implement `pkg/compiler/lexer/token.go` (Token struct and Kind enum).
- [x] Task: Implement `pkg/compiler/lexer/scanner.go` (Cursor-based Lexer).
    - [x] Implement `skipWhitespace` and `skipComment`.
    - [x] Implement `scanGateSugar` for `<ENV-GATE>`.
    - [x] Implement identifier scanning with "Noise Word" and keyword mapping (`INTO`, `->`).
- [x] Task: Write `pkg/compiler/lexer/scanner_test.go` to verify zero-alloc and correctness.

## Phase 2: EBNF Parser & AST
- [x] Task: Define AST nodes in `pkg/compiler/ast/ast.go`.
- [x] Task: Implement the recursive-descent parser in `pkg/compiler/parser/parser.go`.
    - [x] Support `definition`, `assignment`, `void_operation`, and `control_flow`.

## Phase 3: The "INTO" Enforcer & Scope Validator
- [x] Task: Implement `pkg/compiler/parser/validator.go` (Stack depth tracking).
    - [x] Define `StandardWords` map with stack signatures.
    - [x] Implement `ValidateState` check after every statement.
- [x] Task: Implement Scope Tracking for `ADDRESS` blocks.
    - [x] Implement capability check against active scope stack.

## Phase 4: Bytecode Emitter
- [x] Task: Implement `pkg/compiler/emitter/emitter.go`.
    - [x] Convert AST nodes to `uint32` instructions.
    - [x] Build the Constant Pool (`value.Value`).

## Phase 5: Integration & Verification
- [x] Task: Write `tests/compiler_test.go` covering "Dangling Stack" and "Security Violation" scenarios.
- [x] Task: Conductor - User Manual Verification 'Phase 2: Compiler Safety' (Protocol in workflow.md).
