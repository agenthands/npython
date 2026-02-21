# nPython Engine: Technical Specification
**Version:** 1.1.0-ReleaseCandidate
**Architecture:** Zero-Allocation Python VM with Capability-Based Security

---

## 1. System Architecture
The system follows a **Host-Guest Isolation Model**.
* **Host (Go Runtime):** Manages memory, I/O, and Capability Tokens.
* **Guest (nPython VM):** A pure logic engine that executes bytecode compiled from a strict Python subset. It has **zero** direct access to the OS.
* **Python-Native Design:** The LLM writes standard Python. The compiler transforms this into secure bytecode.

---

## 2. The nPython Language Subset
nPython supports a robust subset of Python 3, optimized for agentic logic and security.

### 2.1 Core Types
*   **Scalar:** `int` (64-bit), `float` (64-bit), `bool`, `str` (immutable, utf-8).
*   **Collections:** `list` (mutable), `tuple` (immutable), `dict` (mutable key-value), `set` (mutable unique).
*   **Special:** `bytes` (immutable), `iterator`, `NoneType`.

### 2.2 Control Flow
*   `if` / `elif` / `else`
*   `while` loops (with `break` / `continue`)
*   `for` loops (over iterables using `iter()`/`next()` protocol under the hood)
*   **Functions:** `def` with arguments, return values, and recursion.

### 2.3 Built-in Functions
nPython provides a rich standard library without imports:

**Numeric & Math**
*   `abs(x)`, `divmod(a, b)`, `float(x)`, `int(x)`, `max(iter)`, `min(iter)`, `pow(base, exp)`, `round(x)`, `sum(iter)`

**Collections & Iteration**
*   `len(obj)`, `list(iter)`, `dict()`, `set(iter)`, `tuple(iter)`
*   `range(stop)`, `enumerate(iter)`, `zip(*iters)`, `reversed(seq)`, `sorted(iter)`
*   `filter(func, iter)`, `map(func, iter)`, `all(iter)`, `any(iter)`
*   `iter(obj)`, `next(iter)`

**Type Conversion & Formatting**
*   `bool(x)`, `str(x)`, `repr(x)`, `ascii(x)`, `chr(i)`, `ord(c)`
*   `bin(x)`, `hex(x)`, `oct(x)`, `format_string(fmt, *args)`
*   `bytes(source)`, `bytearray(source)`

**Introspection**
*   `type(obj)`, `callable(obj)`, `hash(obj)`, `id(obj)`
*   `locals()`, `globals()`

**Input/Output (Sandboxed)**
*   `print(*objects)`

### 2.4 Security Gates (`with scope`)
Privileged operations are **only** accessible within a `with scope` block.
```python
with scope("HTTP-ENV", token):
    fetch("https://api.example.com")
```

---

## 3. Virtual Machine (The Core)

### 3.1 Performance Invariants
*   **Zero-Allocation Hot Path:** The `Run()` loop avoids heap allocations during instruction execution.
*   **Frame Pointer Caching:** Local variable access is optimized via cached frame pointers.
*   **Fast-Path Equality:** String comparison avoids unpacking for identical references.

### 3.2 Memory Layout
*   **Stack:** Fixed-size `[128]Value` array per machine.
*   **Frames:** Fixed-size `[32]Frame` stack for call depth.
*   **Arena:** Byte slice for string/bytes storage, deduplicated during compilation.

---

## 4. Embedding & Extension (Host Interface)

### 4.1 Host Functions
The VM allows the Host to register Go functions that are callable from nPython.
*   **Registry:** `Machine.RegisterHostFunction(scope, func)`
*   **Signature:** `func(m *vm.Machine) error`
*   **Argument Passing:** Arguments are popped from the stack; results are pushed.

### 4.2 Calling nPython from Go
The Host can invoke nPython functions defined in the script.
*   **Method:** `Machine.Call(ip, args...)`
*   **Mechanism:** Pushes arguments, executes until return, and captures the result.

---

## 5. Security Model

### 5.1 Capability-Based Access Control
*   **Gatekeeper:** A Host-defined interface `Validate(scope, token)` checks permissions.
*   **Scope Stack:** The VM tracks active scopes. Syscalls fail if the required scope is not active.

### 5.2 Sandboxed Standard Libraries
*   **FS-ENV:** Root-jailed file access (`read_file`, `write_file`).
*   **HTTP-ENV:** Allowlist-based network access (`fetch`, `send_request`).
