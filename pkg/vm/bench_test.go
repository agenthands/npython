package vm_test

import (
	"testing"
	"github.com/agenthands/npython/pkg/vm"
	"github.com/agenthands/npython/pkg/core/value"
)

func BenchmarkVMLoop(b *testing.B) {
	// Simple counter loop:
	// 0 INTO i
	// BEGIN i 1000 LT WHILE i 1 ADD INTO i REPEAT
	
	// Bytecode manually constructed for efficiency
	// 0: PUSH_C 0 (0)
	// 1: POP_L 0 (i)
	// 2: PUSH_L 0 (i)
	// 3: PUSH_C 1 (1000)
	// 4: LT
	// 5: JMP_FALSE 10
	// 6: PUSH_L 0 (i)
	// 7: PUSH_C 2 (1)
	// 8: ADD
	// 9: POP_L 0 (i)
	// 10: JMP 2
	// 11: HALT
	
	code := []uint32{
		(uint32(vm.OP_PUSH_C) << 24) | 0,
		(uint32(vm.OP_POP_L) << 24) | 0,
		(uint32(vm.OP_PUSH_L) << 24) | 0,
		(uint32(vm.OP_PUSH_C) << 24) | 1,
		(uint32(vm.OP_LT) << 24),
		(uint32(vm.OP_JMP_FALSE) << 24) | 11,
		(uint32(vm.OP_PUSH_L) << 24) | 0,
		(uint32(vm.OP_PUSH_C) << 24) | 2,
		(uint32(vm.OP_ADD) << 24),
		(uint32(vm.OP_POP_L) << 24) | 0,
		(uint32(vm.OP_JMP) << 24) | 2,
		(uint32(vm.OP_HALT) << 24),
	}
	
	constants := []value.Value{
		{Type: value.TypeInt, Data: 0},
		{Type: value.TypeInt, Data: 1000},
		{Type: value.TypeInt, Data: 1},
	}
	
	m := &vm.Machine{
		Code:      code,
		Constants: constants,
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.Reset()
		m.Code = code
		m.Constants = constants
		err := m.Run(10000)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkStringEquality(b *testing.B) {
	// Benchmark OP_EQ with strings
	// "hello" "hello" EQ
	
	code := []uint32{
		(uint32(vm.OP_PUSH_C) << 24) | 0,
		(uint32(vm.OP_PUSH_C) << 24) | 1,
		(uint32(vm.OP_EQ) << 24),
		(uint32(vm.OP_HALT) << 24),
	}
	constants := []value.Value{
		{Type: value.TypeString, Data: value.PackString(0, 5)},
		{Type: value.TypeString, Data: value.PackString(5, 5)},
	}
	arena := []byte("hellohello")

	m := &vm.Machine{}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.Reset()
		m.Code = code
		m.Constants = constants
		m.Arena = arena
		err := m.Run(100)
		if err != nil {
			b.Fatal(err)
		}
	}
}
