# nForth Engine: Technical Specification
**Version:** 1.0.0-ReleaseCandidate
**Architecture:** Zero-Allocation Bytecode VM with Capability-Based Security

---

## 1. System Architecture
The system follows a **Host-Guest Isolation Model**.
* **Host (Go Runtime):** Manages memory, I/O, and Capability Tokens.
* **Guest (nForth VM):** A pure logic engine. It has **zero** direct access to the OS.
* **AI-Native Design:** The LLM is completely shielded from stack operations. Only Named State (`INTO`) is allowed. Legacy words like `DUP` or `SWAP` are forbidden.

---

## 2. Virtual Machine (The Core)

### 2.1 Performance Invariants
* **Zero-Allocation:** The `Run()` loop MUST NOT allocate heap memory.
* **Panic-Based Error Handling:** Internal stack overflows/underflows trigger a `panic`, which is caught by a `defer recover()` at the API boundary to return safe `error` values.
* **Dispatch:** The Main Loop uses a **Big Switch** statement for maximum compiler optimization.

### 2.2 Memory Layout
* **Stack:** Fixed-size `[128]Value` array.
* **Value:** 16-byte Tagged Union (Type + 64-bit Data).
    * **Strings:** Stored in a **Byte Buffer Arena**. `Value.Data` holds a packed `(Offset << 32 | Length)` integer. No Go string headers on the stack.

### 2.3 Opcode Map
| Hex | Mnemonic | Description |
| :--- | :--- | :--- |
| `0x00` | `OP_HALT` | Stop execution. |
| `0x02` | `OP_PUSH_C` | Push constant from pool to stack. |
| `0x03` | `OP_PUSH_L` | Push local variable to stack. |
| `0x04` | `OP_POP_L` | **(INTO)** Pop stack top into local var. |
| `0x10` | `OP_ADD` | Integer addition. |
| `0x11` | `OP_SUB` | Integer subtraction. |
| `0x12` | `OP_MUL` | Integer multiplication. |
| `0x1a` | `OP_DIV` | Integer division. |
| `0x13` | `OP_EQ` | Equality check. |
| `0x1b` | `OP_NE` | Inequality check (`!=`). |
| `0x14` | `OP_GT` | Greater than. |
| `0x18` | `OP_LT` | Less than. |
| `0x15` | `OP_PRINT` | Print value (mocked). |
| `0x16` | `OP_CONTAINS` | String containment check. |
| `0x17` | `OP_ERROR` | Trigger nForth runtime error (`THROW`). |
| `0x20` | `OP_JMP` | Unconditional jump. |
| `0x21` | `OP_JMP_FALSE`| Jump if top is false. |
| `0x22` | `OP_CALL` | Call internal function. |
| `0x23` | `OP_RET` | Return from function (`YIELD`, `EXIT`). |
| `0x30` | `OP_ADDRESS`| Pushes a Security Scope to the active stack. |
| `0x31` | `OP_EXIT_ADDR`| Pops the current Security Scope. |
| `0x40` | `OP_SYSCALL`| Invokes a registered Host Function. |

---

## 3. The Compiler (Anti-Hallucination)

### 3.1 Lexer
* **Strategy:** Zero-Copy Cursor Scanner. Returns indices `(Start, End)`.
* **Token Sugar:**
    * `<ENV-GATE>` -> `TOKEN_SUGAR_GATE`
    * `->` -> `TOKEN_INTO`
    * `THE`, `WITH`, `USING`, `FROM` -> `TOKEN_NOISE` (Ignored)

### 3.2 Parser & Validator
* **The "INTO" Rule:** The parser tracks `VirtualStackDepth`. If a statement ends with `Depth > 0`, compilation **Halts Immediately** with a `SyntacticHallucinationError`.
* **Function Signatures:** The compiler tracks if a function is *Void* or *Fruitful* (uses `YIELD`). Calls to fruitful functions increment the stack depth, requiring an `INTO` grounding.
* **Scope Resolution:** Nested `ADDRESS` blocks create a **Cumulative Hierarchy**.

---

## 4. Security & Standard Library

### 4.1 Standard Libraries (Sandboxed)

#### **FS-ENV (Filesystem)**
* **Constraint 1:** **Root Jailing.** Paths resolved relative to workspace root.
* **Constraint 2:** **Capability Split.** `fs:read` and `fs:write` are separate tokens.
* **Constraint 3:** **Size Limits.** `WRITE-FILE` operations reject payloads > 5MB.

#### **HTTP-ENV (Network)**
* **Constraint 1:** **Strict Allowlist.** Capability Token must specify allowed domains.
* **Constraint 2:** **No Localhost.** Blocked by default unless `AllowLocalhost` is enabled for testing.

---

## 5. Verified E2E Scenarios
The following scenarios are verified in the `tests/main_test.go` suite:
1. **The Happy Path (Data Pipeline)**: Fetches data, validates containment, and writes reports.
2. **The Red Team (Jailbreak Attempt)**: Confirms the engine blocks path traversal attacks.
3. **The Anti-Hallucination (Compiler)**: Confirms the compiler rejects code with floating stack values.
4. **Logic & Arithmetic**: Verifies `DIV`, `NE`, `!=`, and nested comparison logic.
5. **Function Definitions**: Verifies argument passing and `YIELD` value returns.
6. **Early Exit**: Verifies `EXIT` correctly terminates execution.
