package python

import (
	"testing"
)

func TestCompilerComprehensive(t *testing.T) {
	c := NewCompiler()

	t.Run("BasicArithmetic", func(t *testing.T) {
		src := "x = 1 + 2 * 3 - 4 / 2"
		bc, err := c.Compile(src)
		if err != nil { t.Fatal(err) }
		if len(bc.Instructions) < 5 { t.Errorf("too few instructions") }
	})

	t.Run("Comparisons", func(t *testing.T) {
		src := "x = 1 != 2; y = 1 > 0; z = 1 < 2"
		c.Compile(src)
	})

	t.Run("Unary", func(t *testing.T) {
		src := "x = -10"
		c.Compile(src)
	})

	t.Run("Return", func(t *testing.T) {
		src := "def f(): return 1; def g(): return"
		c.Compile(src)
	})

	t.Run("Errors", func(t *testing.T) {
		badSrcs := []string{
			"x, y = 1, 2", // Only single assignment
			"with 1: pass", // with expects call to scope()
			"with scope(1): pass", // scope() expects 2 args
			"def f(a): pass; f(1, 2)", // Function call arg mismatch? No, compiler doesn't check count yet.
			"unknown()", // unknown function
		}
		for _, s := range badSrcs {
			_, err := c.Compile(s)
			if err == nil { t.Errorf("expected error for %s", s) }
		}
	})

	t.Run("ControlFlow", func(t *testing.T) {
		src := `
if True:
    pass
while False:
    pass
`
		// 'pass' is not supported, I should check what ast.Pass emits
		// Actually, I'll use simple assignments
		src = `
if 1:
    x = 1
else:
    x = 0
while 0:
    x = 2
`
		_, err := c.Compile(src)
		if err != nil { t.Fatal(err) }
	})

	t.Run("Functions", func(t *testing.T) {
		src := `
def add(a, b):
    return a + b
res = add(1, 2)
`
		bc, err := c.Compile(src)
		if err != nil { t.Fatal(err) }
		if _, ok := bc.Functions["add"]; !ok { t.Errorf("function 'add' not registered") }
	})

	t.Run("ListsAndIndexing", func(t *testing.T) {
		src := `
items = [1, 2, 3]
x = items[0]
`
		_, err := c.Compile(src)
		if err != nil { t.Fatal(err) }
	})

	t.Run("WithScope", func(t *testing.T) {
		src := `
with scope("HTTP-ENV", "token"):
    fetch("url")
`
		_, err := c.Compile(src)
		if err != nil { t.Fatal(err) }
	})

	t.Run("Builtins", func(t *testing.T) {
		src := "l = len(range(10))"
		_, err := c.Compile(src)
		if err != nil { t.Fatal(err) }
	})

	t.Run("Deduplication", func(t *testing.T) {
		src := "s1 = 'hello'; s2 = 'hello'"
		bc, _ := c.Compile(src)
		if len(bc.Arena) != 5 { t.Errorf("expected arena size 5, got %d", len(bc.Arena)) }
	})

	t.Run("IfSimple", func(t *testing.T) {
		src := "if 1: x = 1"
		c.Compile(src)
	})

	t.Run("ExprStmt", func(t *testing.T) {
		src := "1 + 2"
		c.Compile(src)
	})

	t.Run("MoreErrors", func(t *testing.T) {
		badSrcs := []struct {
			src string
			msg string
		}{
			{"x = y = 1", "only single assignment"},
			{"x[0:1]", "only simple indexing"},
			{"global x", "unsupported statement type"},
		}
		for _, tt := range badSrcs {
			_, err := c.Compile(tt.src)
			if err == nil { t.Errorf("expected error for %s", tt.src) }
		}
	})
}
