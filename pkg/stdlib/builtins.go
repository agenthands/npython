package stdlib

import (
	"errors"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"github.com/agenthands/npython/pkg/core/value"
	"github.com/agenthands/npython/pkg/vm"
)

func pushString(m *vm.Machine, s string) error {
	offset := uint32(len(m.Arena))
	length := uint32(len(s))
	m.Arena = append(m.Arena, []byte(s)...)
	m.Push(value.Value{Type: value.TypeString, Data: value.PackString(offset, length)})
	return nil
}

// DivMod: ( a b -- tuple(q, r) )
func DivMod(m *vm.Machine) error {
	b := m.Pop().Int(); a := m.Pop().Int()
	if b == 0 { return errors.New("div0") }
	m.Push(value.Value{Type: value.TypeTuple, Opaque: []value.Value{
		{Type: value.TypeInt, Data: uint64(a / b)},
		{Type: value.TypeInt, Data: uint64(a % b)},
	}})
	return nil
}

func Round(m *vm.Machine) error {
	v := m.Pop()
	var f float64
	if v.Type == value.TypeInt { f = float64(v.Int()) } else { f = math.Float64frombits(v.Data) }
	m.Push(value.Value{Type: value.TypeInt, Data: uint64(int64(math.Round(f)))})
	return nil
}

func Float(m *vm.Machine) error {
	v := m.Pop()
	var f float64
	switch v.Type {
	case value.TypeInt: f = float64(v.Int())
	case value.TypeFloat: f = math.Float64frombits(v.Data)
	case value.TypeString: f, _ = strconv.ParseFloat(value.UnpackString(v.Data, m.Arena), 64)
	}
	m.Push(value.Value{Type: value.TypeFloat, Data: math.Float64bits(f)})
	return nil
}

func Bin(m *vm.Machine) error { return pushString(m, "0b"+strconv.FormatInt(m.Pop().Int(), 2)) }
func Oct(m *vm.Machine) error { return pushString(m, "0o"+strconv.FormatInt(m.Pop().Int(), 8)) }
func Hex(m *vm.Machine) error { return pushString(m, "0x"+strconv.FormatInt(m.Pop().Int(), 16)) }
func Chr(m *vm.Machine) error { return pushString(m, string(rune(m.Pop().Int()))) }
func Ord(m *vm.Machine) error {
	s := value.UnpackString(m.Pop().Data, m.Arena)
	if len(s) == 0 { return errors.New("empty ord") }
	m.Push(value.Value{Type: value.TypeInt, Data: uint64(s[0])})
	return nil
}

func Dict(m *vm.Machine) error {
	m.Push(value.Value{Type: value.TypeDict, Opaque: make(map[string]any)})
	return nil
}

func MakeTuple(m *vm.Machine) error {
	n := int(m.Pop().Int())
	l := make([]value.Value, n)
	for i := n - 1; i >= 0; i-- { l[i] = m.Pop() }
	m.Push(value.Value{Type: value.TypeTuple, Opaque: l})
	return nil
}

func Tuple(m *vm.Machine) error {
	v := m.Pop()
	var list []value.Value
	if v.Type == value.TypeList { list = *(v.Opaque.(*[]value.Value)) } else if v.Type == value.TypeTuple { list = v.Opaque.([]value.Value) }
	m.Push(value.Value{Type: value.TypeTuple, Opaque: list})
	return nil
}

func Set(m *vm.Machine) error {
	v := m.Pop()
	var list []value.Value
	if v.Type == value.TypeList { list = *(v.Opaque.(*[]value.Value)) } else { list = v.Opaque.([]value.Value) }
	s := make(map[any]struct{})
	for _, x := range list { s[x.Data] = struct{}{} }
	m.Push(value.Value{Type: value.TypeSet, Opaque: s})
	return nil
}

func Reversed(m *vm.Machine) error {
	l := *(m.Pop().Opaque.(*[]value.Value))
	res := make([]value.Value, len(l))
	for i, j := 0, len(l)-1; i < len(l); i, j = i+1, j-1 { res[i] = l[j] }
	ptr := new([]value.Value); *ptr = res
	m.Push(value.Value{Type: value.TypeList, Opaque: ptr})
	return nil
}

func Sorted(m *vm.Machine) error {
	l := *(m.Pop().Opaque.(*[]value.Value))
	res := make([]value.Value, len(l)); copy(res, l)
	sort.Slice(res, func(i, j int) bool { return res[i].Data < res[j].Data })
	ptr := new([]value.Value); *ptr = res
	m.Push(value.Value{Type: value.TypeList, Opaque: ptr})
	return nil
}

func Zip(m *vm.Machine) error {
	l2 := *(m.Pop().Opaque.(*[]value.Value)); l1 := *(m.Pop().Opaque.(*[]value.Value))
	min := len(l1); if len(l2) < min { min = len(l2) }
	res := make([]value.Value, min)
	for i := 0; i < min; i++ { res[i] = value.Value{Type: value.TypeTuple, Opaque: []value.Value{l1[i], l2[i]}} }
	ptr := new([]value.Value); *ptr = res
	m.Push(value.Value{Type: value.TypeList, Opaque: ptr})
	return nil
}

func Repr(m *vm.Machine) error { return pushString(m, fmt.Sprintf("%#v", m.Pop())) }
func Ascii(m *vm.Machine) error { return Repr(m) }
func Hash(m *vm.Machine) error { m.Push(value.Value{Type: value.TypeInt, Data: m.Pop().Data}); return nil }
func Id(m *vm.Machine) error { return Hash(m) }

func TypeWord(m *vm.Machine) error {
	v := m.Pop()
	names := map[value.Type]string{value.TypeInt: "int", value.TypeFloat: "float", value.TypeBool: "bool", value.TypeString: "str", value.TypeList: "list", value.TypeDict: "dict"}
	return pushString(m, names[v.Type])
}

func Callable(m *vm.Machine) error {
	name := value.UnpackString(m.Pop().Data, m.Arena)
	res := uint64(0); if _, ok := m.FunctionRegistry[name]; ok { res = 1 }
	m.Push(value.Value{Type: value.TypeBool, Data: res}); return nil
}

func Locals(m *vm.Machine) error {
	res := make(map[string]any)
	for i, name := range m.Frames[m.FP].LocalNames { if name != "" { res[name] = m.Frames[m.FP].Locals[i] } }
	m.Push(value.Value{Type: value.TypeDict, Opaque: res}); return nil
}

func Globals(m *vm.Machine) error { return Locals(m) }

func SliceBuiltin(m *vm.Machine) error {
	step := m.Pop(); stop := m.Pop(); start := m.Pop(); obj := m.Pop()
	if obj.Type == value.TypeString {
		s := value.UnpackString(obj.Data, m.Arena)
		if step.Type == value.TypeInt && int64(step.Data) == -1 {
			runes := []rune(s)
			for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 { runes[i], runes[j] = runes[j], runes[i] }
			return pushString(m, string(runes))
		}
		st, sp := 0, len(s)
		if start.Type == value.TypeInt { st = int(start.Int()) }; if stop.Type == value.TypeInt { sp = int(stop.Int()) }
		return pushString(m, s[st:sp])
	}
	return nil
}

func Bytes(m *vm.Machine) error { m.Push(value.Value{Type: value.TypeBytes, Opaque: []byte(value.UnpackString(m.Pop().Data, m.Arena))}); return nil }
func ByteArray(m *vm.Machine) error { return Bytes(m) }

func Enumerate(m *vm.Machine) error {
	l := *(m.Pop().Opaque.(*[]value.Value))
	res := make([]value.Value, len(l))
	for i, v := range l { res[i] = value.Value{Type: value.TypeTuple, Opaque: []value.Value{{Type: value.TypeInt, Data: uint64(i)}, v}} }
	ptr := new([]value.Value); *ptr = res
	m.Push(value.Value{Type: value.TypeList, Opaque: ptr})
	return nil
}

func Len(m *vm.Machine) error {
	v := m.Pop()
	var ln int
	switch v.Type {
	case value.TypeString: ln = int(uint32(v.Data))
	case value.TypeDict: ln = len(v.Opaque.(map[string]any))
	case value.TypeList: ln = len(*(v.Opaque.(*[]value.Value)))
	case value.TypeTuple: ln = len(v.Opaque.([]value.Value))
	}
	m.Push(value.Value{Type: value.TypeInt, Data: uint64(ln)}); return nil
}

type iteratorState struct {
	listPtr *[]value.Value
	index   int
}

func Iter(m *vm.Machine) error {
	v := m.Pop()
	var lp *[]value.Value
	switch v.Type {
	case value.TypeList: lp = v.Opaque.(*[]value.Value)
	case value.TypeTuple: l := v.Opaque.([]value.Value); lp = &l
	case value.TypeDict:
		d := v.Opaque.(map[string]any); l := make([]value.Value, 0, len(d))
		for k := range d { pushString(m, k); l = append(l, m.Pop()) }
		lp = &l
	}
	m.Push(value.Value{Type: value.TypeIterator, Opaque: &iteratorState{listPtr: lp, index: 0}}); return nil
}

func Next(m *vm.Machine) error {
	s := m.Pop().Opaque.(*iteratorState)
	if s.index >= len(*s.listPtr) { return errors.New("stop iteration") }
	m.Push((*s.listPtr)[s.index]); s.index++; return nil
}

func HasNext(m *vm.Machine) error {
	s := m.Peek().Opaque.(*iteratorState)
	res := uint64(0); if s.index < len(*s.listPtr) { res = 1 }
	m.Push(value.Value{Type: value.TypeBool, Data: res}); return nil
}

func Range(m *vm.Machine) error {
	n := int(m.Pop().Int()); st, sp := 0, 0
	if n == 1 { sp = int(m.Pop().Int()) } else { sp = int(m.Pop().Int()); st = int(m.Pop().Int()) }
	res := make([]value.Value, 0)
	for i := st; i < sp; i++ { res = append(res, value.Value{Type: value.TypeInt, Data: uint64(i)}) }
	ptr := new([]value.Value); *ptr = res
	m.Push(value.Value{Type: value.TypeList, Opaque: ptr}); return nil
}

func List(m *vm.Machine) error {
	v := m.Pop()
	if v.Type == value.TypeList { m.Push(v); return nil }
	if v.Type == value.TypeIterator {
		s := v.Opaque.(*iteratorState); res := make([]value.Value, 0)
		for s.index < len(*s.listPtr) { res = append(res, (*s.listPtr)[s.index]); s.index++ }
		ptr := new([]value.Value); *ptr = res
		m.Push(value.Value{Type: value.TypeList, Opaque: ptr}); return nil
	}
	return nil
}

func Sum(m *vm.Machine) error {
	v := m.Pop(); var l []value.Value
	if v.Type == value.TypeList { l = *(v.Opaque.(*[]value.Value)) } else if v.Type == value.TypeIterator { l = (*v.Opaque.(*iteratorState).listPtr)[v.Opaque.(*iteratorState).index:] }
	var total int64; for _, x := range l { if x.Type == value.TypeInt { total += int64(x.Data) } }
	m.Push(value.Value{Type: value.TypeInt, Data: uint64(total)}); return nil
}

func Max(m *vm.Machine) error {
	l := *(m.Pop().Opaque.(*[]value.Value)); mx := l[0]
	for _, x := range l { if x.Data > mx.Data { mx = x } }
	m.Push(mx); return nil
}

func Min(m *vm.Machine) error {
	l := *(m.Pop().Opaque.(*[]value.Value)); mn := l[0]
	for _, x := range l { if x.Data < mn.Data { mn = x } }
	m.Push(mn); return nil
}

func Map(m *vm.Machine) error {
	l := *(m.Pop().Opaque.(*[]value.Value)); name := value.UnpackString(m.Pop().Data, m.Arena)
	ip := m.FunctionRegistry[name]; res := make([]value.Value, len(l))
	for i, x := range l { res[i], _ = m.Call(ip, x) }
	m.Push(value.Value{Type: value.TypeIterator, Opaque: &iteratorState{listPtr: &res, index: 0}}); return nil
}

func Filter(m *vm.Machine) error {
	l := *(m.Pop().Opaque.(*[]value.Value)); name := value.UnpackString(m.Pop().Data, m.Arena)
	ip := m.FunctionRegistry[name]; res := make([]value.Value, 0)
	for _, x := range l { if r, _ := m.Call(ip, x); r.Data != 0 { res = append(res, x) } }
	m.Push(value.Value{Type: value.TypeIterator, Opaque: &iteratorState{listPtr: &res, index: 0}}); return nil
}

func Abs(m *vm.Machine) error { i := int64(m.Pop().Data); if i < 0 { i = -i }; m.Push(value.Value{Type: value.TypeInt, Data: uint64(i)}); return nil }
func Bool(m *vm.Machine) error { res := uint64(0); if m.Pop().Data != 0 { res = 1 }; m.Push(value.Value{Type: value.TypeBool, Data: res}); return nil }
func Int(m *vm.Machine) error {
	v := m.Pop()
	if v.Type == value.TypeFloat { m.Push(value.Value{Type: value.TypeInt, Data: uint64(int64(math.Float64frombits(v.Data)))}); return nil }
	i, _ := strconv.ParseInt(value.UnpackString(v.Data, m.Arena), 10, 64)
	m.Push(value.Value{Type: value.TypeInt, Data: uint64(i)}); return nil
}
func Str(m *vm.Machine) error { return pushString(m, m.Pop().Format(m.Arena)) }
func Pow(m *vm.Machine) error {
	e := float64(m.Pop().Int()); b := float64(m.Pop().Int())
	m.Push(value.Value{Type: value.TypeInt, Data: uint64(int64(math.Pow(b, e)))}); return nil
}

func All(m *vm.Machine) error {
	l := *(m.Pop().Opaque.(*[]value.Value)); res := uint64(1)
	for _, x := range l { if x.Data == 0 { res = 0; break } }
	m.Push(value.Value{Type: value.TypeBool, Data: res}); return nil
}

func Any(m *vm.Machine) error {
	l := *(m.Pop().Opaque.(*[]value.Value)); res := uint64(0)
	for _, x := range l { if x.Data != 0 { res = 1; break } }
	m.Push(value.Value{Type: value.TypeBool, Data: res}); return nil
}

func SetItem(m *vm.Machine) error {
	v := m.Pop(); i := m.Pop(); o := m.Pop()
	if o.Type == value.TypeList { (*o.Opaque.(*[]value.Value))[i.Int()] = v } else { o.Opaque.(map[string]any)[value.UnpackString(i.Data, m.Arena)] = v }
	return nil
}

func MethodCall(m *vm.Machine) error {
	n := int(m.Pop().Int()); name := value.UnpackString(m.Pop().Data, m.Arena)
	args := make([]value.Value, n); for i := n - 1; i >= 0; i-- { args[i] = m.Pop() }
	obj := m.Pop()
	switch obj.Type {
	case value.TypeDict:
		d := obj.Opaque.(map[string]any)
		switch name {
		case "items":
			res := make([]value.Value, 0, len(d))
			for k, v := range d {
				pushString(m, k); kv := m.Pop(); pushConverted(m, v); vv := m.Pop()
				res = append(res, value.Value{Type: value.TypeTuple, Opaque: []value.Value{kv, vv}})
			}
			m.Push(value.Value{Type: value.TypeIterator, Opaque: &iteratorState{listPtr: &res, index: 0}}); return nil
		case "keys":
			res := make([]value.Value, 0, len(d))
			for k := range d { pushString(m, k); res = append(res, m.Pop()) }
			m.Push(value.Value{Type: value.TypeIterator, Opaque: &iteratorState{listPtr: &res, index: 0}}); return nil
		case "get":
			k := value.UnpackString(args[0].Data, m.Arena); def := value.Value{Type: value.TypeVoid}; if n == 2 { def = args[1] }
			if v, ok := d[k]; ok { return pushConverted(m, v) }
			m.Push(def); return nil
		}
	case value.TypeList:
		l := obj.Opaque.(*[]value.Value)
		if name == "append" { *l = append(*l, args[0]); m.Push(value.Value{Type: value.TypeVoid}); return nil }
	case value.TypeString:
		s := value.UnpackString(obj.Data, m.Arena)
		switch name {
		case "upper": return pushString(m, strings.ToUpper(s))
		case "lower": return pushString(m, strings.ToLower(s))
		case "split":
			sep := " "; if n == 1 { sep = value.UnpackString(args[0].Data, m.Arena) }
			parts := strings.Split(s, sep); res := make([]value.Value, len(parts))
			for i, p := range parts { pushString(m, p); res[i] = m.Pop() }
			ptr := new([]value.Value); *ptr = res; m.Push(value.Value{Type: value.TypeList, Opaque: ptr}); return nil
		case "join":
			l := *(args[0].Opaque.(*[]value.Value)); ss := make([]string, len(l))
			for i, x := range l { ss[i] = x.Format(m.Arena) }
			return pushString(m, strings.Join(ss, s))
		case "find": pushString(m, strconv.Itoa(strings.Index(s, value.UnpackString(args[0].Data, m.Arena)))); return nil
		case "json": m.Push(obj); return ParseJSON(m)
		}
	}
	return nil
}

func Print(m *vm.Machine) error {
	n := int(m.Pop().Int()); ss := make([]string, n)
	for i := n - 1; i >= 0; i-- { ss[i] = m.Pop().Format(m.Arena) }
	fmt.Println(strings.Join(ss, " ")); m.Push(value.Value{Type: value.TypeVoid}); return nil
}

func MakeList(m *vm.Machine) error {
	n := int(m.Pop().Int())
	l := make([]value.Value, n)
	for i := n - 1; i >= 0; i-- { l[i] = m.Pop() }
	ptr := new([]value.Value); *ptr = l
	m.Push(value.Value{Type: value.TypeList, Opaque: ptr})
	return nil
}

func GetItem(m *vm.Machine) error {
	idxVal := m.Pop(); obj := m.Pop()
	if obj.Type == value.TypeList {
		l := *(obj.Opaque.(*[]value.Value))
		idx := int(idxVal.Int())
		if idx < 0 { idx += len(l) }
		m.Push(l[idx])
	} else if obj.Type == value.TypeTuple {
		l := obj.Opaque.([]value.Value)
		idx := int(idxVal.Int())
		if idx < 0 { idx += len(l) }
		m.Push(l[idx])
	} else if obj.Type == value.TypeDict {
		d := obj.Opaque.(map[string]any)
		key := value.UnpackString(idxVal.Data, m.Arena)
		return pushConverted(m, d[key])
	}
	return nil
}
