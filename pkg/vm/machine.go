package vm

import (
	"errors"
	"fmt"
	"runtime"
	"strings"
	"github.com/agenthands/npython/pkg/core/value"
)

var (
	ErrStackOverflow     = errors.New("vm: stack overflow")
	ErrStackUnderflow    = errors.New("vm: stack underflow")
	ErrGasExhausted      = errors.New("vm: gas exhausted")
	ErrSecurityViolation = errors.New("vm: security violation")
)

// Gatekeeper validates capability tokens for a given scope.
type Gatekeeper interface {
	Validate(scope, token string) bool
}

// HostFunction is a Go function registered to the VM.
type HostFunction func(m *Machine) error

// HostFunctionEntry tracks a host function and its required security scope.
type HostFunctionEntry struct {
	Fn            HostFunction
	RequiredScope string
}

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
	Gatekeeper   Gatekeeper
	ScopeStack   []string
	HostRegistry []HostFunctionEntry
}

// Reset clears the machine state for reuse (sync.Pool compliant).
func (m *Machine) Reset() {
	m.SP = 0
	m.IP = 0
	m.FP = 0
	m.ScopeStack = m.ScopeStack[:0]
	
	// Zero out the stack to avoid data leakage between agent runs
	for i := range m.Stack {
		m.Stack[i] = value.Value{}
	}
	
	// Zero out frames
	for i := range m.Frames {
		m.Frames[i] = Frame{}
	}
}

// RegisterHostFunction adds a host-side Go function to the VM's registry.
func (m *Machine) RegisterHostFunction(scope string, fn HostFunction) uint32 {
	m.HostRegistry = append(m.HostRegistry, HostFunctionEntry{
		Fn:            fn,
		RequiredScope: scope,
	})
	return uint32(len(m.HostRegistry) - 1)
}

// HasScope checks if a capability scope is currently active in the cumulative stack.
func (m *Machine) HasScope(scope string) bool {
	for _, s := range m.ScopeStack {
		if s == scope {
			return true
		}
	}
	return false
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
		// Mandatory state sync for syscalls/errors
		m.IP = ip
		m.SP = sp
		m.FP = fp

		instr := code[ip]
		op := uint8(instr >> 24)
		arg := instr & 0x00FFFFFF

		// Sync SP back before each instruction to allow SYSCALLs to see it?
		// No, usually we sync at SYSCALL.
		
		switch op {
		case OP_HALT:
			m.IP = ip
			m.SP = sp
			m.FP = fp
			return nil

		case OP_PUSH_C:
			m.Stack[sp] = m.Constants[arg]
			sp++
			m.SP = sp
			ip++

		case OP_ADD:
			b := int64(m.Stack[sp-1].Data)
			a := int64(m.Stack[sp-2].Data)
			m.Stack[sp-2].Data = uint64(a + b)
			sp--
			m.SP = sp
			ip++

		case OP_SUB:
			b := int64(m.Stack[sp-1].Data)
			a := int64(m.Stack[sp-2].Data)
			m.Stack[sp-2].Data = uint64(a - b)
			sp--
			m.SP = sp
			ip++

		case OP_MUL:
			b := int64(m.Stack[sp-1].Data)
			a := int64(m.Stack[sp-2].Data)
			m.Stack[sp-2].Data = uint64(a * b)
			sp--
			m.SP = sp
			ip++

		case OP_DIV:
			b := int64(m.Stack[sp-1].Data)
			a := int64(m.Stack[sp-2].Data)
			if b == 0 {
				m.IP = ip
				m.SP = sp
				m.FP = fp
				return errors.New("vm: division by zero")
			}
			m.Stack[sp-2].Data = uint64(a / b)
			sp--
			m.SP = sp
			ip++

		case OP_EQ:
			bVal := m.Stack[sp-1]
			aVal := m.Stack[sp-2]
			var res uint64
			if aVal.Type == value.TypeString && bVal.Type == value.TypeString {
				aStr := value.UnpackString(aVal.Data, m.Arena)
				bStr := value.UnpackString(bVal.Data, m.Arena)
				if aStr == bStr {
					res = 1
				}
			} else {
				if aVal.Data == bVal.Data {
					res = 1
				}
			}
			m.Stack[sp-2] = value.Value{Type: value.TypeBool, Data: res}
			sp--
			m.SP = sp
			ip++

		case OP_NE:
			bVal := m.Stack[sp-1]
			aVal := m.Stack[sp-2]
			var res uint64
			if aVal.Type == value.TypeString && bVal.Type == value.TypeString {
				aStr := value.UnpackString(aVal.Data, m.Arena)
				bStr := value.UnpackString(bVal.Data, m.Arena)
				if aStr != bStr {
					res = 1
				}
			} else {
				if aVal.Data != bVal.Data {
					res = 1
				}
			}
			m.Stack[sp-2] = value.Value{Type: value.TypeBool, Data: res}
			sp--
			m.SP = sp
			ip++

		case OP_GT:
			b := int64(m.Stack[sp-1].Data)
			a := int64(m.Stack[sp-2].Data)
			var res uint64
			if a > b {
				res = 1
			}
			m.Stack[sp-2] = value.Value{Type: value.TypeBool, Data: res}
			sp--
			m.SP = sp
			ip++

		case OP_LT:
			b := int64(m.Stack[sp-1].Data)
			a := int64(m.Stack[sp-2].Data)
			var res uint64
			if a < b {
				res = 1
			}
			m.Stack[sp-2] = value.Value{Type: value.TypeBool, Data: res}
			sp--
			m.SP = sp
			ip++

		case OP_DROP:
			sp--
			m.SP = sp
			ip++

		case OP_PRINT:
			// ( val -- )
			_ = m.Stack[sp-1]
			sp--
			m.SP = sp
			ip++

		case OP_CONTAINS:
			// ( str pattern -- bool )
			patternPacked := m.Stack[sp-1].Data
			strPacked := m.Stack[sp-2].Data
			
			pattern := value.UnpackString(patternPacked, m.Arena)
			str := value.UnpackString(strPacked, m.Arena)
			
			var res uint64
			if strings.Contains(str, pattern) {
				res = 1
			}
			m.Stack[sp-2] = value.Value{Type: value.TypeBool, Data: res}
			sp--
			m.SP = sp
			ip++

		case OP_FIND:
			// ( str pattern -- index )
			patternPacked := m.Stack[sp-1].Data
			strPacked := m.Stack[sp-2].Data
			
			pattern := value.UnpackString(patternPacked, m.Arena)
			str := value.UnpackString(strPacked, m.Arena)
			
			idx := strings.Index(str, pattern)
			m.Stack[sp-2] = value.Value{Type: value.TypeInt, Data: uint64(int64(idx))}
			sp--
			m.SP = sp
			ip++

		case OP_SLICE:
			// ( str start length -- sub )
			length := int64(m.Stack[sp-1].Data)
			start := int64(m.Stack[sp-2].Data)
			strVal := m.Stack[sp-3]
			
			if strVal.Type != value.TypeString {
				m.IP = ip
				m.SP = sp
				m.FP = fp
				return errors.New("vm: SLICE expects string")
			}

			origOffset := uint32(strVal.Data >> 32)
			origLength := uint32(strVal.Data)

			if start < 0 || length < 0 || uint32(start+length) > origLength {
				m.IP = ip
				m.SP = sp
				m.FP = fp
				return errors.New("vm: SLICE out of bounds")
			}

			newOffset := origOffset + uint32(start)
			newLength := uint32(length)
			
			m.Stack[sp-3] = value.Value{
				Type: value.TypeString,
				Data: value.PackString(newOffset, newLength),
			}
			sp -= 2
			m.SP = sp
			ip++

		case OP_LEN:
			// ( str -- len )
			strVal := m.Stack[sp-1]
			if strVal.Type != value.TypeString {
				m.IP = ip
				m.SP = sp
				m.FP = fp
				return errors.New("vm: LEN expects string")
			}
			length := uint32(strVal.Data)
			m.Stack[sp-1] = value.Value{Type: value.TypeInt, Data: uint64(length)}
			m.SP = sp
			ip++

		case OP_TRIM:
			// ( str -- str )
			strVal := m.Stack[sp-1]
			if strVal.Type != value.TypeString {
				m.IP = ip
				m.SP = sp
				m.FP = fp
				return errors.New("vm: TRIM expects string")
			}
			str := value.UnpackString(strVal.Data, m.Arena)
			trimmed := strings.TrimSpace(str)
			
			// If it's already trimmed, do nothing.
			// But wait, strings in npython are offset+length. 
			// Strings.TrimSpace might change start/length.
			// Actually, let's just push a NEW string into arena to be safe?
			// NO, we can just find the new offset/length in the existing arena if it's a subslice.
			// But strings.TrimSpace might return a copy if it's complex.
			// For now, let's just append to arena.
			
			offset := uint32(len(m.Arena))
			length := uint32(len(trimmed))
			m.Arena = append(m.Arena, []byte(trimmed)...)
			m.Stack[sp-1] = value.Value{Type: value.TypeString, Data: value.PackString(offset, length)}
			m.SP = sp
			ip++

		case OP_ERROR:
			// ( msg -- )
			msgPacked := m.Stack[sp-1].Data
			msg := value.UnpackString(msgPacked, m.Arena)
			m.IP = ip
			m.SP = sp
			m.FP = fp
			return errors.New("npython error: " + msg)

		case OP_PUSH_L:
			m.Stack[sp] = m.Frames[fp].Locals[arg]
			sp++
			m.SP = sp
			ip++

		case OP_POP_L:
			if sp <= 0 {
				m.IP = ip
				m.SP = sp
				m.FP = fp
				return fmt.Errorf("vm: stack underflow at POP_L index %d (IP: %d)", arg, ip)
			}
			val := m.Stack[sp-1]
			m.Frames[fp].Locals[arg] = val
			sp--
			m.SP = sp
			ip++

		case OP_JMP:
			ip = int(arg)

		case OP_JMP_FALSE:
			cond := m.Stack[sp-1].Data
			sp--
			m.SP = sp
			if cond == 0 {
				ip = int(arg)
			} else {
				ip++
			}

		case OP_CALL:
			// Save current state
			m.Frames[fp+1].ReturnIP = ip + 1
			fp++
			ip = int(arg)

		case OP_RET:
			ip = m.Frames[fp].ReturnIP
			fp--

		case OP_ADDRESS:
			// Stack: ( scope token -- )
			// token is at SP-1, scope is at SP-2
			tokenPacked := m.Stack[sp-1].Data
			scopePacked := m.Stack[sp-2].Data
			sp -= 2

			m.SP = sp // Sync before unpacking if needed, though UnpackString uses m.Arena
			token := value.UnpackString(tokenPacked, m.Arena)
			scope := value.UnpackString(scopePacked, m.Arena)

			if m.Gatekeeper == nil || !m.Gatekeeper.Validate(scope, token) {
				m.IP = ip
				m.SP = sp
				m.FP = fp
				return ErrSecurityViolation
			}

			m.ScopeStack = append(m.ScopeStack, scope)
			ip++

		case OP_EXIT_ADDR:
			if len(m.ScopeStack) > 0 {
				m.ScopeStack = m.ScopeStack[:len(m.ScopeStack)-1]
			}
			ip++

		case OP_SYSCALL:
			entry := m.HostRegistry[arg]
			if entry.RequiredScope != "" && !m.HasScope(entry.RequiredScope) {
				m.IP = ip
				m.SP = sp
				m.FP = fp
				return ErrSecurityViolation
			}

			// Sync Machine state before Syscall
			m.IP = ip
			m.SP = sp
			m.FP = fp

			if err := entry.Fn(m); err != nil {
				return err
			}

			// Restore state after Syscall (SP might have changed)
			sp = m.SP
			ip = m.IP
			ip++ // Advance past syscall

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
