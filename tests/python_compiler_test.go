package main_test

import (
	"testing"
	"github.com/agenthands/npython/pkg/compiler/python"
	"github.com/agenthands/npython/pkg/vm"
)

func TestPythonCompiler(t *testing.T) {
	src := `
x = 10
y = 20
z = x + y
`
	compiler := python.NewCompiler()
	bytecode, err := compiler.Compile(src)
	if err != nil {
		t.Fatalf("Compilation failed: %v", err)
	}

	machine := &vm.Machine{
		Code:      bytecode.Instructions,
		Constants: bytecode.Constants,
		Arena:     bytecode.Arena,
	}
	machine.Reset()

	err = machine.Run(100)
	if err != nil {
		t.Fatalf("VM Execution failed: %v", err)
	}

	zVal := machine.Frames[0].Locals[2]
	if zVal.Int() != 30 {
		t.Errorf("Expected z (local 2) to be 30, got %d", zVal.Int())
	}
}

func TestPythonIf(t *testing.T) {
	src := `
x = 10
if x > 5:
    y = 1
else:
    y = 0
`
	compiler := python.NewCompiler()
	bytecode, err := compiler.Compile(src)
	if err != nil {
		t.Fatalf("Compilation failed: %v", err)
	}

	machine := &vm.Machine{
		Code:      bytecode.Instructions,
		Constants: bytecode.Constants,
		Arena:     bytecode.Arena,
	}
	machine.Reset()

	err = machine.Run(100)
	if err != nil {
		t.Fatalf("VM Execution failed: %v", err)
	}

	yVal := machine.Frames[0].Locals[1]
	if yVal.Int() != 1 {
		t.Errorf("Expected y to be 1, got %d", yVal.Int())
	}
}

func TestPythonWhile(t *testing.T) {
	src := `
i = 0
count = 0
while i < 5:
    count = count + 2
    i = i + 1
`
	compiler := python.NewCompiler()
	bytecode, err := compiler.Compile(src)
	if err != nil {
		t.Fatalf("Compilation failed: %v", err)
	}

	machine := &vm.Machine{
		Code:      bytecode.Instructions,
		Constants: bytecode.Constants,
		Arena:     bytecode.Arena,
	}
	machine.Reset()

	err = machine.Run(500) // More gas for loop
	if err != nil {
		t.Fatalf("VM Execution failed: %v", err)
	}

	countVal := machine.Frames[0].Locals[1]
	if countVal.Int() != 10 {
		t.Errorf("Expected count to be 10, got %d", countVal.Int())
	}
}

func TestPythonString(t *testing.T) {
	src := `
s = "hello npython"
`
	compiler := python.NewCompiler()
	bytecode, err := compiler.Compile(src)
	if err != nil {
		t.Fatalf("Compilation failed: %v", err)
	}

	machine := &vm.Machine{
		Code:      bytecode.Instructions,
		Constants: bytecode.Constants,
		Arena:     bytecode.Arena,
	}
	machine.Reset()

	err = machine.Run(100)
	if err != nil {
		t.Fatalf("VM Execution failed: %v", err)
	}
}
