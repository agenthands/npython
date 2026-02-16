package vm_test

import (
	"testing"
	"github.com/agenthands/nforth/pkg/vm"
	"github.com/agenthands/nforth/pkg/core/value"
)

func TestMachineReset(t *testing.T) {
	m := &vm.Machine{}
	
	// Dirty the machine
	m.SP = 10
	m.IP = 5
	m.FP = 2
	m.Stack[0] = value.Value{Type: value.TypeInt, Data: 100}

	m.Reset()

	if m.SP != 0 || m.IP != 0 || m.FP != 0 {
		t.Errorf("Reset failed: SP=%d, IP=%d, FP=%d", m.SP, m.IP, m.FP)
	}
	
	if m.Stack[0].Type != value.TypeVoid {
		t.Errorf("Reset failed to zero out stack")
	}
}

func TestMachineStackOps(t *testing.T) {
	m := &vm.Machine{}

	// Test Push/Pop
	m.Push(value.Value{Type: value.TypeInt, Data: 42})
	if m.SP != 1 {
		t.Errorf("expected SP=1, got %d", m.SP)
	}

	val := m.Pop()
	if val.Data != 42 {
		t.Errorf("expected 42, got %d", val.Data)
	}
	if m.SP != 0 {
		t.Errorf("expected SP=0, got %d", m.SP)
	}
}

func TestMachineStackOverflow(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic on stack overflow")
		}
	}()

	m := &vm.Machine{}
	for i := 0; i <= vm.StackDepth; i++ {
		m.Push(value.Value{Type: value.TypeInt, Data: uint64(i)})
	}
}

func TestMachineStackUnderflow(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic on stack underflow")
		}
	}()

	m := &vm.Machine{}
	m.Pop()
}

func TestMachineRun(t *testing.T) {
	m := &vm.Machine{}
	
	// Bytecode for: 1 2 ADD INTO result
	// Assuming result is Local 0
	m.Constants = []value.Value{
		{Type: value.TypeInt, Data: 1},
		{Type: value.TypeInt, Data: 2},
	}
	
	m.Code = []uint32{
		(uint32(vm.OP_PUSH_C) << 24) | 0, // PUSH_C 1
		(uint32(vm.OP_PUSH_C) << 24) | 1, // PUSH_C 2
		(uint32(vm.OP_ADD) << 24),        // ADD
		(uint32(vm.OP_POP_L) << 24) | 0,  // POP_L 0 (INTO)
		(uint32(vm.OP_HALT) << 24),       // HALT
	}

	// Initialize the first frame
	m.FP = 0
	
	err := m.Run(100) // 100 gas limit
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Assert Local 0 holds 3
	res := m.Frames[0].Locals[0]
	if res.Type != value.TypeInt || res.Data != 3 {
		t.Errorf("expected 3, got %v (Type: %v)", res.Data, res.Type)
	}
}
