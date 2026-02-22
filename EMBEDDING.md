# nPython Embedding Guide

nPython is designed from the ground up for secure, high-performance integration into Go applications. This guide covers how to pre-compile Python code, execute it in the VM, extend the environment with custom Host Functions, and retrieve results back in Go.

## 1. Core Architecture

The system consists of three main packages:
- `pkg/compiler/python`: Translates Python source into a custom Bytecode format.
- `pkg/vm`: The stack-based Virtual Machine that executes bytecode.
- `pkg/core/value`: The common data representation for interoperability between Python and Go.

## 2. Basic Embedding Workflow

Embedding nPython involves compiling source code, obtaining a machine instance, and running it.

```go
package main

import (
    "fmt"
    "log"
    "github.com/agenthands/npython/pkg/compiler/python"
    "github.com/agenthands/npython/pkg/vm"
)

func main() {
    src := "x = 10; y = 20; x + y"

    // 1. Compile Python source to Bytecode
    compiler := python.NewCompiler()
    bc, err := compiler.Compile(src)
    if err != nil {
        log.Fatal(err)
    }

    // 2. Obtain a VM instance from the pool
    m := vm.GetMachine()
    defer vm.PutMachine(m)

    // 3. Load the pre-compiled bytecode
    m.Code = bc.Instructions
    m.Constants = bc.Constants
    m.Arena = bc.Arena

    // 4. Run the VM with a gas limit (instruction count)
    err = m.Run(1000) 
    if err != nil {
        log.Fatal(err)
    }

    // 5. Retrieve the result (the module-level expression 'x + y' stays on stack)
    result := m.Pop()
    fmt.Printf("Result: %d\n", result.Int()) // Output: 30
}
```

## 3. Extending with Host Functions (Go Extensions)

Host Functions allow the Python environment to call into Go performance-critical logic or restricted resources.

### Step 1: Define the Go Function
A Host Function must satisfy the signature `func(*vm.Machine) error`.

```go
func GoMultiply(m *vm.Machine) error {
    // Arguments are popped from the stack in reverse order
    b := m.Pop().Int()
    a := m.Pop().Int()
    
    result := a * b
    
    // Push result back to Python environment
    m.Push(value.Value{
        Type: value.TypeInt,
        Data: uint64(result),
    })
    return nil
}
```

### Step 2: Register and Map the Function
Registration involves mapping a Python name to a "Syscall Index" in the compiler and then providing the implementation in the VM's `HostRegistry`.

```go
func main() {
    // 1. Inform compiler about the new function
    python.PythonBuiltins["go_mul"] = 64 // Use indices >= 64 to avoid collisions

    // 2. Compile source using the new name
    bc, _ := python.NewCompiler().Compile("go_mul(7, 6)")

    m := vm.GetMachine()
    
    // 3. Register implementation in HostRegistry
    // HostRegistry must be sized to include your highest index
    m.HostRegistry = make([]vm.HostFunctionEntry, 65)
    m.HostRegistry[64] = vm.HostFunctionEntry{
        Fn: GoMultiply,
    }
    
    // ... load bytecode and Run() ...
}
```

## 4. Capability-Based Security (Sandboxing)

Host Functions can be restricted behind "Scope Tokens". If a function requires a scope that hasn't been granted via Python's `with scope(name, token):` block, the VM will return `ErrSecurityViolation`.

```go
m.HostRegistry[1] = vm.HostFunctionEntry{
    RequiredScope: "NETWORK",
    Fn: MyNetworkFetcher,
}
```

Python usage:
```python
# This will fail
fetch("http://example.com")

# This works (if token matches Gatekeeper validation)
with scope("NETWORK", "secret-token"):
    fetch("http://example.com")
```

## 5. Working with Python Values in Go

The `value.Value` type is a 128-bit tagged union designed for zero-allocation access to primitives.

| Python Type | Go Value Access |
| :--- | :--- |
| `int` | `val.Int()` (returns `int64`) |
| `bool` | `val.Data != 0` |
| `float` | `math.Float64frombits(val.Data)` |
| `str` | `value.UnpackString(val.Data, m.Arena)` |
| `list` | `val.Opaque.(*[]value.Value)` |
| `dict` | `val.Opaque.(map[string]any)` |

## 6. Performance Considerations

1. **Machine Pooling**: Always use `vm.GetMachine()` and `vm.PutMachine(m)` to avoid GC pressure.
2. **Arena Memory**: String allocations happen in the `m.Arena` byte slice. For high-volume string creation, monitor `MaxArenaSize`.
3. **Pre-compilation**: Compile source once and share the `Bytecode` struct across multiple VM runs/instances.
