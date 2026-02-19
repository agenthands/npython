# nPython Engine: Comparative Benchmarks

This document measures the "nPython Advantage" against Python across three primary dimensions: **Token Efficiency**, **Security surface Area**, and **Execution Reliability**.

---

## 1. Token Efficiency (Density)

We compared equivalent code for common agentic tasks. 

| Task | nPython (Tokens) | Python (Tokens) | Density Advantage |
| :--- | :--- | :--- | :--- |
| **Arithmetic & Tax** | 18 | 12 | -50% (Python is denser for pure math) |
| **Secure HTTP Fetch** | 20 | 17 | -17.6% |
| **Secure File Write** | 13 | 21 | **+38.1% (nPython Advantage)** |

**Analysis:** nPython becomes significantly more token-efficient as task complexity and security requirements increase. The lack of imports, try-except blocks, and context managers in nPython saves significant context window for the LLM.

---

## 2. Security Surface Area

We measured the number of **Ambient Authorities** granted to the script.

| Metric | nPython | Python |
| :--- | :--- | :--- |
| **Default Authority** | Zero (Sterile Sandbox) | Global (Full OS Access) |
| **I/O Access** | Explicit (Gatekeeper tokens) | Ambient (Direct syscalls) |
| **Path Traversal** | Blocked (Root Jail) | Possible (Manual validation) |

**The nPython Advantage:** nPython implements "Least Privilege" by default. An agent cannot even *see* the network or filesystem without an explicit `ADDRESS` gate.

---

## 3. Execution Reliability (Anti-Hallucination)

| Hallucination Type | nPython Enforcement | Python Behavior |
| :--- | :--- | :--- |
| **Variable Drift** | **Blocked** (SyntacticHallucinationError) | Allowed (Runtime logic errors) |
| **Implicit State** | **Forbidden** (INTO rule) | Common (Unused return values) |
| **Tool Misuse** | **Prevented** (Scoped dictionary) | Likely (Hallucinated imports) |

**Conclusion:** While Python excels at mathematical conciseness, nPython is architecturally superior for **Autonomous Agent Orchestration** where security, reliability, and token density for complex logic are the priority.
