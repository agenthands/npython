# Production-Grade Specification for the nFORTH Engine (v1.0)

## 1. System Overview
The nFORTH Engine is a **secure, zero-allocation Bytecode Virtual Machine** designed to execute the Natural FORTH language. It acts as the actuation layer for AI Agents, replacing "ambient authority" interpreters (like Python) with a constrained, capability-gated execution environment.

### 1.1 Architectural Invariants
1. **Host-Guest Isolation:** The Go Runtime (`Host`) controls the nFORTH VM (`Guest`). The Guest has **zero** access to the OS unless explicitly granted via a Capability Token.
2. **Zero-Allocation Hot Path:** The main execution loop (`vm.Run`) must not generate garbage. All run-time memory is pre-allocated via `sync.Pool` or on the stack.
3. **Canonical State Enforcement:** The VM enforces the "Named State" paradigm. Values cannot drift on the stack; they must be consumed into local variables immediately.

## 2. Bytecode & Opcode Map
The VM operates on a flat slice of `uint32` instructions. To ensure performance and simplicity, we use a **Register-Based / Stack-Hybrid** approach where locals are accessed via frame indices.

### 2.1 Instruction Format
Each instruction is a `uint32`:
* **Top 8 bits:** Opcode (`0x00` - `0xFF`)
* **Lower 24 bits:** Operand (Immediate value, Constant Pool Index, or Frame Index)

### 2.2 The Opcode Map
| Hex | Mnemonic | Operand | Stack Transition | Description |
| --- | --- | --- | --- | --- |
| **0x00** | `OP_HALT` | - | - | Stop execution. |
| **0x01** | `OP_NOOP` | - | - | No operation. |
| **0x02** | `OP_PUSH_C` | `ConstIdx` | `( -- val )` | Push constant from pool to stack. |
| **0x03** | `OP_PUSH_L` | `FrameIdx` | `( -- val )` | Push local variable to stack. |
| **0x04** | `OP_POP_L` | `FrameIdx` | `( val -- )` | **(INTO)** Pop stack top into local var. |
| **0x10** | `OP_ADD` | - | `( a b -- a+b )` | Integer addition. |
| **0x11** | `OP_SUB` | - | `( a b -- a-b )` | Integer subtraction. |
| **0x12** | `OP_MUL` | - | `( a b -- a*b )` | Integer multiplication. |
| **0x13** | `OP_EQ` | - | `( a b -- bool )` | Equality check. |
| **0x14** | `OP_GT` | - | `( a b -- bool )` | Greater than. |
| **0x20** | `OP_JMP` | `Offset` | - | Unconditional jump (relative). |
| **0x21** | `OP_JMP_FALSE` | `Offset` | `( bool -- )` | Jump if top is false (pop). |
| **0x22** | `OP_CALL` | `FuncIdx` | - | Call internal function (push frame). |
| **0x23** | `OP_RET` | - | - | Return from function (pop frame). |
| **0x30** | `OP_ADDRESS` | `ScopeIdx` | `( token -- )` | **Security Gate.** Activate capability scope. |
| **0x31** | `OP_EXIT_ADDR` | - | - | Close current security scope. |
| **0x40** | `OP_SYSCALL` | `HostFnIdx` | `( ... -- ... )` | Call registered Host Function (I/O). |

## 3. Language Specification (EBNF)

```ebnf
program          = { definition | statement } ;
definition       = ":" , ws , identifier , ws , locals_decl , ws , block , ";" ;
locals_decl      = "{" , { ws , identifier } , "}" ;
block            = { ws , statement } ;
statement        = assignment | void_operation | control_flow | security_gate ;
assignment       = expression , ws , arrow_op , ws , identifier ;
arrow_op         = "INTO" | "->" ;
expression       = term , { ws , term } ;
void_operation   = identifier , { ws , term } ;
security_gate    = explicit_addr | sugar_addr ;
explicit_addr    = "ADDRESS" , ws , env_identifier , ws , cap_token ;
sugar_addr       = "<" , env_identifier , "-GATE>" ; 
control_flow     = if_stmt ;
if_stmt          = term , ws , "IF" , ws , block , [ "ELSE" , ws , block ] , "THEN" ;
term             = literal | identifier ;
identifier       = letter , { letter | digit | "-" | "_" } ;
literal          = string_lit | number_lit | bool_lit ;
string_lit       = '"' , { ? all_chars_except_quote ? } , '"' ;
number_lit       = [ "-" ] , digit , { digit } ;
bool_lit         = "TRUE" | "FALSE" ;
ws               = ? whitespace ? ;
```

## 4. Module Specifications

### 4.1 `pkg/compiler` (The Parser)
**Validation Rules:**
1. **The "Dangling Stack" Check:** If `stack_depth > 0` at the end of a statement (not consumed by `INTO` or a void op), raise `ErrFloatingState`.
2. **The "Scope" Check:** Restricted words (e.g., `FETCH`) require an active `ADDRESS` block in the current or parent scope; otherwise, raise `ErrSecurityViolation`.

*(Other modules: pkg/core/value, pkg/vm, pkg/security as previously defined)*

## 5. File & Directory Structure
```text
/
├── cmd/
│   └── engine/             # CLI entry point
├── pkg/
│   ├── core/
│   │   └── value/          # Data representation
│   ├── vm/                 # Core VM
│   ├── compiler/           # Source processing (Parser/Lexer/Emitter)
│   ├── security/           # Capability logic
│   └── stdlib/             # Host function implementations
└── tests/                  # Integration tests
```

## 6. Implementation Roadmap
1. **Phase 1: The Core** (Value Type, VM Machine, Step Loop)
2. **Phase 2: The Compiler** (Lexer, EBNF Parser, Validation Logic)
3. **Phase 3: Security & Stdlib** (ADDRESS, SYSCALL, Safe I/O)
