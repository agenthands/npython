package lexer_test

import (
	"testing"
	"github.com/agenthands/nforth/pkg/compiler/lexer"
)

func TestScannerZeroAlloc(t *testing.T) {
	src := []byte(`<HTTP-GATE> FETCH THE url INTO html \ This is a comment`)
	s := lexer.NewScanner(src)

	// Measure allocations
	allocs := testing.AllocsPerRun(10, func() {
		s.Reset(src)
		for {
			tok := s.Next()
			if tok.Kind == lexer.KindEOF || tok.Kind == lexer.KindError {
				break
			}
		}
	})

	if allocs > 0 {
		t.Errorf("expected 0 allocations, got %f", allocs)
	}
}

func TestScannerSugarAndNoise(t *testing.T) {
	src := []byte(`<HTTP-GATE> FETCH THE url -> html`)
	s := lexer.NewScanner(src)

	expected := []lexer.Kind{
		lexer.KindSugarGate,
		lexer.KindIdentifier,
		lexer.KindNoise,
		lexer.KindIdentifier,
		lexer.KindInto,
		lexer.KindIdentifier,
		lexer.KindEOF,
	}

	for i, exp := range expected {
		tok := s.Next()
		if tok.Kind != exp {
			t.Errorf("token %d: expected kind %v, got %v", i, exp, tok.Kind)
		}
	}
}
