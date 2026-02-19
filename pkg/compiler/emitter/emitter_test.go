package emitter_test

import (
	"testing"
	"github.com/agenthands/npython/pkg/compiler/emitter"
	"github.com/agenthands/npython/pkg/compiler/lexer"
	"github.com/agenthands/npython/pkg/compiler/parser"
	"github.com/agenthands/npython/pkg/vm"
)

func TestEmitterBasic(t *testing.T) {
	src := []byte("1 2 ADD INTO result")
	s := lexer.NewScanner(src)
	p := parser.NewParser(s, src)
	prog, err := p.Parse()
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	e := emitter.NewEmitter(src)
	bc, err := e.Emit(prog)
	if err != nil {
		t.Fatalf("Emit failed: %v", err)
	}

	// Expected: PUSH_C(1), PUSH_C(2), ADD, POP_L(0), HALT
	expectedOps := []uint8{
		vm.OP_PUSH_C,
		vm.OP_PUSH_C,
		vm.OP_ADD,
		vm.OP_POP_L,
		vm.OP_HALT,
	}

	if len(bc.Instructions) != len(expectedOps) {
		t.Fatalf("expected %d instructions, got %d", len(expectedOps), len(bc.Instructions))
	}

	for i, op := range expectedOps {
		gotOp := uint8(bc.Instructions[i] >> 24)
		if gotOp != op {
			t.Errorf("instr %d: expected op %v, got %v", i, op, gotOp)
		}
	}
}
