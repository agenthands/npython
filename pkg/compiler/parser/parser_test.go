package parser_test

import (
	"testing"
	"github.com/agenthands/nforth/pkg/compiler/lexer"
	"github.com/agenthands/nforth/pkg/compiler/parser"
)

func TestDanglingStackEnforcer(t *testing.T) {
	tests := []struct {
		name    string
		src     string
		wantErr bool
	}{
		{
			name:    "Valid Assignment",
			src:     "1 2 ADD INTO result",
			wantErr: false,
		},
		{
			name:    "Invalid Floating State",
			src:     "1 2 ADD",
			wantErr: true,
		},
		{
			name:    "Invalid Underflow",
			src:     "ADD INTO result",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			src := []byte(tt.src)
			s := lexer.NewScanner(src)
			p := parser.NewParser(s, src)
			_, err := p.Parse()
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestScopeValidation(t *testing.T) {
	tests := []struct {
		name    string
		src     string
		wantErr bool
	}{
		{
			name:    "Security Violation: FETCH without scope",
			src:     "\"google.com\" FETCH INTO html",
			wantErr: true,
		},
		{
			name:    "Valid Scoped Access: FETCH with ADDRESS",
			src:     "ADDRESS HTTP-ENV token \"google.com\" FETCH INTO html",
			wantErr: false,
		},
		{
			name:    "Valid Nested Access: FETCH and WRITE",
			src:     "ADDRESS FS-ENV token1 ADDRESS HTTP-ENV token2 \"google.com\" FETCH INTO html \"file.txt\" html WRITE",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			src := []byte(tt.src)
			s := lexer.NewScanner(src)
			p := parser.NewParser(s, src)
			_, err := p.Parse()
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
