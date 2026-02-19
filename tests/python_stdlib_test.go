package main_test

import (
	"testing"
	"github.com/agenthands/npython/pkg/compiler/python"
	"github.com/agenthands/npython/pkg/vm"
	"github.com/agenthands/npython/pkg/core/value"
	"github.com/agenthands/npython/pkg/stdlib"
)

type MockGate struct{}
func (m *MockGate) Validate(scope, token string) bool {
	return token == "valid-token"
}

func TestPythonStdlib(t *testing.T) {
	src := `
with scope("HTTP-ENV", "valid-token"):
    res = fetch("http://example.com")

if res == "MOCK_BODY":
    with scope("FS-ENV", "valid-token"):
        write_file("OK", "result.txt")
`
	compiler := python.NewCompiler()
	bytecode, err := compiler.Compile(src)
	if err != nil {
		t.Fatalf("Compilation failed: %v", err)
	}

	machine := &vm.Machine{
		Code:       bytecode.Instructions,
		Constants:  bytecode.Constants,
		Arena:      bytecode.Arena,
		Gatekeeper: &MockGate{},
	}
	machine.Reset()

	// Register Host Functions (matching PythonBuiltins indices)
	// 0: write_file
	// 1: fetch
	machine.RegisterHostFunction("FS-ENV", func(m *vm.Machine) error {
		pathPacked := m.Pop().Data
		bodyPacked := m.Pop().Data
		path := value.UnpackString(pathPacked, m.Arena)
		body := value.UnpackString(bodyPacked, m.Arena)
		if path != "result.txt" || body != "OK" {
			t.Errorf("write_file got unexpected args: %s, %s", path, body)
		}
		return nil
	})
	machine.RegisterHostFunction("HTTP-ENV", func(m *vm.Machine) error {
		urlPacked := m.Pop().Data
		url := value.UnpackString(urlPacked, m.Arena)
		if url != "http://example.com" {
			t.Errorf("fetch got unexpected url: %s", url)
		}
		// Push Mock Result
		offset := uint32(len(m.Arena))
		body := "MOCK_BODY"
		m.Arena = append(m.Arena, []byte(body)...)
		m.Push(value.Value{Type: value.TypeString, Data: value.PackString(offset, uint32(len(body)))})
		return nil
	})

	err = machine.Run(1000)
	if err != nil {
		t.Fatalf("VM Execution failed: %v", err)
	}
}

func TestPythonStringOps(t *testing.T) {
	src := `
s = format_string("Hello %s", "nPython")
if s == "Hello nPython":
    ok1 = 1
else:
    ok1 = 0

if is_empty(""):
    ok2 = 1
else:
    ok2 = 0

if is_empty("not empty"):
    ok3 = 0
else:
    ok3 = 1
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

	// Register Host Functions (matching PythonBuiltins indices)
	machine.HostRegistry = make([]vm.HostFunctionEntry, 11)
	machine.HostRegistry[9] = vm.HostFunctionEntry{Fn: stdlib.FormatString}
	machine.HostRegistry[10] = vm.HostFunctionEntry{Fn: stdlib.IsEmpty}

	err = machine.Run(1000)
	if err != nil {
		t.Fatalf("VM Execution failed: %v", err)
	}

	// Verify results in locals
	if machine.Frames[0].Locals[1].Int() != 1 { t.Errorf("ok1 failed") }
	if machine.Frames[0].Locals[2].Int() != 1 { t.Errorf("ok2 failed") }
	if machine.Frames[0].Locals[3].Int() != 1 { t.Errorf("ok3 failed") }
}

func TestPythonHttpBuilder(t *testing.T) {
	src := `
with scope("HTTP-ENV", "valid-token"):
    with_client()
    set_url("http://example.com/api")
    set_method("POST")
    resp = send_request()
    status = check_status(resp)
`
	compiler := python.NewCompiler()
	bytecode, err := compiler.Compile(src)
	if err != nil {
		t.Fatalf("Compilation failed: %v", err)
	}

	machine := &vm.Machine{
		Code:       bytecode.Instructions,
		Constants:  bytecode.Constants,
		Arena:      bytecode.Arena,
		Gatekeeper: &MockGate{},
	}
	machine.Reset()

	// Register Host Functions
	httpSandbox := stdlib.NewHTTPSandbox([]string{"example.com"})
	httpSandbox.AllowLocalhost = true
	
	machine.HostRegistry = make([]vm.HostFunctionEntry, 16)
	machine.HostRegistry[11] = vm.HostFunctionEntry{Fn: httpSandbox.WithClient}
	machine.HostRegistry[12] = vm.HostFunctionEntry{Fn: httpSandbox.SetURL}
	machine.HostRegistry[13] = vm.HostFunctionEntry{Fn: httpSandbox.SetMethod}
	machine.HostRegistry[5] = vm.HostFunctionEntry{RequiredScope: "HTTP-ENV", Fn: func(m *vm.Machine) error {
		// Mock SendRequest
		respMap := map[string]any{
			"status": int64(201),
			"body":   value.Value{Type: value.TypeString},
		}
		m.Push(value.Value{Type: value.TypeMap, Opaque: respMap})
		return nil
	}}
	machine.HostRegistry[6] = vm.HostFunctionEntry{Fn: httpSandbox.CheckStatus}

	err = machine.Run(1000)
	if err != nil {
		t.Fatalf("VM Execution failed: %v", err)
	}

	statusVal := machine.Frames[0].Locals[1]
	if statusVal.Int() != 201 {
		t.Errorf("Expected status 201, got %d", statusVal.Int())
	}
}


