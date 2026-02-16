# nForth Engine

## The Problem: Architectural Impedance
General-purpose languages (Python, TypeScript, JavaScript) were built for human cognition, relying on visual formatting and ambient authority. For AI agents, these introduce fatal structural flaws:
1. **Hallucination via Ambiguity:** Languages with massive syntax flexibility cause models to lose track of state, leading to fabricated variables, "stack drift," and catastrophic logic errors.
2. **The Confused Deputy Attack:** Running agents in containers with broad tool access means a single prompt injection can compromise the entire underlying system.

## The nForth Solution
nForth replaces chaotic "ambient authority" and verbose syntax with a mathematically sound, LLM-native architecture:

1. **Concatenative Token Efficiency:** Strips out brackets, commas, and formatting. Saves 15-30% of the LLM's context window, allowing for deeper reasoning traces and lower inference costs.
2. **Explicit State Constraints (The `INTO` Rule):** Agents cannot leave data floating in memory. Every data transformation must explicitly name its output state (e.g., `10 20 ADD INTO sum`), providing constant-time semantic anchors and reducing logic hallucinations.
3. **Capability Security:** Agents operate with **zero** ambient authority. To access tools (HTTP, Filesystem), the agent must hold a valid Capability Token and explicitly switch scopes via `ADDRESS`.

## Quick Start

### 1. Build the Engine
```bash
go build -o nforth ./cmd/nforth
```

### 2. Run an nForth Script
```bash
./nforth run script.nf
```

### 3. Run Authoritative E2E Tests
```bash
go test -v ./tests/main_test.go
```

## Architecture & Engineering

nForth operates a strictly sandboxed Virtual Machine. The core execution loop is designed for high-performance, zero-allocation horizontal scaling.

*   **pkg/vm**: Core concatenative execution engine with fixed-size stack and frames.
*   **pkg/compiler**: EBNF-compliant parser that enforces the `INTO` rule and validates security scopes at compile-time.
*   **pkg/security**: Capability-based access control for standard library syscalls.
*   **pkg/stdlib**: Sandboxed Filesystem and HTTP clients.

🤝 Contributing & TDD Workflow

We operate under a strict TDD mandate. All features must be verified via the `tests/main_test.go` integration suite. Pull requests without matching tests will be automatically rejected.
