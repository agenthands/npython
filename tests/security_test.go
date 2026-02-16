package tests

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"github.com/agenthands/nforth/pkg/vm"
	"github.com/agenthands/nforth/pkg/compiler/lexer"
	"github.com/agenthands/nforth/pkg/compiler/parser"
	"github.com/agenthands/nforth/pkg/compiler/emitter"
	"github.com/agenthands/nforth/pkg/stdlib"
)

type Gatekeeper struct {
	tokens map[string]string
}

func (g *Gatekeeper) Validate(scope, token string) bool {
	return g.tokens[scope] == token
}

func TestEndToEndSecurity(t *testing.T) {
	tempDir, _ := os.MkdirTemp("", "nforth-e2e-security")
	defer os.RemoveAll(tempDir)

	// Source: Scoped Write
	src := []byte(`
		ADDRESS FS-ENV "fs-token"
			"hello world" "output.txt" WRITE
		<EXIT>
	`)

	// 1. Compile
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

	// 2. Setup VM
	m := &vm.Machine{}
	m.Code = bc.Instructions
	m.Constants = bc.Constants
	m.Arena = bc.Arena

	m.Gatekeeper = &Gatekeeper{
		tokens: map[string]string{
			"FS-ENV": "fs-token",
		},
	}

	fsSandbox := stdlib.NewFSSandbox(tempDir, 1024)
	// Registration order must match Emitter's hardcoded IDs
	// WRITE = 0, FETCH = 1
	m.RegisterHostFunction("FS-ENV", fsSandbox.WriteFile)

	// 3. Run
	err = m.Run(100)
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	// 4. Verify
	data, err := os.ReadFile(filepath.Join(tempDir, "output.txt"))
	if err != nil {
		t.Fatalf("failed to read output file: %v", err)
	}
	if string(data) != "hello world" {
		t.Errorf("expected 'hello world', got '%s'", string(data))
	}
}

func TestSecurityViolation(t *testing.T) {
	// Source: Unscoped Write
	src := []byte(`"hello" "test.txt" WRITE`)

	s := lexer.NewScanner(src)
	p := parser.NewParser(s, src)
	
	// The "Safety First" Parser should catch this at compile time
	_, err := p.Parse()
	if err == nil {
		t.Fatal("expected compiler security violation, got nil")
	}
	
	if !strings.Contains(err.Error(), "Security Violation") {
		t.Errorf("expected security violation error message, got: %v", err)
	}
}
