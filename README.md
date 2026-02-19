# nPython Engine

## The Problem: Architectural Impedance
General-purpose Python runtimes were built for human cognition, relying on visual formatting and ambient authority. For AI agents, these introduce fatal structural flaws:
1. **The Confused Deputy Attack:** Running agents in standard Python environments with broad tool access means a single prompt injection can compromise the entire underlying system.
2. **Resource Inefficiency:** Standard Python is heavy and hard to sandbox at the instruction level.

## The nPython Solution
nPython provides a secure, zero-allocation Python-native architecture:

1. **Instruction-Level Sandboxing:** Instead of OS-level containers, nPython executes Python code on a custom VM with zero ambient authority.
2. **Capability Security:** Agents operate with **zero** ambient authority. To access tools (HTTP, Filesystem), the agent must hold a valid Capability Token and explicitly use `with scope(name, token):` blocks.
3. **High-Performance Core:** A zero-allocation Go-based VM ensures low-latency execution and horizontal scaling for agent swarms.

## Quick Start

### 1. Build the Engine
```bash
go build -o npython ./cmd/npython
```

### 2. Run a Python Script
```bash
./npython run script.py
```

### 3. Run Authoritative E2E Tests
```bash
go test -v ./tests/python_compiler_test.go
```

## Documentation

- [Python Usage Guide](./docs/USAGE.md): Learn the supported Python subset.
- [System Prompt](./docs/PROMPT.md): The master prompt for LLM Python generation.
- [Datacard](./docs/DATACARD.md): High-density reference card for LLMs.
- [Embedding Guide](./docs/EMBEDDING.md): How to integrate the VM into your Go app.

## Architecture & Engineering

nPython operates a strictly sandboxed Virtual Machine. The core execution loop is designed for high-performance, zero-allocation horizontal scaling.

*   **pkg/vm**: Core execution engine with fixed-size stack and frames.
*   **pkg/compiler/python**: Python-to-Bytecode compiler that enforces security scopes.
*   **pkg/stdlib**: Sandboxed Filesystem and HTTP clients.

🤝 Contributing & TDD Workflow

We operate under a strict TDD mandate. All features must be verified via the `tests/main_test.go` integration suite. Pull requests without matching tests will be automatically rejected.
