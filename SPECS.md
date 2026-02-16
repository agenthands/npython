# nFORTH Engine & Gemini CLI: Technical Specification
**Version:** 1.0.0-ReleaseCandidate
**Architecture:** Zero-Allocation Bytecode VM with Capability-Based Security

---

## 1. System Architecture
The system follows a **Host-Guest Isolation Model**.
* **Host (Go Runtime):** Manages memory, I/O, and Capability Tokens.
* **Guest (nFORTH VM):** A pure logic engine. It has **zero** direct access to the OS.
* **Bridge (The Gateway):** The `OP_SYSCALL` instruction, guarded by `OP_ADDRESS`.

---

## 2. Virtual Machine (The Core)

### 2.1 Performance Invariants
* **Zero-Allocation:** The `Run()` loop MUST NOT allocate heap memory.
* **Panic-Based Error Handling:** Internal stack overflows/underflows trigger a `panic`, which is caught by a `defer recover()` at the API boundary.
* **Dispatch:** The Main Loop uses a **Big Switch** statement for maximum compiler optimization.

### 2.2 Memory Layout
* **Stack:** Fixed-size `[128]Value` array.
* **Value:** 16-byte Tagged Union (Type + 64-bit Data).
    * **Strings:** Stored in a **Byte Buffer Arena**. `Value.Data` holds a packed `(Offset << 32 | Length)` integer.

### 2.3 Opcode Map (Partial)
| Opcode | Mnemonic | Description |
| :--- | :--- | :--- |
| `0x04` | `OP_POP_L` | **(INTO)** Pops stack to Local Var. Enforces explicit state. |
| `0x30` | `OP_ADDRESS`| Pushes a Security Scope to the active stack. |
| `0x31` | `OP_EXIT_ADDR`| Pops the current Security Scope. |
| `0x40` | `OP_SYSCALL`| Invokes a Host Function (Dynamic Registry). |

---

## 3. The Compiler (Anti-Hallucination)

### 3.1 Lexer
* **Strategy:** Zero-Copy Cursor Scanner. Returns indices `(Start, End)`.
* **Token Sugar:**
    * `<ENV-GATE>` -> `TOKEN_SUGAR_GATE`
    * `->` -> `TOKEN_INTO`
    * `THE`, `WITH` -> `TOKEN_NOISE` (Ignored)

### 3.2 Parser & Validator
* **The "INTO" Rule:** The parser tracks `VirtualStackDepth`. If a statement ends with `Depth > 0`, compilation **Halts Immediately**.
* **Scope Resolution:** Nested `ADDRESS` blocks create a **Cumulative Hierarchy**.

---

## 4. Security & Standard Library

### 4.1 The Gateway Protocol
* **Dynamic Registry:** Host functions are registered at runtime via `vm.Register(func)`.
* **Privilege Drop:** Scopes are strictly LIFO. `OP_EXIT_ADDR` must be emitted at the end of every `ADDRESS` block.

### 4.2 Standard Libraries (Sandboxed)

#### **FS-ENV (Filesystem)**
* **Constraint 1:** **Root Jailing.** Paths resolved relative to workspace root.
* **Constraint 2:** **Capability Split.** `fs:read` and `fs:write` are separate tokens.
* **Constraint 3:** **Size Limits.** `WRITE` operations reject payloads > 5MB.

#### **HTTP-ENV (Network)**
* **Constraint 1:** **Strict Allowlist.** Capability Token must specify allowed domains.
* **Constraint 2:** **No Localhost.** Block `127.0.0.1`, `localhost`, and internal IP ranges.
