# Embedding nPython in Go

The nPython Engine is designed to be easily embedded as a secure logic layer within any Go application.

## Integration Steps

### 1. Compile Source to Bytecode
The `compiler` package takes raw source and produces a `Bytecode` struct containing instructions, constants, and the string arena.

```go
import (
    "github.com/agenthands/npython/pkg/compiler/lexer"
    "github.com/agenthands/npython/pkg/compiler/parser"
    "github.com/agenthands/npython/pkg/compiler/emitter"
)

src := []byte(`10 20 ADD INTO result`)
s := lexer.NewScanner(src)
p := parser.NewParser(s, src)
prog, _ := p.Parse()

e := emitter.NewEmitter(src)
bc, _ := e.Emit(prog)
```

### 2. Initialize the Virtual Machine
Create a `vm.Machine` and load the compiled bytecode.

```go
import "github.com/agenthands/npython/pkg/vm"

m := &vm.Machine{}
m.Code = bc.Instructions
m.Constants = bc.Constants
m.Arena = bc.Arena
```

### 3. Setup Security (Optional)
If your script uses `ADDRESS` gates, you must provide a `Gatekeeper`.

```go
type MyGatekeeper struct{}
func (g *MyGatekeeper) Validate(scope, token string) bool {
    return token == "secret" // Your logic here
}

m.Gatekeeper = &MyGatekeeper{}
```

### 4. Execute with Gas Limits
The `Run()` loop executes bytecode until completion or until the gas limit is reached.

```go
err := m.Run(1000000) // Execute up to 1M instructions
if err != nil {
    panic(err)
}
```

### 5. Accessing Results
After execution, you can inspect the machine's state or local variables in the first frame.

```go
result := m.Frames[0].Locals[0]
fmt.Printf("Result: %v
", result.Data)
```
