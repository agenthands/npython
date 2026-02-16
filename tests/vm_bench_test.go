package tests

import (
	"testing"
	"github.com/agenthands/nforth/pkg/vm"
	"github.com/agenthands/nforth/pkg/core/value"
)

func BenchmarkVMRun(b *testing.B) {
	m := &vm.Machine{}
	
	// Bytecode: 1 1 ADD 1 ADD 1 ADD ... (repeat to fill gas)
	// We want a tight loop that doesn't halt early
	
	m.Constants = []value.Value{
		{Type: value.TypeInt, Data: 1},
	}
	
	// Create a long sequence of ADDs
	code := make([]uint32, 1001)
	code[0] = (uint32(vm.OP_PUSH_C) << 24) | 0
	for i := 1; i < 1000; i++ {
		if i % 2 == 1 {
			code[i] = (uint32(vm.OP_PUSH_C) << 24) | 0
		} else {
			code[i] = (uint32(vm.OP_ADD) << 24)
		}
	}
	code[1000] = (uint32(vm.OP_HALT) << 24)
	
	m.Code = code

	b.ResetTimer()
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		m.Reset()
		_ = m.Run(2000)
	}
}
