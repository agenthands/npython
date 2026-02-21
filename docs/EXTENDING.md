# Extending nPython

You can extend nPython by adding new built-in functions or security environments.

## Adding a Host Function

A Host Function is a Go function that interacts with the VM stack.

```go
// Signature: func(m *vm.Machine) error
func MyCustomFunc(m *vm.Machine) error {
    // 1. Pop arguments (reverse order)
    arg := m.Pop()
    
    // 2. Perform logic
    res := arg.Int() * 2
    
    // 3. Push result
    m.Push(value.Value{Type: value.TypeInt, Data: uint64(res)})
    
    return nil
}

// Register it (index must match compiler expectation if built-in)
machine.RegisterHostFunction("", MyCustomFunc)
```

## Adding a Security Environment

To add a new protected capability (e.g., `DB-ENV`):

1.  **Define the Scope Name:** e.g., "DB-ENV".
2.  **Register Protected Functions:**
    ```go
    machine.RegisterHostFunction("DB-ENV", func(m *vm.Machine) error {
        // This code only runs if "DB-ENV" scope is active
        return nil
    })
    ```
3.  **Update Gatekeeper:** Ensure your Gatekeeper validates tokens for "DB-ENV".

## Extending the Compiler

To support new Python syntax:

1.  **Update `pkg/compiler/python/compiler.go`**: Add a case to `emitStmt` or `emitExpr`.
2.  **Map to Opcodes**: Translate the AST node to existing VM opcodes or syscalls.
3.  **Update `PythonBuiltins`**: If adding a new built-in function, add it to the map with its HostRegistry index.
