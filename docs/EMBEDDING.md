# Embedding nPython in Go

nPython is designed to be a high-performance, secure scripting layer for Go applications.

## Quick Start

```go
package main

import (
    "fmt"
    "github.com/agenthands/npython/pkg/compiler/python"
    "github.com/agenthands/npython/pkg/vm"
    "github.com/agenthands/npython/pkg/core/value"
)

func main() {
    // 1. Compile Python Source
    compiler := python.NewCompiler()
    bytecode, err := compiler.Compile("x = 10 + 20")
    if err != nil {
        panic(err)
    }

    // 2. Initialize VM
    machine := vm.GetMachine()
    defer vm.PutMachine(machine)
    
    machine.Code = bytecode.Instructions
    machine.Constants = bytecode.Constants
    machine.Arena = bytecode.Arena

    // 3. Register Custom Host Function
    machine.RegisterHostFunction("", func(m *vm.Machine) error {
        fmt.Println("Hello from Host!")
        return nil
    })

    // 4. Run
    if err := machine.Run(1000); err != nil {
        panic(err)
    }
}
```

## Security Integration

To enforce security, implement the `vm.Gatekeeper` interface:

```go
type MyGatekeeper struct{}

func (g *MyGatekeeper) Validate(scope, token string) bool {
    // Check if token is valid for the requested scope
    return token == "valid-secret"
}

// Attach to machine
machine.Gatekeeper = &MyGatekeeper{}
```

## Calling nPython Functions

You can call functions defined in the Python script from Go:

```go
// Script: def add(a, b): return a + b
ip := bytecode.Functions["add"]
arg1 := value.Value{Type: value.TypeInt, Data: 10}
arg2 := value.Value{Type: value.TypeInt, Data: 20}

result, err := machine.Call(ip, arg1, arg2)
fmt.Println(result.Int()) // 30
```
