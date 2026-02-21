package main_test

import (
	"encoding/binary"
	"testing"

	"github.com/agenthands/npython/pkg/core/value"
	"github.com/agenthands/npython/pkg/vm"
)

func FuzzRunVM(f *testing.F) {
	// Add some valid seed corpus
	f.Add([]byte{
		byte(vm.OP_PUSH_C), 0, 0, 0,
		byte(vm.OP_HALT), 0, 0, 0,
	})
	f.Add([]byte{
		byte(vm.OP_PUSH_C), 0, 0, 0, // push const 0
		byte(vm.OP_PUSH_C), 0, 0, 1, // push const 1
		byte(vm.OP_ADD), 0, 0, 0, // add
		byte(vm.OP_HALT), 0, 0, 0,
	})

	f.Fuzz(func(t *testing.T, data []byte) {
		// Only run if we have a valid multiple of 4 bytes (instruction size)
		if len(data)%4 != 0 {
			return
		}

		m := vm.GetMachine()
		// Always reset on return
		defer vm.PutMachine(m)

		// Convert random bytes into instruction opcodes
		code := make([]uint32, len(data)/4)
		for i := 0; i < len(code); i++ {
			code[i] = binary.BigEndian.Uint32(data[i*4 : (i+1)*4])
		}

		m.Code = code

		// Setup a sterile but usable sandbox state
		m.Constants = []value.Value{
			{Type: value.TypeInt, Data: 42},
			{Type: value.TypeString, Data: value.PackString(0, 5)},
		}
		m.Arena = []byte("hello")

		m.HostRegistry = []vm.HostFunctionEntry{
			{RequiredScope: "", Fn: func(*vm.Machine) error { return nil }},
		}
		m.FunctionRegistry = map[string]int{"mock": 0}

		// Execute with a small gas limit to prevent infinite loops in fuzzed jumps
		_ = m.Run(1000)
	})
}
