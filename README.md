# Natural FORTH (nFORTH) Engine

## The Problem: Architectural Impedance
General-purpose languages (Python, TypeScript, JavaScript) were built for human cognition, relying on visual formatting and ambient authority. For AI agents, these introduce fatal structural flaws:
1. **Hallucination via Ambiguity:** Languages with massive syntax flexibility cause models to lose track of state, leading to fabricated variables, "stack drift," and catastrophic logic errors.
2. **The Confused Deputy Attack:** Running agents in containers with broad tool access (like `bash` or `python` REPLs) means a single prompt injection can compromise the entire underlying system.

##  The nFORTH Solution
nFORTH replaces chaotic "ambient authority" and verbose syntax with a mathematically sound, LLM-native architecture:

1. **Concatenative Token Efficiency:** Strips out brackets, commas, and formatting. Saves 15-30% of the LLM's context window, allowing for deeper reasoning traces and lower inference costs.
2. **Explicit State Constraints (The `INTO` Rule):** Agents cannot leave data floating in memory. Every data transformation must explicitly name its output state (e.g., `FILTER data INTO clean_data`), providing constant-time semantic anchors and reducing logic hallucinations to near-zero.
3. **The `ADDRESS` Security Gateway:** Agents operate with **zero** ambient authority. To access tools (HTTP, Filesystem, SQL), the agent must hold an unforgeable cryptographic Capability Token and explicitly switch execution contexts.

## Architecture & Engineering

nFORTH operates a strictly sandboxed Virtual Machine. The core execution loop is designed for massive horizontal scaling of agentic threads. It uses zero-allocation patterns (Go's sync.Pool and Tagged Unions) to prevent Garbage Collection latency during high-concurrency workflows.

    pkg/vm: The core concatenative execution engine and explicit state tracker.

    pkg/security: The ADDRESS Gateway that dynamically swaps Sealed Wordlists based on unforgeable cryptographic capability tokens.

    pkg/compiler: Enforces canonical sentence structures and static INTO checks before runtime.

🤝 Contributing & TDD Workflow

We operate under a strict Zero-Trust TDD mandate. 👉 Read GEMINI.MD before writing a single line of code. Pull requests without 100% test coverage for new logic will be automatically rejected by CI.
