package emitter

import (
	"fmt"
	"strconv"
	"github.com/agenthands/nforth/pkg/compiler/ast"
	"github.com/agenthands/nforth/pkg/compiler/parser"
	"github.com/agenthands/nforth/pkg/core/value"
	"github.com/agenthands/nforth/pkg/vm"
)

type Bytecode struct {
	Instructions []uint32
	Constants    []value.Value
}

type Emitter struct {
	instructions []uint32
	constants    []value.Value
	locals       map[string]int
	src          []byte
}

func NewEmitter(src []byte) *Emitter {
	return &Emitter{
		locals: make(map[string]int),
		src:    src,
	}
}

func (e *Emitter) Emit(prog *ast.Program) (*Bytecode, error) {
	for _, node := range prog.Nodes {
		if err := e.emitNode(node); err != nil {
			return nil, err
		}
	}

	// Always end with HALT if not already present
	e.emitOp(vm.OP_HALT, 0)

	return &Bytecode{
		Instructions: e.instructions,
		Constants:    e.constants,
	}, nil
}

func (e *Emitter) emitNode(node ast.Node) error {
	switch n := node.(type) {
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

	case *ast.NumberLiteral:
		valStr := string(e.src[n.Token.Offset : n.Token.Offset+n.Token.Length])
		val, _ := strconv.ParseInt(valStr, 10, 64)
		constIdx := e.addConstant(value.Value{Type: value.TypeInt, Data: uint64(val)})
		e.emitOp(vm.OP_PUSH_C, uint32(constIdx))

	case *ast.Identifier:
		name := string(e.src[n.Token.Offset : n.Token.Offset+n.Token.Length])
		
		// Check if it's a standard word
		if _, ok := parser.StandardWords[name]; ok {
			e.emitStandardWord(name)
		} else {
			// Assume it's a local variable
			idx, ok := e.locals[name]
			if !ok {
				return fmt.Errorf("undefined identifier: %s", name)
			}
			e.emitOp(vm.OP_PUSH_L, uint32(idx))
		}

	case *ast.SecurityGate:
		// OP_ADDRESS ScopeIdx
		// For now, we don't have a Scope Registry, so we'll push the name to constants
		envName := string(e.src[n.Env.Offset : n.Env.Offset+n.Env.Length])
		constIdx := e.addConstant(value.Value{Type: value.TypeString, Data: uint64(e.addStringConstant(envName))})
		e.emitOp(vm.OP_ADDRESS, uint32(constIdx))

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

func (e *Emitter) addStringConstant(s string) int {
	// This is a placeholder. In a real implementation, strings would be added to the Arena.
	// For now, we'll just return a fake offset.
	return 0 
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
	}
}
