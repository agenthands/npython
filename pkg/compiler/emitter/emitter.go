package emitter

import (
	"fmt"
	"strconv"
	"github.com/agenthands/nforth/pkg/compiler/ast"
	"github.com/agenthands/nforth/pkg/compiler/lexer"
	"github.com/agenthands/nforth/pkg/compiler/parser"
	"github.com/agenthands/nforth/pkg/core/value"
	"github.com/agenthands/nforth/pkg/vm"
)

type Bytecode struct {
	Instructions []uint32
	Constants    []value.Value
	Arena        []byte
}

type Emitter struct {
	instructions []uint32
	constants    []value.Value
	arena        []byte
	locals       map[string]int
	functions    map[string]int // Name -> Start IP
	src          []byte
}

func NewEmitter(src []byte) *Emitter {
	return &Emitter{
		locals:    make(map[string]int),
		functions: make(map[string]int),
		src:       src,
	}
}

func (e *Emitter) Emit(prog *ast.Program) (*Bytecode, error) {
	if prog != nil {
		for _, node := range prog.Nodes {
			if err := e.emitNode(node); err != nil {
				return nil, err
			}
		}
	}

	// Always end with HALT if not already present
	e.emitOp(vm.OP_HALT, 0)

	return &Bytecode{
		Instructions: e.instructions,
		Constants:    e.constants,
		Arena:        e.arena,
	}, nil
}

func (e *Emitter) emitNode(node ast.Node) error {
	if node == nil {
		return nil
	}
	switch n := node.(type) {
	case *ast.Definition:
		// 1. Record function entry
		name := string(e.src[n.Name.Offset : n.Name.Offset+n.Name.Length])
		
		// 2. Skip over function body during main execution
		jmpIdx := len(e.instructions)
		e.emitOp(vm.OP_JMP, 0)
		
		e.functions[name] = len(e.instructions)
		
		// 3. Reset locals for function scope
		oldLocals := e.locals
		e.locals = make(map[string]int)
		
		// Map parameters to locals 0, 1, 2...
		for i, argTok := range n.Args {
			argName := string(e.src[argTok.Offset : argTok.Offset+argTok.Length])
			e.locals[argName] = i
		}

		// Pop parameters from the stack into their locals
		// Since stack is [arg1, arg2], we pop in reverse order
		// BUT we MUST pop only if they are on the stack!
		// For a 1-arg function, i=0. Loop runs for i=0.
		for i := len(n.Args) - 1; i >= 0; i-- {
			e.emitOp(vm.OP_POP_L, uint32(i))
		}

		// 4. Emit Body
		for _, stmt := range n.Body {
			if err := e.emitNode(stmt); err != nil {
				return err
			}
		}
		
		e.emitOp(vm.OP_RET, 0)
		
		// 5. Restore locals and backpatch jump
		e.locals = oldLocals
		e.instructions[jmpIdx] = (uint32(vm.OP_JMP) << 24) | (uint32(len(e.instructions)) & 0x00FFFFFF)

	case *ast.Assignment:
		for _, expr := range n.Expression {
			if err := e.emitNode(expr); err != nil {
				return err
			}
		}
		
		// Target local variable
		name := string(e.src[n.Target.Offset : n.Target.Offset+n.Target.Length])
		idx, ok := e.locals[name]
		if !ok {
			idx = len(e.locals)
			e.locals[name] = idx
		}
		e.emitOp(vm.OP_POP_L, uint32(idx))

	case *ast.VoidOperation:
		for _, arg := range n.Args {
			if err := e.emitNode(arg); err != nil {
				return err
			}
		}
		// The identifier part of the void op is already handled if it was an Expr.
		// However, in our parser, parseExpr already emits for Identifiers.
		// So VoidOperation is mostly a semantic container for the validator's benefit.

	case *ast.NumberLiteral:
		valStr := string(e.src[n.Token.Offset : n.Token.Offset+n.Token.Length])
		val, _ := strconv.ParseInt(valStr, 10, 64)
		constIdx := e.addConstant(value.Value{Type: value.TypeInt, Data: uint64(val)})
		e.emitOp(vm.OP_PUSH_C, uint32(constIdx))

	case *ast.StringLiteral:
		valStr := string(e.src[n.Token.Offset : n.Token.Offset+n.Token.Length])
		// Strip quotes
		valStr = valStr[1 : len(valStr)-1]
		constIdx := e.addConstant(value.Value{Type: value.TypeString, Data: e.packNewString(valStr)})
		e.emitOp(vm.OP_PUSH_C, uint32(constIdx))

	case *ast.Identifier:
		name := string(e.src[n.Token.Offset : n.Token.Offset+n.Token.Length])
		
		// Check if it's a standard word
		if sig, ok := parser.StandardWords[name]; ok {
			if sig.RequiredScope != "" || name == "PRINT" {
				var hostIdx uint32
				switch name {
				case "WRITE-FILE":
					hostIdx = 0
				case "FETCH":
					hostIdx = 1
				case "PRINT":
					hostIdx = 2
				}
				e.emitOp(vm.OP_SYSCALL, hostIdx)
			} else {
				e.emitStandardWord(name)
			}
		} else if startIP, isFunc := e.functions[name]; isFunc {
			e.emitOp(vm.OP_CALL, uint32(startIP))
		} else {
			// Assume it's a local variable (includes params)
			idx, ok := e.locals[name]
			if !ok {
				return fmt.Errorf("undefined identifier: %s at line %d", name, n.Token.Line)
			}
			e.emitOp(vm.OP_PUSH_L, uint32(idx))
		}

	case *ast.IfStmt:
		// 1. Setup (e.g. "10 10")
		for _, s := range n.Setup {
			if err := e.emitNode(s); err != nil {
				return err
			}
		}

		// 2. Condition (e.g. "EQ")
		if err := e.emitNode(n.Condition); err != nil {
			return err
		}

		// 3. JMP_FALSE to ELSE (or END)
		jumpFalseIdx := len(e.instructions)
		e.emitOp(vm.OP_JMP_FALSE, 0)

		// 3. THEN block
		for _, stmt := range n.ThenBranch {
			if err := e.emitNode(stmt); err != nil {
				return err
			}
		}

		if len(n.ElseBranch) > 0 {
			// 4. JMP to END (skip ELSE)
			jumpEndIdx := len(e.instructions)
			e.emitOp(vm.OP_JMP, 0)

			// Backpatch JMP_FALSE to start of ELSE
			e.instructions[jumpFalseIdx] = (uint32(vm.OP_JMP_FALSE) << 24) | (uint32(len(e.instructions)) & 0x00FFFFFF)

			// 5. ELSE block
			for _, stmt := range n.ElseBranch {
				if err := e.emitNode(stmt); err != nil {
					return err
				}
			}

			// Backpatch JMP to END
			e.instructions[jumpEndIdx] = (uint32(vm.OP_JMP) << 24) | (uint32(len(e.instructions)) & 0x00FFFFFF)
		} else {
			// Backpatch JMP_FALSE to END
			e.instructions[jumpFalseIdx] = (uint32(vm.OP_JMP_FALSE) << 24) | (uint32(len(e.instructions)) & 0x00FFFFFF)
		}

	case *ast.WhileStmt:
		// 1. Loop Start (Target for jump back)
		loopStartIdx := len(e.instructions)

		// 2. Setup/Condition code (evaluates to bool)
		for _, s := range n.Setup {
			if err := e.emitNode(s); err != nil {
				return err
			}
		}

		// 3. JMP_FALSE to END (Exit loop if bool is 0)
		jumpFalseIdx := len(e.instructions)
		e.emitOp(vm.OP_JMP_FALSE, 0)

		// 4. Body
		for _, stmt := range n.Body {
			if err := e.emitNode(stmt); err != nil {
				return err
			}
		}

		// 5. JMP to START (Back to condition check)
		e.emitOp(vm.OP_JMP, uint32(loopStartIdx))

		// 6. Backpatch JMP_FALSE to END
		e.instructions[jumpFalseIdx] = (uint32(vm.OP_JMP_FALSE) << 24) | (uint32(len(e.instructions)) & 0x00FFFFFF)

	case *ast.SecurityGate:
		if n.IsExit {
			e.emitOp(vm.OP_EXIT_ADDR, 0)
			return nil
		}
		// Push Scope Name and Token to stack, then OP_ADDRESS
		envName := string(e.src[n.Env.Offset : n.Env.Offset+n.Env.Length])
						capToken := string(e.src[n.CapToken.Offset : n.CapToken.Offset+n.CapToken.Length])
						if n.CapToken.Kind == lexer.KindString {
							// Strip quotes if it was a string literal
							capToken = capToken[1 : len(capToken)-1]
						}
				
						// Add to constant pool and Arena
				
		
		scopeIdx := e.addConstant(value.Value{Type: value.TypeString, Data: e.packNewString(envName)})
		tokenIdx := e.addConstant(value.Value{Type: value.TypeString, Data: e.packNewString(capToken)})

		e.emitOp(vm.OP_PUSH_C, uint32(scopeIdx))
		e.emitOp(vm.OP_PUSH_C, uint32(tokenIdx))
		e.emitOp(vm.OP_ADDRESS, 0)

	default:
		// Skip other nodes for this phase
	}
	return nil
}

func (e *Emitter) emitOp(op uint8, arg uint32) {
	instr := (uint32(op) << 24) | (arg & 0x00FFFFFF)
	e.instructions = append(e.instructions, instr)
}

func (e *Emitter) addConstant(v value.Value) int {
	for i, c := range e.constants {
		if c.Type == v.Type && c.Data == v.Data {
			return i
		}
	}
	e.constants = append(e.constants, v)
	return len(e.constants) - 1
}

func (e *Emitter) packNewString(s string) uint64 {
	offset := uint32(len(e.arena))
	length := uint32(len(s))
	e.arena = append(e.arena, []byte(s)...)
	return value.PackString(offset, length)
}

func (e *Emitter) emitStandardWord(name string) {
	switch name {
	case "ADD":
		e.emitOp(vm.OP_ADD, 0)
	case "SUB":
		e.emitOp(vm.OP_SUB, 0)
	case "MUL":
		e.emitOp(vm.OP_MUL, 0)
	case "EQ":
		e.emitOp(vm.OP_EQ, 0)
	case "GT":
		e.emitOp(vm.OP_GT, 0)
	case "LT":
		e.emitOp(vm.OP_LT, 0)
	case "PRINT":
		e.emitOp(vm.OP_PRINT, 0)
	case "CONTAINS":
		e.emitOp(vm.OP_CONTAINS, 0)
	case "ERROR":
		e.emitOp(vm.OP_ERROR, 0)
	}
}
