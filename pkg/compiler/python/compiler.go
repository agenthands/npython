package python

import (
	"fmt"
	"strings"
	"strconv"

	"github.com/go-python/gpython/ast"
	"github.com/go-python/gpython/parser"
	"github.com/go-python/gpython/py"
	
	"github.com/agenthands/npython/pkg/core/value"
	"github.com/agenthands/npython/pkg/vm"
)

var PythonBuiltins = map[string]uint32{
	"write_file":    0,
	"fetch":         1,
	"print":         2,
	"parse_json":    3,
	"format_string": 9,
	"is_empty":      10,
	"with_client":   11,
	"set_url":       12,
	"set_method":    13,
	"send_request":   5,
	"check_status":   6,
}

// Compiler transforms Python AST into nPython bytecode
type Compiler struct {
	instructions []uint32
	constants    []value.Value
	locals       map[string]int
	nextLocal    int
	arena        []byte
}

func NewCompiler() *Compiler {
	return &Compiler{
		locals:    make(map[string]int),
		constants: make([]value.Value, 0),
		arena:     make([]byte, 0),
	}
}

// Compile compiles Python source code into nPython bytecode
func (c *Compiler) Compile(src string) (*vm.Bytecode, error) {
	// Reset state
	c.instructions = nil
	c.constants = nil
	c.locals = make(map[string]int)
	c.nextLocal = 0
	c.arena = nil

	mod, err := parser.Parse(strings.NewReader(src), "<string>", py.ExecMode)
	if err != nil {
		return nil, fmt.Errorf("python parse error: %w", err)
	}

	module, ok := mod.(*ast.Module)
	if !ok {
		return nil, fmt.Errorf("expected *ast.Module, got %T", mod)
	}

	for _, stmt := range module.Body {
		if err := c.emitStmt(stmt); err != nil {
			return nil, err
		}
	}

	// Always end with HALT
	c.emitOp(vm.OP_HALT, 0)

	return &vm.Bytecode{
		Instructions: c.instructions,
		Constants:    c.constants,
		Arena:        c.arena,
	}, nil
}

func (c *Compiler) emitOp(op uint8, arg uint32) {
	instr := (uint32(op) << 24) | (arg & 0x00FFFFFF)
	c.instructions = append(c.instructions, instr)
}

func (c *Compiler) addConstant(v value.Value) uint32 {
	for i, existing := range c.constants {
		if existing.Type == v.Type && existing.Data == v.Data {
			return uint32(i)
		}
	}
	c.constants = append(c.constants, v)
	return uint32(len(c.constants) - 1)
}

func (c *Compiler) packNewString(s string) uint64 {
	offset := uint32(len(c.arena))
	length := uint32(len(s))
	c.arena = append(c.arena, []byte(s)...)
	return value.PackString(offset, length)
}

func (c *Compiler) getLocalIndex(name string) int {
	if idx, ok := c.locals[name]; ok {
		return idx
	}
	idx := c.nextLocal
	c.locals[name] = idx
	c.nextLocal++
	return idx
}

func (c *Compiler) emitStmt(stmt ast.Stmt) error {
	switch s := stmt.(type) {
	case *ast.Assign:
		if len(s.Targets) != 1 {
			return fmt.Errorf("only single assignment supported")
		}
		target, ok := s.Targets[0].(*ast.Name)
		if !ok {
			return fmt.Errorf("only variable assignment supported, got %T", s.Targets[0])
		}
		if err := c.emitExpr(s.Value); err != nil {
			return err
		}
		idx := c.getLocalIndex(string(target.Id))
		c.emitOp(vm.OP_POP_L, uint32(idx))
		return nil

	case *ast.ExprStmt:
		if err := c.emitExpr(s.Value); err != nil {
			return err
		}
		// Special case: if it's a builtin call that we know is void-ish or we want to ignore result
		if call, ok := s.Value.(*ast.Call); ok {
			if name, ok := call.Func.(*ast.Name); ok {
				if string(name.Id) == "print" || string(name.Id) == "write_file" {
					return nil
				}
			}
		}
		c.emitOp(vm.OP_DROP, 0)
		return nil

	case *ast.If:
		if err := c.emitExpr(s.Test); err != nil {
			return err
		}
		jumpFalseIdx := len(c.instructions)
		c.emitOp(vm.OP_JMP_FALSE, 0)
		for _, stmt := range s.Body {
			if err := c.emitStmt(stmt); err != nil {
				return err
			}
		}
		if len(s.Orelse) > 0 {
			jumpEndIdx := len(c.instructions)
			c.emitOp(vm.OP_JMP, 0)
			c.instructions[jumpFalseIdx] = (uint32(vm.OP_JMP_FALSE) << 24) | (uint32(len(c.instructions)) & 0x00FFFFFF)
			for _, stmt := range s.Orelse {
				if err := c.emitStmt(stmt); err != nil {
					return err
				}
			}
			c.instructions[jumpEndIdx] = (uint32(vm.OP_JMP) << 24) | (uint32(len(c.instructions)) & 0x00FFFFFF)
		} else {
			c.instructions[jumpFalseIdx] = (uint32(vm.OP_JMP_FALSE) << 24) | (uint32(len(c.instructions)) & 0x00FFFFFF)
		}
		return nil

	case *ast.While:
		loopStartIdx := len(c.instructions)
		if err := c.emitExpr(s.Test); err != nil {
			return err
		}
		jumpFalseIdx := len(c.instructions)
		c.emitOp(vm.OP_JMP_FALSE, 0)
		for _, stmt := range s.Body {
			if err := c.emitStmt(stmt); err != nil {
				return err
			}
		}
		c.emitOp(vm.OP_JMP, uint32(loopStartIdx))
		c.instructions[jumpFalseIdx] = (uint32(vm.OP_JMP_FALSE) << 24) | (uint32(len(c.instructions)) & 0x00FFFFFF)
		return nil

	case *ast.With:
		// Support 'with scope(name, token):'
		if len(s.Items) != 1 {
			return fmt.Errorf("only single with item supported")
		}
		item := s.Items[0]
		call, ok := item.ContextExpr.(*ast.Call)
		if !ok {
			return fmt.Errorf("with expects call to scope()")
		}
		name, ok := call.Func.(*ast.Name)
		if !ok || string(name.Id) != "scope" {
			return fmt.Errorf("with expects scope() context manager")
		}
		if len(call.Args) != 2 {
			return fmt.Errorf("scope() expects 2 arguments (name, token)")
		}

		// Push args and emit OP_ADDRESS
		if err := c.emitExpr(call.Args[0]); err != nil {
			return err
		}
		if err := c.emitExpr(call.Args[1]); err != nil {
			return err
		}
		c.emitOp(vm.OP_ADDRESS, 0)

		// Body
		for _, stmt := range s.Body {
			if err := c.emitStmt(stmt); err != nil {
				return err
			}
		}

		// Exit Scope
		c.emitOp(vm.OP_EXIT_ADDR, 0)
		return nil

	default:
		return fmt.Errorf("unsupported statement type: %T", stmt)
	}
}

func (c *Compiler) emitExpr(expr ast.Expr) error {
	switch e := expr.(type) {
	case *ast.Num:
		s := fmt.Sprintf("%v", e.N)
		i, err := strconv.ParseInt(s, 10, 64)
		if err == nil {
			val := value.Value{Type: value.TypeInt, Data: uint64(i)}
			idx := c.addConstant(val)
			c.emitOp(vm.OP_PUSH_C, idx)
			return nil
		}
		return fmt.Errorf("unsupported number literal: %v", e.N)

	case *ast.Str:
		valStr := string(e.S)
		val := value.Value{Type: value.TypeString, Data: c.packNewString(valStr)}
		idx := c.addConstant(val)
		c.emitOp(vm.OP_PUSH_C, idx)
		return nil

	case *ast.Name:
		idx := c.getLocalIndex(string(e.Id))
		c.emitOp(vm.OP_PUSH_L, uint32(idx))
		return nil

	case *ast.BinOp:
		if err := c.emitExpr(e.Left); err != nil {
			return err
		}
		if err := c.emitExpr(e.Right); err != nil {
			return err
		}
		switch e.Op {
		case ast.Add:
			c.emitOp(vm.OP_ADD, 0)
		case ast.Sub:
			c.emitOp(vm.OP_SUB, 0)
		case ast.Mult:
			c.emitOp(vm.OP_MUL, 0)
		case ast.Div:
			c.emitOp(vm.OP_DIV, 0)
		default:
			return fmt.Errorf("unsupported binary operator: %v", e.Op)
		}
		return nil

	case *ast.Compare:
		if len(e.Ops) != 1 {
			return fmt.Errorf("only single comparison supported")
		}
		if err := c.emitExpr(e.Left); err != nil {
			return err
		}
		if err := c.emitExpr(e.Comparators[0]); err != nil {
			return err
		}
		switch e.Ops[0] {
		case ast.Eq:
			c.emitOp(vm.OP_EQ, 0)
		case ast.NotEq:
			c.emitOp(vm.OP_NE, 0)
		case ast.Gt:
			c.emitOp(vm.OP_GT, 0)
		case ast.Lt:
			c.emitOp(vm.OP_LT, 0)
		default:
			return fmt.Errorf("unsupported comparison operator: %v", e.Ops[0])
		}
		return nil

	case *ast.Call:
		nameTok, ok := e.Func.(*ast.Name)
		if !ok {
			return fmt.Errorf("only direct function calls supported")
		}
		name := string(nameTok.Id)

		if sysIdx, ok := PythonBuiltins[name]; ok {
			// Emit arguments
			for _, arg := range e.Args {
				if err := c.emitExpr(arg); err != nil {
					return err
				}
			}
			c.emitOp(vm.OP_SYSCALL, sysIdx)
			
			// Handle return values for built-ins
			if name == "print" || name == "write_file" || name == "with_client" || name == "set_url" || name == "set_method" {
				val := value.Value{Type: value.TypeVoid, Data: 0}
				idx := c.addConstant(val)
				c.emitOp(vm.OP_PUSH_C, idx)
			}
			// Others (fetch, format_string, is_empty, parse_json) leave result on stack.
			return nil
		}
		return fmt.Errorf("unknown function: %s", name)

	default:
		return fmt.Errorf("unsupported expression type: %T", expr)
	}
}
