package tests

import (
	"testing"
	"github.com/agenthands/nforth/pkg/compiler/lexer"
	"github.com/agenthands/nforth/pkg/compiler/parser"
	"github.com/agenthands/nforth/pkg/compiler/emitter"
	"github.com/agenthands/nforth/pkg/vm"
)

func TestCompilerSafety(t *testing.T) {
	tests := []struct {
		name    string
		src     string
		wantErr bool
	}{
		{
			name:    "Floating State Failure",
			src:     "1 2 ADD",
			wantErr: true,
		},
		{
			name:    "Security Violation Failure",
			src:     `"google.com" FETCH INTO html`,
			wantErr: true,
		},
		{
			name:    "Valid Scoped Code",
			src:     `ADDRESS HTTP-ENV token "google.com" FETCH INTO html`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			src := []byte(tt.src)
			s := lexer.NewScanner(src)
			p := parser.NewParser(s, src)
			prog, err := p.Parse()
			if err != nil {
				if tt.wantErr {
					return // Expected error
				}
				t.Fatalf("Parse failed: %v", err)
			}

			if tt.wantErr {
				t.Errorf("expected error but got none")
				return
			}

			e := emitter.NewEmitter(src)
			_, err = e.Emit(prog)
			if err != nil {
				t.Fatalf("Emit failed: %v", err)
			}
		})
	}
}

func TestEndToEnd(t *testing.T) {
	src := []byte("10 20 ADD INTO result")
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

	m := &vm.Machine{}
	m.Code = bc.Instructions
	m.Constants = bc.Constants

	err = m.Run(100)
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	// Verify result
	res := m.Frames[0].Locals[0]
	if res.Data != 30 {
		t.Errorf("expected 30, got %d", res.Data)
	}
}
