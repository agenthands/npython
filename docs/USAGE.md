# nForth Usage Guide

nForth is a concatenative, stack-based language optimized for LLM code generation. It prioritizes **Explicit State** and **Zero Ambient Authority**.

## Core Concepts

### 1. The `INTO` Rule (State Grounding)
Unlike traditional Forth, nForth forbids "implicit" stack drift. Every value pushed to the stack *must* be consumed by an `INTO` assignment or a void function.

**Invalid:**
```forth
10 20 ADD  \ Error: Floating State Detected
```

**Valid:**
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
: SQUARE { n }
    n n MUL INTO result
    result RETURN
;

5 SQUARE INTO five_sq
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
| `ADD`, `SUB`, `MUL` | `( a b -- res )` | Arithmetic |
| `EQ`, `GT`, `LT` | `( a b -- bool )` | Comparison |
| `PRINT` | `( val -- )` | Print to stdout |
| `FETCH` | `( url -- data )` | HTTP GET (Requires `HTTP-ENV`) |
| `WRITE-FILE` | `( data path -- )` | FS Write (Requires `FS-ENV`) |
| `CONTAINS` | `( str pat -- bool )` | String search |
