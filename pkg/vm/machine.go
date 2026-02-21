package vm

import (
	"errors"
	"fmt"
	"math"
	"strings"
	"sync"
	"github.com/agenthands/npython/pkg/core/value"
)

const (
	StackDepth = 128
	MaxFrames  = 32
	MaxLocals  = 16
)

var (
	ErrStackOverflow    = errors.New("vm: stack overflow")
	ErrStackUnderflow   = errors.New("vm: stack underflow")
	ErrFrameOverflow    = errors.New("vm: call stack overflow")
	ErrGasExhausted     = errors.New("vm: gas exhausted")
	ErrSecurityViolation = errors.New("vm: security violation")
)

type Frame struct {
	ReturnIP   int
	BaseSP     int 
	ArgCount   int
	Locals     [MaxLocals]value.Value
	LocalNames [MaxLocals]string
}

type Machine struct {
	Stack      [StackDepth]value.Value
	SP         int
	IP         int
	FP         int
	Frames     [MaxFrames]Frame
	Code       []uint32
	Constants  []value.Value
	Arena      []byte
	Gatekeeper Gatekeeper
	TokenMap   map[string]string
	ScopeStack []string
	FunctionRegistry map[string]int
	HostRegistry     []HostFunctionEntry
}

type Gatekeeper interface {
	Validate(scope, token string) bool
}

type HostFunctionEntry struct {
	RequiredScope string
	Fn            func(*Machine) error
}

var machinePool = sync.Pool{
	New: func() any {
		return &Machine{
			TokenMap:         make(map[string]string),
			ScopeStack:       make([]string, 0, 8),
			FunctionRegistry: make(map[string]int),
		}
	},
}

func GetMachine() *Machine {
	return machinePool.Get().(*Machine)
}

func PutMachine(m *Machine) {
	m.Reset()
	machinePool.Put(m)
}

func (m *Machine) Reset() {
	m.SP = 0
	m.IP = 0
	m.FP = 0
	for i := range m.Frames { m.Frames[i] = Frame{} }
	m.ScopeStack = m.ScopeStack[:0]
	for k := range m.TokenMap { delete(m.TokenMap, k) }
}

func (m *Machine) Push(v value.Value) {
	m.Stack[m.SP] = v
	m.SP++
}

func (m *Machine) Pop() value.Value {
	m.SP--
	return m.Stack[m.SP]
}

func (m *Machine) Peek() value.Value {
	return m.Stack[m.SP-1]
}

func (m *Machine) RegisterHostFunction(scope string, fn func(*Machine) error) {
	m.HostRegistry = append(m.HostRegistry, HostFunctionEntry{
		RequiredScope: scope,
		Fn:            fn,
	})
}

func (m *Machine) HasScope(scope string) bool {
	for _, s := range m.ScopeStack { if s == scope { return true } }
	return false
}

func (m *Machine) Call(ip int, args ...value.Value) (value.Value, error) {
	m.FP++
	f := &m.Frames[m.FP]
	f.ReturnIP = -1
	f.BaseSP = m.SP
	f.ArgCount = len(args)
	for i, a := range args { f.Locals[i] = a }
	m.IP = ip
	err := m.Run(1000000)
	if err != nil && err.Error() != "vm: stop marker" { return value.Value{}, err }
	return m.Pop(), nil
}

func isTruthy(v value.Value) bool {
	switch v.Type {
	case value.TypeBool: return v.Data != 0
	case value.TypeInt: return v.Data != 0
	case value.TypeFloat: return math.Float64frombits(v.Data) != 0
	case value.TypeString: return uint32(v.Data) != 0
	case value.TypeList, value.TypeTuple:
		var list []value.Value
		if v.Type == value.TypeList { if lp, ok := v.Opaque.(*[]value.Value); ok { list = *lp } } else { if l, ok := v.Opaque.([]value.Value); ok { list = l } }
		return len(list) > 0
	case value.TypeDict: if d, ok := v.Opaque.(map[string]any); ok { return len(d) > 0 }; return false
	case value.TypeIterator: return true
	}
	return false
}

func contains(m *Machine, container, item value.Value) bool {
	switch container.Type {
	case value.TypeString: return strings.Contains(value.UnpackString(container.Data, m.Arena), value.UnpackString(item.Data, m.Arena))
	case value.TypeList:
		l := *(container.Opaque.(*[]value.Value))
		for _, v := range l { if v.Type == item.Type && v.Data == item.Data { return true } }
	case value.TypeDict:
		d := container.Opaque.(map[string]any)
		_, ok := d[value.UnpackString(item.Data, m.Arena)]
		return ok
	}
	return false
}

func (m *Machine) Run(gasLimit int) (err error) {
	var op uint8
	defer func() {
		if r := recover(); r != nil {
			if e, ok := r.(error); ok && (e == ErrStackUnderflow) { err = fmt.Errorf("vm: underflow at OP_%02X (IP: %d)", op, m.IP) } else { panic(r) }
		}
	}()

	for i := 0; i < gasLimit; i++ {
		instr := m.Code[m.IP]
		op = uint8(instr >> 24)
		arg := int(instr & 0x00FFFFFF)

		switch op {
		case OP_HALT: return nil
		case OP_PUSH_C: m.Push(m.Constants[arg]); m.IP++
		case OP_ADD:
			b := m.Pop(); a := m.Pop()
			if a.Type == value.TypeString {
				res := value.UnpackString(a.Data, m.Arena) + b.Format(m.Arena)
				off := uint32(len(m.Arena)); m.Arena = append(m.Arena, []byte(res)...)
				m.Push(value.Value{Type: value.TypeString, Data: value.PackString(off, uint32(len(res)))})
			} else { m.Push(value.Value{Type: value.TypeInt, Data: uint64(int64(a.Data) + int64(b.Data))}) }
			m.IP++
		case OP_SUB: b := m.Pop().Int(); a := m.Pop().Int(); m.Push(value.Value{Type: value.TypeInt, Data: uint64(a - b)}); m.IP++
		case OP_MUL: b := m.Pop().Int(); a := m.Pop().Int(); m.Push(value.Value{Type: value.TypeInt, Data: uint64(a * b)}); m.IP++
		case OP_DIV: b := m.Pop().Int(); a := m.Pop().Int(); if b == 0 { return errors.New("vm: div0") }; m.Push(value.Value{Type: value.TypeInt, Data: uint64(a / b)}); m.IP++
		case OP_MOD:
			b := m.Pop(); a := m.Pop()
			if a.Type == value.TypeString {
				res := strings.Replace(value.UnpackString(a.Data, m.Arena), "%s", b.Format(m.Arena), 1)
				off := uint32(len(m.Arena)); m.Arena = append(m.Arena, []byte(res)...)
				m.Push(value.Value{Type: value.TypeString, Data: value.PackString(off, uint32(len(res)))})
			} else { if b.Data == 0 { return errors.New("vm: div0") }; m.Push(value.Value{Type: value.TypeInt, Data: uint64(int64(a.Data) % int64(b.Data))}) }
			m.IP++
		case OP_EQ: b := m.Pop(); a := m.Pop(); r := uint64(0); if a.Type == b.Type && a.Data == b.Data { r = 1 }; m.Push(value.Value{Type: value.TypeBool, Data: r}); m.IP++
		case OP_NE: b := m.Pop(); a := m.Pop(); r := uint64(0); if a.Type != b.Type || a.Data != b.Data { r = 1 }; m.Push(value.Value{Type: value.TypeBool, Data: r}); m.IP++
		case OP_GT: b := m.Pop().Int(); a := m.Pop().Int(); r := uint64(0); if a > b { r = 1 }; m.Push(value.Value{Type: value.TypeBool, Data: r}); m.IP++
		case OP_LT: b := m.Pop().Int(); a := m.Pop().Int(); r := uint64(0); if a < b { r = 1 }; m.Push(value.Value{Type: value.TypeBool, Data: r}); m.IP++
		case OP_LTE: b := m.Pop().Int(); a := m.Pop().Int(); r := uint64(0); if a <= b { r = 1 }; m.Push(value.Value{Type: value.TypeBool, Data: r}); m.IP++
		case OP_GTE: b := m.Pop().Int(); a := m.Pop().Int(); r := uint64(0); if a >= b { r = 1 }; m.Push(value.Value{Type: value.TypeBool, Data: r}); m.IP++
		case OP_POW:
			bVal := m.Pop(); aVal := m.Pop()
			var f1, f2 float64
			if aVal.Type == value.TypeInt { f1 = float64(aVal.Int()) } else { f1 = math.Float64frombits(aVal.Data) }
			if bVal.Type == value.TypeInt { f2 = float64(bVal.Int()) } else { f2 = math.Float64frombits(bVal.Data) }
			m.Push(value.Value{Type: value.TypeFloat, Data: math.Float64bits(math.Pow(f1, f2))}); m.IP++
		case OP_AND: b := m.Pop(); a := m.Pop(); r := uint64(0); if isTruthy(a) && isTruthy(b) { r = 1 }; m.Push(value.Value{Type: value.TypeBool, Data: r}); m.IP++
		case OP_OR: b := m.Pop(); a := m.Pop(); r := uint64(0); if isTruthy(a) || isTruthy(b) { r = 1 }; m.Push(value.Value{Type: value.TypeBool, Data: r}); m.IP++
		case OP_IN: c := m.Pop(); i := m.Pop(); r := uint64(0); if contains(m, c, i) { r = 1 }; m.Push(value.Value{Type: value.TypeBool, Data: r}); m.IP++
		case OP_NOT_IN: c := m.Pop(); i := m.Pop(); r := uint64(0); if !contains(m, c, i) { r = 1 }; m.Push(value.Value{Type: value.TypeBool, Data: r}); m.IP++
		case OP_DROP: m.Pop(); m.IP++
		case OP_DUP: m.Push(m.Stack[m.SP-1-int(arg)]); m.IP++
		case OP_JMP: m.IP = arg
		case OP_JMP_FALSE: if !isTruthy(m.Pop()) { m.IP = arg } else { m.IP++ }
		case OP_PUSH_L: m.Push(m.Frames[m.FP].Locals[arg]); m.IP++
		case OP_POP_L: m.Frames[m.FP].Locals[arg] = m.Pop(); m.IP++
		case OP_CALL:
			target, argc := arg >> 8, arg & 0xFF
			m.Frames[m.FP+1].ReturnIP, m.Frames[m.FP+1].ArgCount, m.Frames[m.FP+1].BaseSP = m.IP+1, argc, m.SP-argc
			for j := 0; j < argc; j++ { m.Frames[m.FP+1].Locals[j] = m.Stack[m.SP-argc+j] }
			m.SP -= argc; m.FP++; m.IP = target
		case OP_RET:
			res := m.Pop()
			if m.Frames[m.FP].ReturnIP == -1 { m.SP = m.Frames[m.FP].BaseSP; m.Push(res); m.IP = -1; m.FP--; return errors.New("vm: stop marker") }
			m.IP = m.Frames[m.FP].ReturnIP; m.SP = m.Frames[m.FP].BaseSP; m.FP--; m.Push(res)
		case OP_ADDRESS:
			tVal := m.Pop(); sVal := m.Pop()
			s := value.UnpackString(sVal.Data, m.Arena); t := value.UnpackString(tVal.Data, m.Arena)
			if m.Gatekeeper == nil || !m.Gatekeeper.Validate(s, t) { return ErrSecurityViolation }
			m.ScopeStack = append(m.ScopeStack, s); m.TokenMap[s] = t; m.IP++
		case OP_EXIT_ADDR:
			if len(m.ScopeStack) > 0 { m.ScopeStack = m.ScopeStack[:len(m.ScopeStack)-1] }
			m.IP++
		case OP_SYSCALL:
			entry := m.HostRegistry[arg]
			if entry.RequiredScope != "" && !m.HasScope(entry.RequiredScope) { return ErrSecurityViolation }
			if err := entry.Fn(m); err != nil { return err }
			m.IP++
		default: return fmt.Errorf("vm: unknown op %02X", op)
		}
	}
	return ErrGasExhausted
}
