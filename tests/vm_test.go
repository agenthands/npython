package tests

import (
	"testing"
	"github.com/agenthands/nforth/pkg/vm"
	"github.com/agenthands/nforth/pkg/core/value"
)

func TestCoreOpcodes(t *testing.T) {
	m := &vm.Machine{}
	
	// Bytecode: 10 20 ADD INTO result
	m.Constants = []value.Value{
		{Type: value.TypeInt, Data: 10},
		{Type: value.TypeInt, Data: 20},
	}
	
	m.Code = []uint32{
		(uint32(vm.OP_PUSH_C) << 24) | 0, // PUSH_C 10
		(uint32(vm.OP_PUSH_C) << 24) | 1, // PUSH_C 20
		(uint32(vm.OP_ADD) << 24),        // ADD
		(uint32(vm.OP_POP_L) << 24) | 0,  // POP_L 0
		(uint32(vm.OP_HALT) << 24),       // HALT
	}

	err := m.Run(10)
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	res := m.Frames[0].Locals[0]
	if res.Data != 30 {
		t.Errorf("expected 30, got %d", res.Data)
	}
}

func TestStackUnderflowSafety(t *testing.T) {
	m := &vm.Machine{}
	m.Code = []uint32{
		(uint32(vm.OP_ADD) << 24), // ADD with empty stack
		(uint32(vm.OP_HALT) << 24),
	}

	err := m.Run(10)
	if err != vm.ErrStackUnderflow {
		t.Errorf("expected ErrStackUnderflow, got %v", err)
	}
}
