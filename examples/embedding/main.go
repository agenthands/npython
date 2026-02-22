package main

import (
	"fmt"
	"log"

	"github.com/agenthands/npython/pkg/compiler/python"
	"github.com/agenthands/npython/pkg/core/value"
	"github.com/agenthands/npython/pkg/stdlib"
	"github.com/agenthands/npython/pkg/vm"
)

// MyCustomFunction is a Go function exposed to nPython.
// It takes one integer argument from the VM stack, doubles it, and pushes it back.
func MyCustomFunction(m *vm.Machine) error {
	// 1. Pop arguments from the stack (in reverse order of appearance in Python call)
	val := m.Pop()

	if val.Type != value.TypeInt {
		return fmt.Errorf("TypeError: expected int, got %v", val.Type)
	}

	// 2. Perform logic
	result := val.Int() * 2

	// 3. Push result back to the stack
	m.Push(value.Value{
		Type: value.TypeInt,
		Data: uint64(result),
	})

	return nil
}

func main() {
	// Python code that calls our custom function.
	src := `
x = 21
y = custom_doubler(x)
print("Result from Go inside VM:", y)
y
`

	// 1. Pre-compile the Python code
	// Patch the compiler to recognize our custom function.
	python.PythonBuiltins["custom_doubler"] = 64

	compiler := python.NewCompiler()
	bc, err := compiler.Compile(src)
	if err != nil {
		log.Fatalf("Compile error: %v", err)
	}

	// 2. Get a VM instance
	m := vm.GetMachine()
	defer vm.PutMachine(m)

	// 3. Initialize Host Registry with standard built-ins and our custom one
	m.HostRegistry = make([]vm.HostFunctionEntry, 65)

	// Register 'print' at index 2 (matching PythonBuiltins standard)
	m.HostRegistry[2] = vm.HostFunctionEntry{
		Fn: stdlib.Print,
	}

	// Register our custom function at index 64
	m.HostRegistry[64] = vm.HostFunctionEntry{
		Fn: MyCustomFunction,
	}

	// 4. Load the bytecode
	m.Code = bc.Instructions
	m.Constants = bc.Constants
	m.Arena = bc.Arena

	// 5. Run the VM
	err = m.Run(1000) // 1000 gas limit
	if err != nil {
		log.Fatalf("VM error: %v", err)
	}

	// 6. Retrieve the result
	// The result of the last expression (y) is on top of the stack.
	finalResult := m.Pop()
	fmt.Printf("Final result in Go: %v\n", finalResult.Int())
}
