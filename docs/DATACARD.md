# nForth Language Datacard (v1.0)
Output ONLY valid nForth code.

### **Mandatory Syntax Pillars**
1. **Word Definition**: `: NAME { arg1 arg2 } ... ;`
2. **Explicit State (INTO)**: Every value producer MUST be followed by `INTO <var>`.
3. **Capability Gating**: `ADDRESS <ENV> <TOKEN> ... <EXIT>`
4. **Postfix Logic**: `10 20 ADD INTO sum`

### **The Banned List**
- **NO Stack Juggling**: `DUP`, `SWAP`, `ROT`, `OVER`, `DROP` are forbidden.
- **NO Ambient Authority**: Never call `FETCH` or `WRITE` without an `ADDRESS` block.
- **NO Implicit Returns**: Use `YIELD` to explicitly return a value from a function.

### **Common Operations**
- `ADD`, `SUB`, `MUL`, `DIV` (Arithmetic)
- `EQ`, `NE`, `GT`, `LT` (Logic)
- `PRINT`, `THROW` (I/O & Errors)
- `CONTAINS` (String search)
- `FETCH` (HTTP GET - Requires `HTTP-ENV`)
- `WRITE-FILE` (FS Write - Requires `FS-ENV`)

### **Code Examples**

**1. Data Extraction & Storage:**
```forth
: SAVE-DATA { url path token }
  ADDRESS HTTP-ENV token
    url FETCH INTO raw
  <EXIT>
  ADDRESS FS-ENV token
    raw path WRITE-FILE
  <EXIT>
;
```

**2. Conditional Logic:**
```forth
: CHECK-VALUE { val }
  val 100 GT IF
    "High" PRINT
  ELSE
    "Low" PRINT
  THEN
;
```

### **Token Efficiency vs Python**
| Metric | nForth | Python |
| :--- | :--- | :--- |
| **Logic** | Postfix (High Density) | Infix (Verbose) |
| **Structure** | Space-delimited | Indentation + Brackets |
| **Boilerplate** | Zero (Sandboxed) | Extensive (Imports/Try-Except) |
