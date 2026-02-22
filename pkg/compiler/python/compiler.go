package python

import (
	"fmt"
	"math"
	"strconv"
	"strings"

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
	"send_request":  5,
	"check_status":  6,
	"len":           14,
	"range":         15,
	"list":          16,
	"sum":           17,
	"max":           18,
	"min":           19,
	"map":           20,
	"abs":           21,
	"bool":          22,
	"int":           23,
	"str":           24,
	"filter":        25,
	"pow":           26,
	"all":           27,
	"any":           28,
	"divmod":        32,
	"round":         33,
	"float":         34,
	"bin":           35,
	"oct":           36,
	"hex":           37,
	"chr":           38,
	"ord":           39,
	"dict":          40,
	"tuple":         41,
	"set":           42,
	"reversed":      43,
	"sorted":        44,
	"zip":           45,
	"enumerate":     46,
	"repr":          47,
	"ascii":         48,
	"hash":          49,
	"id":            50,
	"type":          51,
	"callable":      52,
	"iter":          53,
	"next":          54,
	"locals":        55,
	"globals":       56,
	"slice":         57,
	"bytes":         58,
	"bytearray":     59,
	"has_next":      60,
	"isinstance":    63,
}

type loopContext struct {
	startIP    uint32
	breakJumps []int
}

type Compiler struct {
	instructions  []uint32
	constants     []value.Value
	locals        map[string]int
	nextLocal     int
	arena         []byte
	stringOffsets map[string]uint32
	functions     map[string]*funcSignature
	loops         []*loopContext
}

type funcSignature struct {
	ip   int
	args []string
}

func NewCompiler() *Compiler {
	return &Compiler{
		locals:        make(map[string]int),
		constants:     make([]value.Value, 0),
		arena:         make([]byte, 0, 1024),
		stringOffsets: make(map[string]uint32),
		functions:     make(map[string]*funcSignature),
		loops:         make([]*loopContext, 0),
	}
}

func (c *Compiler) Compile(src string) (*vm.Bytecode, error) {
	c.instructions = c.instructions[:0]
	c.constants = c.constants[:0]
	c.locals = make(map[string]int)
	c.nextLocal = 0
	c.loops = c.loops[:0]
	c.arena = c.arena[:0]
	c.stringOffsets = make(map[string]uint32)
	c.functions = make(map[string]*funcSignature)

	mod, err := parser.Parse(strings.NewReader(src), "<string>", py.ExecMode)
	if err != nil {
		return nil, fmt.Errorf("python parse error: %w", err)
	}

	module, ok := mod.(*ast.Module)
	if !ok {
		return nil, fmt.Errorf("expected *ast.Module")
	}

	for _, stmt := range module.Body {
		if err := c.emitStmt(stmt); err != nil {
			return nil, err
		}
	}
	c.emitOp(vm.OP_HALT, 0)

	return &vm.Bytecode{
		Instructions: c.instructions,
		Constants:    c.constants,
		Arena:        c.arena,
		Functions:    c.exportFunctions(),
	}, nil
}

func (c *Compiler) exportFunctions() map[string]int {
	res := make(map[string]int)
	for k, v := range c.functions {
		res[k] = v.ip
	}
	return res
}

func (c *Compiler) emitOp(op uint8, arg uint32) {
	c.instructions = append(c.instructions, (uint32(op)<<24)|(arg&0x00FFFFFF))
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
	if offset, ok := c.stringOffsets[s]; ok {
		return value.PackString(offset, uint32(len(s)))
	}
	offset := uint32(len(c.arena))
	c.arena = append(c.arena, []byte(s)...)
	c.stringOffsets[s] = offset
	return value.PackString(offset, uint32(len(s)))
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
		switch target := s.Targets[0].(type) {
		case *ast.Name:
			if err := c.emitExpr(s.Value); err != nil {
				return err
			}
			c.emitOp(vm.OP_POP_L, uint32(c.getLocalIndex(string(target.Id))))
		case *ast.Tuple:
			// Unpack via temporary local
			tmpIdx := c.getLocalIndex("__tmp_unpack")
			if err := c.emitExpr(s.Value); err != nil {
				return err
			}
			c.emitOp(vm.OP_POP_L, uint32(tmpIdx))
			for i, el := range target.Elts {
				c.emitOp(vm.OP_PUSH_L, uint32(tmpIdx))
				c.emitOp(vm.OP_PUSH_C, c.addConstant(value.Value{Type: value.TypeInt, Data: uint64(i)}))
				c.emitOp(vm.OP_SYSCALL, 30) // get_item
				if name, ok := el.(*ast.Name); ok {
					c.emitOp(vm.OP_POP_L, uint32(c.getLocalIndex(string(name.Id))))
				} else if sub, ok := el.(*ast.Subscript); ok {
					if err := c.emitExpr(sub.Value); err != nil {
						return err
					}
					if err := c.emitExpr(sub.Slice.(*ast.Index).Value); err != nil {
						return err
					}
					c.emitOp(vm.OP_DUP, 2) // Dup the value from 2 down
					c.emitOp(vm.OP_SYSCALL, 31)
					c.emitOp(vm.OP_DROP, 0) // Drop the extra dup
				}
			}
		case *ast.Subscript:
			if err := c.emitExpr(target.Value); err != nil {
				return err
			}
			if err := c.emitExpr(target.Slice.(*ast.Index).Value); err != nil {
				return err
			}
			if err := c.emitExpr(s.Value); err != nil {
				return err
			}
			c.emitOp(vm.OP_SYSCALL, 31) // set_item
		}
	case *ast.AugAssign:
		idx := uint32(c.getLocalIndex(string(s.Target.(*ast.Name).Id)))
		c.emitOp(vm.OP_PUSH_L, idx)
		if err := c.emitExpr(s.Value); err != nil {
			return err
		}
		switch s.Op {
		case ast.Add:
			c.emitOp(vm.OP_ADD, 0)
		case ast.Sub:
			c.emitOp(vm.OP_SUB, 0)
		case ast.Mult:
			c.emitOp(vm.OP_MUL, 0)
		case ast.Div:
			c.emitOp(vm.OP_DIV, 0)
		}
		c.emitOp(vm.OP_POP_L, idx)
	case *ast.ExprStmt:
		if err := c.emitExpr(s.Value); err != nil {
			return err
		}
		c.emitOp(vm.OP_DROP, 0)
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
	case *ast.While:
		ctx := &loopContext{startIP: uint32(len(c.instructions))}
		c.loops = append(c.loops, ctx)
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
		c.emitOp(vm.OP_JMP, ctx.startIP)
		c.instructions[jumpFalseIdx] = (uint32(vm.OP_JMP_FALSE) << 24) | (uint32(len(c.instructions)) & 0x00FFFFFF)
		for _, idx := range ctx.breakJumps {
			c.instructions[idx] = (uint32(vm.OP_JMP) << 24) | (uint32(len(c.instructions)) & 0x00FFFFFF)
		}
		c.loops = c.loops[:len(c.loops)-1]
	case *ast.For:
		if err := c.emitExpr(s.Iter); err != nil {
			return err
		}
		c.emitOp(vm.OP_SYSCALL, 53) // iter()
		ctx := &loopContext{startIP: uint32(len(c.instructions))}
		c.loops = append(c.loops, ctx)
		c.emitOp(vm.OP_SYSCALL, 60) // has_next()
		jumpEndIdx := len(c.instructions)
		c.emitOp(vm.OP_JMP_FALSE, 0)
		c.emitOp(vm.OP_DUP, 0)
		c.emitOp(vm.OP_SYSCALL, 54) // next()
		switch target := s.Target.(type) {
		case *ast.Name:
			c.emitOp(vm.OP_POP_L, uint32(c.getLocalIndex(string(target.Id))))
		case *ast.Tuple:
			tmpIdx := c.getLocalIndex("__tmp_for")
			c.emitOp(vm.OP_POP_L, uint32(tmpIdx))
			for i, el := range target.Elts {
				c.emitOp(vm.OP_PUSH_L, uint32(tmpIdx))
				c.emitOp(vm.OP_PUSH_C, c.addConstant(value.Value{Type: value.TypeInt, Data: uint64(i)}))
				c.emitOp(vm.OP_SYSCALL, 30) // get_item
				if name, ok := el.(*ast.Name); ok {
					c.emitOp(vm.OP_POP_L, uint32(c.getLocalIndex(string(name.Id))))
				} else if sub, ok := el.(*ast.Subscript); ok {
					if err := c.emitExpr(sub.Value); err != nil {
						return err
					}
					if err := c.emitExpr(sub.Slice.(*ast.Index).Value); err != nil {
						return err
					}
					c.emitOp(vm.OP_DUP, 2)
					c.emitOp(vm.OP_SYSCALL, 31)
					c.emitOp(vm.OP_DROP, 0)
				}
			}
		}
		for _, stmt := range s.Body {
			if err := c.emitStmt(stmt); err != nil {
				return err
			}
		}
		c.emitOp(vm.OP_JMP, ctx.startIP)
		c.instructions[jumpEndIdx] = (uint32(vm.OP_JMP_FALSE) << 24) | (uint32(len(c.instructions)) & 0x00FFFFFF)
		for _, idx := range ctx.breakJumps {
			c.instructions[idx] = (uint32(vm.OP_JMP) << 24) | (uint32(len(c.instructions)) & 0x00FFFFFF)
		}
		c.emitOp(vm.OP_DROP, 0)
		c.loops = c.loops[:len(c.loops)-1]
	case *ast.Break:
		ctx := c.loops[len(c.loops)-1]
		ctx.breakJumps = append(ctx.breakJumps, len(c.instructions))
		c.emitOp(vm.OP_JMP, 0)
	case *ast.Continue:
		c.emitOp(vm.OP_JMP, c.loops[len(c.loops)-1].startIP)
	case *ast.With:
		call, ok := s.Items[0].ContextExpr.(*ast.Call)
		if !ok || len(call.Args) != 2 {
			return fmt.Errorf("with expects call to scope() with 2 arguments")
		}
		if err := c.emitExpr(call.Args[0]); err != nil {
			return err
		}
		if err := c.emitExpr(call.Args[1]); err != nil {
			return err
		}
		c.emitOp(vm.OP_ADDRESS, 0)
		for _, stmt := range s.Body {
			if err := c.emitStmt(stmt); err != nil {
				return err
			}
		}
		c.emitOp(vm.OP_EXIT_ADDR, 0)
	case *ast.FunctionDef:
		jmpIdx := len(c.instructions)
		c.emitOp(vm.OP_JMP, 0)
		args := make([]string, len(s.Args.Args))
		for i, a := range s.Args.Args {
			args[i] = string(a.Arg)
		}
		c.functions[string(s.Name)] = &funcSignature{ip: len(c.instructions), args: args}
		oldL, oldN := c.locals, c.nextLocal
		c.locals = make(map[string]int)
		c.nextLocal = len(s.Args.Args)
		for i, arg := range s.Args.Args {
			c.locals[string(arg.Arg)] = i
		}
		for _, stmt := range s.Body {
			if err := c.emitStmt(stmt); err != nil {
				return err
			}
		}
		c.emitOp(vm.OP_PUSH_C, c.addConstant(value.Value{Type: value.TypeVoid}))
		c.emitOp(vm.OP_RET, 0)
		c.locals, c.nextLocal = oldL, oldN
		c.instructions[jmpIdx] = (uint32(vm.OP_JMP) << 24) | (uint32(len(c.instructions)) & 0x00FFFFFF)
	case *ast.Return:
		if s.Value != nil {
			if err := c.emitExpr(s.Value); err != nil {
				return err
			}
		} else {
			c.emitOp(vm.OP_PUSH_C, c.addConstant(value.Value{Type: value.TypeVoid}))
		}
		c.emitOp(vm.OP_RET, 0)
	case *ast.Try:
		// nPython doesn't support exception catching natively in the VM yet.
		// To allow LLMs to write standard defensive Python, we simply compile and execute the happy-path Body inline.
		for _, stmt := range s.Body {
			if err := c.emitStmt(stmt); err != nil {
				return err
			}
		}
	default:
		return fmt.Errorf("unsupported statement type: %T", stmt)
	}
	return nil
}

func (c *Compiler) emitExpr(expr ast.Expr) error {
	switch e := expr.(type) {
	case *ast.Num:
		s := fmt.Sprintf("%v", e.N)
		if i, err := strconv.ParseInt(s, 10, 64); err == nil {
			c.emitOp(vm.OP_PUSH_C, c.addConstant(value.Value{Type: value.TypeInt, Data: uint64(i)}))
		} else if f, err := strconv.ParseFloat(s, 64); err == nil {
			c.emitOp(vm.OP_PUSH_C, c.addConstant(value.Value{Type: value.TypeFloat, Data: math.Float64bits(f)}))
		}
	case *ast.Str:
		c.emitOp(vm.OP_PUSH_C, c.addConstant(value.Value{Type: value.TypeString, Data: c.packNewString(string(e.S))}))
	case *ast.NameConstant:
		val := value.Value{Type: value.TypeVoid}
		if e.Value == py.True {
			val = value.Value{Type: value.TypeBool, Data: 1}
		} else if e.Value == py.False {
			val = value.Value{Type: value.TypeBool, Data: 0}
		}
		c.emitOp(vm.OP_PUSH_C, c.addConstant(val))
	case *ast.Name:
		name := string(e.Id)
		if sig, ok := c.functions[name]; ok {
			c.emitOp(vm.OP_PUSH_C, c.addConstant(value.Value{Type: value.TypeString, Data: c.packNewString(name)}))
			_ = sig // Unused but keep for symmetry
		} else {
			c.emitOp(vm.OP_PUSH_L, uint32(c.getLocalIndex(name)))
		}
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
		case ast.FloorDiv:
			c.emitOp(vm.OP_FLOOR_DIV, 0)
		case ast.Modulo:
			c.emitOp(vm.OP_MOD, 0)
		case ast.Pow:
			c.emitOp(vm.OP_POW, 0)
		case ast.BitAnd:
			c.emitOp(vm.OP_BIT_AND, 0)
		case ast.BitOr:
			c.emitOp(vm.OP_BIT_OR, 0)
		case ast.BitXor:
			c.emitOp(vm.OP_BIT_XOR, 0)
		case ast.LShift:
			c.emitOp(vm.OP_LSHIFT, 0)
		case ast.RShift:
			c.emitOp(vm.OP_RSHIFT, 0)
		}
	case *ast.BoolOp:
		for _, v := range e.Values {
			if err := c.emitExpr(v); err != nil {
				return err
			}
		}
		if e.Op == ast.And {
			c.emitOp(vm.OP_AND, 0)
		} else {
			c.emitOp(vm.OP_OR, 0)
		}
	case *ast.Compare:
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
		case ast.GtE:
			c.emitOp(vm.OP_GTE, 0)
		case ast.LtE:
			c.emitOp(vm.OP_LTE, 0)
		case ast.In:
			c.emitOp(vm.OP_IN, 0)
		case ast.NotIn:
			c.emitOp(vm.OP_NOT_IN, 0)
		case ast.Is:
			c.emitOp(vm.OP_EQ, 0) // Treat 'is' identically to '=='
		case ast.IsNot:
			c.emitOp(vm.OP_NE, 0) // Treat 'is not' identically to '!='
		}
	case *ast.Call:
		switch fn := e.Func.(type) {
		case *ast.Name:
			name := string(fn.Id)
			if sysIdx, ok := PythonBuiltins[name]; ok {
				for _, arg := range e.Args {
					if err := c.emitExpr(arg); err != nil {
						return err
					}
				}
				if name == "print" || name == "range" || name == "round" || name == "min" || name == "max" || name == "sum" {
					c.emitOp(vm.OP_PUSH_C, c.addConstant(value.Value{Type: value.TypeInt, Data: uint64(len(e.Args))}))
				}
				c.emitOp(vm.OP_SYSCALL, sysIdx)
				if name == "write_file" || name == "with_client" || name == "set_url" || name == "set_method" {
					c.emitOp(vm.OP_PUSH_C, c.addConstant(value.Value{Type: value.TypeVoid}))
				}
				return nil
			}
			if sig, ok := c.functions[name]; ok {
				if len(e.Keywords) == 0 {
					for _, arg := range e.Args {
						if err := c.emitExpr(arg); err != nil {
							return err
						}
					}
				} else {
					argMap := make(map[string]ast.Expr)
					for i, arg := range e.Args {
						if i < len(sig.args) {
							argMap[sig.args[i]] = arg
						}
					}
					for _, kw := range e.Keywords {
						argMap[string(kw.Arg)] = kw.Value
					}
					for _, argName := range sig.args {
						if _, ok := argMap[argName]; !ok {
							return fmt.Errorf("missing argument '%s'", argName)
						}
						if err := c.emitExpr(argMap[argName]); err != nil {
							return err
						}
					}
				}
				c.emitOp(vm.OP_CALL, (uint32(sig.ip)<<8)|(uint32(len(sig.args))&0xFF))
				return nil
			}
			return fmt.Errorf("unknown function '%s'", name)
		case *ast.Attribute:
			if err := c.emitExpr(fn.Value); err != nil {
				return err
			}
			for _, arg := range e.Args {
				if err := c.emitExpr(arg); err != nil {
					return err
				}
			}
			c.emitOp(vm.OP_PUSH_C, c.addConstant(value.Value{Type: value.TypeString, Data: c.packNewString(string(fn.Attr))}))
			c.emitOp(vm.OP_PUSH_C, c.addConstant(value.Value{Type: value.TypeInt, Data: uint64(len(e.Args))}))
			c.emitOp(vm.OP_SYSCALL, 62)
			return nil
		}
	case *ast.UnaryOp:
		c.emitOp(vm.OP_PUSH_C, c.addConstant(value.Value{Type: value.TypeInt, Data: 0}))
		c.emitExpr(e.Operand)
		c.emitOp(vm.OP_SUB, 0)
	case *ast.List:
		for _, el := range e.Elts {
			if err := c.emitExpr(el); err != nil {
				return err
			}
		}
		c.emitOp(vm.OP_PUSH_C, c.addConstant(value.Value{Type: value.TypeInt, Data: uint64(len(e.Elts))}))
		c.emitOp(vm.OP_SYSCALL, 29)
	case *ast.ListComp:
		c.emitOp(vm.OP_PUSH_C, c.addConstant(value.Value{Type: value.TypeInt, Data: 0}))
		c.emitOp(vm.OP_SYSCALL, 29)
		return c.emitComprehension(e.Elt, e.Generators)
	case *ast.DictComp:
		c.emitOp(vm.OP_SYSCALL, 40) // New Dict
		return c.emitDictComprehension(e.Key, e.Value, e.Generators)
	case *ast.GeneratorExp:
		c.emitOp(vm.OP_PUSH_C, c.addConstant(value.Value{Type: value.TypeInt, Data: 0}))
		c.emitOp(vm.OP_SYSCALL, 29)
		if err := c.emitComprehension(e.Elt, e.Generators); err != nil {
			return err
		}
		c.emitOp(vm.OP_SYSCALL, 53)
	case *ast.Tuple:
		for _, el := range e.Elts {
			if err := c.emitExpr(el); err != nil {
				return err
			}
		}
		c.emitOp(vm.OP_PUSH_C, c.addConstant(value.Value{Type: value.TypeInt, Data: uint64(len(e.Elts))}))
		c.emitOp(vm.OP_SYSCALL, 61)
	case *ast.Dict:
		c.emitOp(vm.OP_SYSCALL, 40)
		for i := range e.Keys {
			c.emitOp(vm.OP_DUP, 0)
			if err := c.emitExpr(e.Keys[i]); err != nil {
				return err
			}
			if err := c.emitExpr(e.Values[i]); err != nil {
				return err
			}
			c.emitOp(vm.OP_SYSCALL, 31)
		}
	case *ast.Attribute:
		if err := c.emitExpr(e.Value); err != nil {
			return err
		}
		c.emitOp(vm.OP_PUSH_C, c.addConstant(value.Value{Type: value.TypeString, Data: c.packNewString(string(e.Attr))}))
		c.emitOp(vm.OP_SYSCALL, 4)
		return nil
	case *ast.Subscript:
		c.emitExpr(e.Value)
		switch sl := e.Slice.(type) {
		case *ast.Index:
			c.emitExpr(sl.Value)
			c.emitOp(vm.OP_SYSCALL, 30)
		case *ast.Slice:
			if sl.Lower != nil {
				c.emitExpr(sl.Lower)
			} else {
				c.emitOp(vm.OP_PUSH_C, c.addConstant(value.Value{Type: value.TypeVoid}))
			}
			if sl.Upper != nil {
				c.emitExpr(sl.Upper)
			} else {
				c.emitOp(vm.OP_PUSH_C, c.addConstant(value.Value{Type: value.TypeVoid}))
			}
			if sl.Step != nil {
				c.emitExpr(sl.Step)
			} else {
				c.emitOp(vm.OP_PUSH_C, c.addConstant(value.Value{Type: value.TypeVoid}))
			}
			c.emitOp(vm.OP_SYSCALL, 57)
		}
	case *ast.Lambda:
		name := fmt.Sprintf("__lambda_%d", len(c.functions))
		jmp := len(c.instructions)
		c.emitOp(vm.OP_JMP, 0)
		args := make([]string, len(e.Args.Args))
		for i, a := range e.Args.Args {
			args[i] = string(a.Arg)
		}
		c.functions[name] = &funcSignature{ip: len(c.instructions), args: args}
		oldL, oldN := c.locals, c.nextLocal
		c.locals = make(map[string]int)
		c.nextLocal = len(e.Args.Args)
		for i, a := range e.Args.Args {
			c.locals[string(a.Arg)] = i
		}
		if err := c.emitExpr(e.Body); err != nil {
			return err
		}
		c.emitOp(vm.OP_RET, 0)
		c.locals, c.nextLocal = oldL, oldN
		c.instructions[jmp] = (uint32(vm.OP_JMP) << 24) | (uint32(len(c.instructions)) & 0x00FFFFFF)
		c.emitOp(vm.OP_PUSH_C, c.addConstant(value.Value{Type: value.TypeString, Data: c.packNewString(name)}))
	default:
		return fmt.Errorf("unsupported expression type: %T", expr)
	}
	return nil
}

func (c *Compiler) emitComprehension(elt ast.Expr, generators []ast.Comprehension) error {
	gen := generators[0]
	if err := c.emitExpr(gen.Iter); err != nil {
		return err
	}
	c.emitOp(vm.OP_SYSCALL, 53)
	start := uint32(len(c.instructions))
	c.emitOp(vm.OP_SYSCALL, 60)
	end := len(c.instructions)
	c.emitOp(vm.OP_JMP_FALSE, 0)
	c.emitOp(vm.OP_DUP, 0)
	c.emitOp(vm.OP_SYSCALL, 54)
	c.emitOp(vm.OP_POP_L, uint32(c.getLocalIndex(string(gen.Target.(*ast.Name).Id))))
	var ifs []int
	for _, cond := range gen.Ifs {
		if err := c.emitExpr(cond); err != nil {
			return err
		}
		ifs = append(ifs, len(c.instructions))
		c.emitOp(vm.OP_JMP_FALSE, 0)
	}
	c.emitOp(vm.OP_DUP, 1)
	if err := c.emitExpr(elt); err != nil {
		return err
	}
	c.emitOp(vm.OP_PUSH_C, c.addConstant(value.Value{Type: value.TypeString, Data: c.packNewString("append")}))
	c.emitOp(vm.OP_PUSH_C, c.addConstant(value.Value{Type: value.TypeInt, Data: 1}))
	c.emitOp(vm.OP_SYSCALL, 62)
	c.emitOp(vm.OP_DROP, 0)
	for _, idx := range ifs {
		c.instructions[idx] = (uint32(vm.OP_JMP_FALSE) << 24) | (uint32(len(c.instructions)) & 0x00FFFFFF)
	}
	c.emitOp(vm.OP_JMP, start)
	c.instructions[end] = (uint32(vm.OP_JMP_FALSE) << 24) | (uint32(len(c.instructions)) & 0x00FFFFFF)
	c.emitOp(vm.OP_DROP, 0)
	return nil
}

func (c *Compiler) emitDictComprehension(key, val ast.Expr, generators []ast.Comprehension) error {
	gen := generators[0]
	if err := c.emitExpr(gen.Iter); err != nil {
		return err
	}
	c.emitOp(vm.OP_SYSCALL, 53)
	start := uint32(len(c.instructions))
	c.emitOp(vm.OP_SYSCALL, 60)
	end := len(c.instructions)
	c.emitOp(vm.OP_JMP_FALSE, 0)
	c.emitOp(vm.OP_DUP, 0)
	c.emitOp(vm.OP_SYSCALL, 54)
	c.emitOp(vm.OP_POP_L, uint32(c.getLocalIndex(string(gen.Target.(*ast.Name).Id))))
	var ifs []int
	for _, cond := range gen.Ifs {
		if err := c.emitExpr(cond); err != nil {
			return err
		}
		ifs = append(ifs, len(c.instructions))
		c.emitOp(vm.OP_JMP_FALSE, 0)
	}
	c.emitOp(vm.OP_DUP, 1) // Dup the Dict
	if err := c.emitExpr(key); err != nil {
		return err
	}
	if err := c.emitExpr(val); err != nil {
		return err
	}
	c.emitOp(vm.OP_SYSCALL, 31) // SetItem
	for _, idx := range ifs {
		c.instructions[idx] = (uint32(vm.OP_JMP_FALSE) << 24) | (uint32(len(c.instructions)) & 0x00FFFFFF)
	}
	c.emitOp(vm.OP_JMP, start)
	c.instructions[end] = (uint32(vm.OP_JMP_FALSE) << 24) | (uint32(len(c.instructions)) & 0x00FFFFFF)
	c.emitOp(vm.OP_DROP, 0)
	return nil
}
