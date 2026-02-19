# Extending nPython

You can extend nPython by adding new host functions (syscalls) and custom security scopes.

## Adding a New Host Function

### 1. Implement the Go Function
Host functions must match the `vm.HostFunction` signature: `func(*Machine) error`. They interact directly with the VM's stack.

```go
func MyCustomTool(m *vm.Machine) error {
    // 1. Pop arguments
    val := m.Pop()
    
    // 2. Perform logic
    fmt.Printf("Custom tool received: %v
", val.Data)
    
    // 3. Push result back
    m.Push(value.Value{Type: value.TypeInt, Data: 1})
    return nil
}
```

### 2. Register with the Machine
Assign your function to a security scope (or use `""` for globally accessible logic).

```go
// Register returns a uint32 ID that the Emitter must use.
m.RegisterHostFunction("MY-SCOPE", MyCustomTool)
```

### 3. Update the Compiler
To make the new word available in nPython source code, you must add it to the `parser.StandardWords` map.

```go
// pkg/compiler/parser/parser.go

var StandardWords = map[string]OpSignature{
    // ...
    "MY-TOOL": {In: 1, Out: 1, RequiredScope: "MY-SCOPE"},
}
```

And update the `Emitter` to map the word to the correct host index:

```go
// pkg/compiler/emitter/emitter.go

case "MY-TOOL":
    e.emitOp(vm.OP_SYSCALL, 3) // Assuming it's the 4th registered function
```

## Creating Custom Sandboxes
Follow the pattern in `pkg/stdlib/fs.go` or `pkg/stdlib/http.go` to create isolated environments with root jailing or validation logic.
