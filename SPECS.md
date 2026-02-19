# nPython Engine: Technical Specification
**Version:** 1.0.0-ReleaseCandidate
**Architecture:** Zero-Allocation Python VM with Capability-Based Security

---

## 1. System Architecture
The system follows a **Host-Guest Isolation Model**.
* **Host (Go Runtime):** Manages memory, I/O, and Capability Tokens.
* **Guest (nPython VM):** A pure logic engine that executes bytecode compiled from Python. It has **zero** direct access to the OS.
* **Python-Native Design:** The LLM writes standard Python. The compiler transforms this into secure bytecode.

---

## 2. Virtual Machine (The Core)

### 2.1 Performance Invariants
* **Zero-Allocation:** The `Run()` loop MUST NOT allocate heap memory.
* **Panic-Based Error Handling:** Internal stack overflows/underflows trigger a `panic`, returning safe `error` values.
* **Dispatch:** Big Switch main loop.

### 2.2 Memory Layout
* **Stack:** Fixed-size `[128]Value` array.
* **Frame:** Tracks local variables and return addresses for function calls.
* **Value:** 16-byte Tagged Union.

---

## 3. The Compiler (Python to Bytecode)

### 3.1 AST Transformation
* **Tool:** `github.com/go-python/gpython/ast`
* **Strategy:** Translates Python statements (Assign, If, While, Call, With) into VM opcodes.

### 3.2 Security Gating (The `with scope` Rule)
* **Mechanism:** Python `with scope(name, token):` blocks are compiled into `OP_ADDRESS` and `OP_EXIT_ADDR`.
* **Enforcement:** Syscalls (Fetch, WriteFile) are only permitted if the required scope is active.

---

## 4. Security & Standard Library

### 4.1 Standard Libraries (Sandboxed)

#### **FS-ENV (Filesystem)**
* **Words:** `write_file(content, path)`, `read_file(path)`
* **Constraint:** Root Jailing and Size Limits.

#### **HTTP-ENV (Network)**
* **Words:** `fetch(url)`
* **Constraint:** Strict Domain Allowlist.
