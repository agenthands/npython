# nPython Engine: AI-Native Execution Context

This project is a secure, high-performance Python-native Virtual Machine (VM) implemented in Go. It is specifically designed for AI agents to execute Python code safely with instruction-level sandboxing and capability-based security.

## Project Overview

- **Core Technology:** Go (Golang) 1.25.1+
- **Architecture:** Stack-based Virtual Machine with a custom bytecode instruction set.
- **Security Model:** Zero Ambient Authority. External resources (Filesystem, Network) require explicit capability tokens via `with scope(name, token):` blocks.
- **Parser:** Uses `github.com/go-python/gpython` for Python AST parsing.

### Key Directories

- `cmd/npython/`: The main CLI entry point. Supports `run` and `query` commands.
- `pkg/vm/`: The core execution engine, machine state, and bytecode operations.
- `pkg/compiler/python/`: Python-to-Bytecode compiler.
- `pkg/stdlib/`: Implementation of Python built-ins and sandboxed system calls.
- `conductor/`: Project management, workflow definitions, and development tracks.
- `tests/`: Integration and end-to-end tests for the compiler and VM.

## Building and Running

### Build the CLI
```bash
go build -o npython ./cmd/npython
```

### Run a Python Script
```bash
./npython run examples/math.py
```

### Run with Gas Limit
```bash
./npython run script.py -gas 1000000
```

## Development Conventions

### Strict TDD Mandate
All new features and bug fixes **MUST** follow the 7-step TDD cycle as defined in `conductor/workflow.md`:
1. Write a failing test (RED).
2. Confirm the failure.
3. Write minimal implementation.
4. Confirm the test passes (GREEN).
5. Refactor.
6. Run all tests to ensure no regressions.
7. Commit.

### Verification Protocol
Before completing a task or phase, ensure:
- [ ] All tests are passing: `go test ./...`
- [ ] Linter/Type-checker passes: `go vet ./...`
- [ ] Code follows existing zero-allocation patterns in `pkg/vm`.

### Testing Commands
- **Run all tests:** `go test ./...`
- **Run compiler tests:** `go test -v ./pkg/compiler/python/...`
- **Run VM integration tests:** `go test -v ./tests/python_e2e_test.go`

## Security Guidelines

- **No Ambient Authority:** Never implement host functions that access global state or the OS without going through the `Gatekeeper` and `scope` mechanism.
- **Instruction Level Sandboxing:** Every operation must be checked against the machine's constraints (e.g., gas limits, stack depth).
- **Capability Tokens:** Tools like `fetch` and `write_file` must be protected by scope-specific tokens.
