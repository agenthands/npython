package vm

import (
	"errors"
	"runtime"
	"github.com/agenthands/nforth/pkg/core/value"
)

var (
	ErrStackOverflow  = errors.New("vm: stack overflow")
	ErrStackUnderflow = errors.New("vm: stack underflow")
	ErrGasExhausted   = errors.New("vm: gas exhausted")
)

const (
	StackDepth = 128
	MaxFrames  = 32
	MaxLocals  = 16
)

// Frame tracks local variables and return addresses for function calls.
type Frame struct {
	ReturnIP int
	Locals   [MaxLocals]value.Value
}

// Machine represents a single Agent's execution sandbox.
// It uses fixed-size arrays to ensure a predictable memory footprint.
type Machine struct {
	Stack [StackDepth]value.Value
	SP    int // Stack Pointer

	Frames [MaxFrames]Frame
	FP     int // Frame Pointer

	IP    int      // Instruction Pointer
	Code  []uint32 // Bytecode instructions
	
	Constants []value.Value // Constant pool
	
	// Arena for string data
	Arena []byte

	// Security Context
	ActiveScope string
}

// Reset clears the machine state for reuse (sync.Pool compliant).
func (m *Machine) Reset() {
	m.SP = 0
	m.IP = 0
	m.FP = 0
	m.ActiveScope = ""
	
	// Zero out the stack to avoid data leakage between agent runs
	for i := range m.Stack {
		m.Stack[i] = value.Value{}
	}
	
	// Zero out frames
	for i := range m.Frames {
		m.Frames[i] = Frame{}
	}
}

// Push adds a value to the stack. Panics on overflow.
func (m *Machine) Push(v value.Value) {
	if m.SP >= StackDepth {
		panic(ErrStackOverflow)
	}
	m.Stack[m.SP] = v
	m.SP++
}

// Pop removes and returns the top value from the stack. Panics on underflow.
func (m *Machine) Pop() value.Value {
	if m.SP <= 0 {
		panic(ErrStackUnderflow)
	}
	m.SP--
	return m.Stack[m.SP]
}

// Run executes instructions until HALT, error, or gas exhaustion.
func (m *Machine) Run(gasLimit int) (err error) {
	// 1. Safety net: Convert internal stack panics to errors
	defer func() {
		if r := recover(); r != nil {
			// Catch our custom stack errors
			if e, ok := r.(error); ok && (e == ErrStackOverflow || e == ErrStackUnderflow) {
				err = e
				return
			}
			
			// Catch Go runtime errors (like index out of range)
			if _, ok := r.(runtime.Error); ok {
				err = ErrStackUnderflow 
				return
			}
			
			panic(r)
		}
	}()

	// Cache hot fields in local variables for register allocation
	ip := m.IP
	sp := m.SP
	fp := m.FP
	code := m.Code

	for i := 0; i < gasLimit; i++ {
		instr := code[ip]
		op := uint8(instr >> 24)
		arg := instr & 0x00FFFFFF

		switch op {
		case OP_HALT:
			m.IP = ip
			m.SP = sp
			m.FP = fp
			return nil

		case OP_PUSH_C:
			m.Stack[sp] = m.Constants[arg]
			sp++
			ip++

		case OP_ADD:
			b := m.Stack[sp-1].Data
			a := m.Stack[sp-2].Data
			m.Stack[sp-2].Data = a + b
			sp--
			ip++

		case OP_POP_L:
			val := m.Stack[sp-1]
			m.Frames[fp].Locals[arg] = val
			sp--
			ip++

		default:
			m.IP = ip
			m.SP = sp
			m.FP = fp
			return errors.New("vm: unknown opcode")
		}
	}

	m.IP = ip
	m.SP = sp
	m.FP = fp
	return ErrGasExhausted
}
