package vm_test

import (
	"testing"
	"github.com/agenthands/nforth/pkg/vm"
	"github.com/agenthands/nforth/pkg/core/value"
)

// MockGatekeeper for testing ADDRESS validation
type MockGatekeeper struct {
	ValidTokens map[string]string
}

func (g *MockGatekeeper) Validate(scope, token string) bool {
	return g.ValidTokens[scope] == token
}

func TestSecurityScopes(t *testing.T) {
	m := &vm.Machine{}
	m.Gatekeeper = &MockGatekeeper{
		ValidTokens: map[string]string{
			"HTTP-ENV": "secret-http-token",
		},
	}

	// Bytecode: 
	// 1. PUSH "HTTP-ENV"
	// 2. PUSH "secret-http-token"
	// 3. OP_ADDRESS
	// 4. OP_EXIT_ADDR
	// 5. HALT
	
	m.Constants = []value.Value{
		{Type: value.TypeString, Data: value.PackString(0, 8)}, // "HTTP-ENV"
		{Type: value.TypeString, Data: value.PackString(8, 17)}, // "secret-http-token"
	}
	m.Arena = []byte("HTTP-ENVsecret-http-token")
	
	m.Code = []uint32{
		(uint32(vm.OP_PUSH_C) << 24) | 0,
		(uint32(vm.OP_PUSH_C) << 24) | 1,
		(uint32(vm.OP_ADDRESS) << 24),
		(uint32(vm.OP_EXIT_ADDR) << 24),
		(uint32(vm.OP_HALT) << 24),
	}

	err := m.Run(100)
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}
}

func TestSyscallSecurity(t *testing.T) {
	m := &vm.Machine{}
	
	// Register a mock host function that requires HTTP-ENV
	m.RegisterHostFunction("HTTP-ENV", func(m *vm.Machine) error {
		return nil
	})

	// Bytecode: 
	// 1. OP_SYSCALL 0 (Should fail because no scope is active)
	// 2. HALT
	m.Code = []uint32{
		(uint32(vm.OP_SYSCALL) << 24) | 0,
		(uint32(vm.OP_HALT) << 24),
	}

	err := m.Run(100)
	if err == nil {
		t.Error("expected security violation error, got nil")
	}
}
