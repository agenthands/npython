# nForth Usage Guide

nForth is a concatenative, stack-based language optimized for LLM code generation. It prioritizes **Explicit State** and **Zero Ambient Authority**.

## 🧠 AI-Native Philosophy

**The Problem:** Traditional stack languages (Forth) require mental simulation of the stack state (e.g., `DUP SWAP DROP`). For Large Language Models, this hidden state acts as a "cognitive load," leading to "stack drift" hallucinations where the model loses track of variables.

**The nForth Solution:**
We enforce a strict **Named State** paradigm. The LLM interacts *purely* with named variables.

> **Rule:** No `PUSH`, `POP`, `DUP`, or `SWAP`.

The LLM is completely shielded from the underlying stack reality.

## 🛡️ Compiler-Level Enforcement

Under the hood, the Go Virtual Machine still executes blazing-fast stack bytecodes. However, the `pkg/compiler` runs a strict static analysis pass before execution.

If the LLM hallucinates and attempts to:
1. Leave a value on the stack without an accompanying `INTO` statement.
2. Use a legacy stack word (which doesn't exist in the dictionary).

The compiler intercepts it and throws a `SyntacticHallucinationError`. The code is rejected before it ever runs, ensuring the agent's state remains perfectly synchronized and secure.

## Core Concepts

### 1. The `INTO` Rule (State Grounding)
Every data transformation must explicitly name its output state.

**Invalid (Hallucination):**
```forth
10 20 ADD  \ Error: Syntactic Hallucination. Stack not empty.
```

**Valid (Grounded):**
```forth
10 20 ADD INTO sum
```

### 2. Control Flow
nForth uses structured control flow to prevent chaotic jumps.

#### IF/ELSE/THEN
The condition must be evaluated immediately before the `IF`.
```forth
i 10 LT IF
    "Small" PRINT
ELSE
    "Large" PRINT
THEN
```

#### BEGIN/WHILE/REPEAT
A standard loop that checks a condition at the start of every iteration.
```forth
0 INTO i
BEGIN
    i 10 LT
WHILE
    i PRINT
    i 1 ADD INTO i
REPEAT
```

### 3. Function Definitions
Functions are defined using the `:` word. Arguments are named in `{ }` and are automatically popped into local variables at the start of the function.

```forth
: CALC-TOTAL { price tax-rate }
  price tax-rate MUL 100 DIV INTO tax-amount
  price tax-amount ADD INTO total
  total RETURN
;
```

### 4. Security Gates (`ADDRESS`)
Privileged operations (like `FETCH` or `WRITE-FILE`) are only accessible within an `ADDRESS` block.

```forth
ADDRESS FS-ENV "my-token"
    "Hello" "log.txt" WRITE-FILE
<EXIT>
```

## Built-in Words

| Word | Stack | Description |
| :--- | :--- | :--- |
| `ADD`, `SUB`, `MUL`, `DIV` | `( a b -- res )` | Arithmetic |
| `EQ`, `NE`, `GT`, `LT` | `( a b -- bool )` | Comparison (`!=` is alias for `NE`) |
| `PRINT` | `( val -- )` | Print to stdout |
| `FETCH` | `( url -- data )` | HTTP GET (Requires `HTTP-ENV`) |
| `WRITE-FILE` | `( data path -- )` | FS Write (Requires `FS-ENV`) |
| `CONTAINS` | `( str pat -- bool )` | String search |
| `YIELD` | `( val -- )` | Explicitly return a value from a function |
| `EXIT` | `( -- )` | Early exit from function or block |
| `THROW` | `( msg -- )` | Raise a runtime error |
