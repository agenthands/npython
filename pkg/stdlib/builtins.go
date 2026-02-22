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
	offset, err := m.WriteArena([]byte(s))
	if err != nil {
		return err
	}
	m.Push(value.Value{Type: value.TypeString, Data: value.PackString(offset, uint32(len(s)))})
	return nil
}

// DivMod: ( a b -- tuple(q, r) )
func DivMod(m *vm.Machine) error {
	bVal := m.Pop()
	aVal := m.Pop()
	if bVal.Type != value.TypeInt || aVal.Type != value.TypeInt {
		return errors.New("TypeError: unsupported operand type(s) for divmod()")
	}
	b := bVal.Int()
	a := aVal.Int()
	if b == 0 {
		return errors.New("ZeroDivisionError: integer division or modulo by zero")
	}
	m.Push(value.Value{Type: value.TypeTuple, Opaque: []value.Value{
		{Type: value.TypeInt, Data: uint64(a / b)},
		{Type: value.TypeInt, Data: uint64(a % b)},
	}})
	return nil
}

func Round(m *vm.Machine) error {
	nVal := m.Pop()
	if nVal.Type != value.TypeInt {
		return errors.New("TypeError: expected int for arg count")
	}
	n := nVal.Int()

	var ndigits int64
	if n == 2 {
		ndVal := m.Pop()
		if ndVal.Type != value.TypeInt {
			return errors.New("TypeError: ndigits must be an integer")
		}
		ndigits = ndVal.Int()
	} else if n != 1 {
		return fmt.Errorf("TypeError: round() takes at most 2 arguments (%d given)", n)
	}

	v := m.Pop()
	var f float64
	if v.Type == value.TypeInt {
		f = float64(v.Int())
	} else if v.Type == value.TypeFloat {
		f = math.Float64frombits(v.Data)
	} else {
		return fmt.Errorf("TypeError: type %d doesn't define __round__ method", v.Type)
	}

	if n == 1 {
		m.Push(value.Value{Type: value.TypeInt, Data: uint64(int64(math.Round(f)))})
	} else {
		p := math.Pow(10, float64(ndigits))
		res := math.Round(f*p) / p
		m.Push(value.Value{Type: value.TypeFloat, Data: math.Float64bits(res)})
	}
	return nil
}

func Float(m *vm.Machine) error {
	v := m.Pop()
	var f float64
	switch v.Type {
	case value.TypeInt:
		f = float64(v.Int())
	case value.TypeFloat:
		f = math.Float64frombits(v.Data)
	case value.TypeString:
		var err error
		f, err = strconv.ParseFloat(value.UnpackString(v.Data, m.Arena), 64)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("TypeError: float() argument must be a string or a number, not '%v'", v.Type)
	}
	m.Push(value.Value{Type: value.TypeFloat, Data: math.Float64bits(f)})
	return nil
}

func Bin(m *vm.Machine) error {
	v := m.Pop()
	if v.Type != value.TypeInt {
		return errors.New("TypeError: expected integer")
	}
	return pushString(m, "0b"+strconv.FormatInt(v.Int(), 2))
}
func Oct(m *vm.Machine) error {
	v := m.Pop()
	if v.Type != value.TypeInt {
		return errors.New("TypeError: expected integer")
	}
	return pushString(m, "0o"+strconv.FormatInt(v.Int(), 8))
}
func Hex(m *vm.Machine) error {
	v := m.Pop()
	if v.Type != value.TypeInt {
		return errors.New("TypeError: expected integer")
	}
	return pushString(m, "0x"+strconv.FormatInt(v.Int(), 16))
}
func Chr(m *vm.Machine) error {
	v := m.Pop()
	if v.Type != value.TypeInt {
		return errors.New("TypeError: expected integer")
	}
	return pushString(m, string(rune(v.Int())))
}
func Ord(m *vm.Machine) error {
	v := m.Pop()
	if v.Type != value.TypeString {
		return errors.New("TypeError: expected string")
	}
	s := value.UnpackString(v.Data, m.Arena)
	if len(s) == 0 {
		return errors.New("empty ord")
	}
	m.Push(value.Value{Type: value.TypeInt, Data: uint64(s[0])})
	return nil
}

func Dict(m *vm.Machine) error {
	m.Push(value.Value{Type: value.TypeDict, Opaque: make(map[string]any)})
	return nil
}

func MakeTuple(m *vm.Machine) error {
	nVal := m.Pop()
	if nVal.Type != value.TypeInt {
		return errors.New("TypeError: tuple size must be integer")
	}
	n := int(nVal.Int())
	l := make([]value.Value, n)
	for i := n - 1; i >= 0; i-- {
		l[i] = m.Pop()
	}
	m.Push(value.Value{Type: value.TypeTuple, Opaque: l})
	return nil
}

func Tuple(m *vm.Machine) error {
	v := m.Pop()
	var list []value.Value
	if v.Type == value.TypeList {
		list = *(v.Opaque.(*[]value.Value))
	} else if v.Type == value.TypeTuple {
		list = v.Opaque.([]value.Value)
	} else {
		return errors.New("TypeError: not iterable")
	}
	m.Push(value.Value{Type: value.TypeTuple, Opaque: list})
	return nil
}

func Set(m *vm.Machine) error {
	v := m.Pop()
	var list []value.Value
	if v.Type == value.TypeList {
		list = *(v.Opaque.(*[]value.Value))
	} else if v.Type == value.TypeTuple {
		list = v.Opaque.([]value.Value)
	} else {
		return errors.New("TypeError: not iterable")
	}
	s := make(map[any]struct{})
	for _, x := range list {
		s[x.Data] = struct{}{}
	}
	m.Push(value.Value{Type: value.TypeSet, Opaque: s})
	return nil
}

func Reversed(m *vm.Machine) error {
	v := m.Pop()
	if v.Type != value.TypeList {
		return errors.New("TypeError: not iterable")
	}
	l := *(v.Opaque.(*[]value.Value))
	res := make([]value.Value, len(l))
	for i, j := 0, len(l)-1; i < len(l); i, j = i+1, j-1 {
		res[i] = l[j]
	}
	ptr := new([]value.Value)
	*ptr = res
	m.Push(value.Value{Type: value.TypeList, Opaque: ptr})
	return nil
}

func Sorted(m *vm.Machine) error {
	v := m.Pop()
	if v.Type != value.TypeList {
		return errors.New("TypeError: not iterable")
	}
	l := *(v.Opaque.(*[]value.Value))
	res := make([]value.Value, len(l))
	copy(res, l)
	sort.Slice(res, func(i, j int) bool { return res[i].Data < res[j].Data })
	ptr := new([]value.Value)
	*ptr = res
	m.Push(value.Value{Type: value.TypeList, Opaque: ptr})
	return nil
}

func Zip(m *vm.Machine) error {
	v2 := m.Pop()
	v1 := m.Pop()
	if v1.Type != value.TypeList || v2.Type != value.TypeList {
		return errors.New("TypeError: not iterable")
	}
	l2 := *(v2.Opaque.(*[]value.Value))
	l1 := *(v1.Opaque.(*[]value.Value))
	min := len(l1)
	if len(l2) < min {
		min = len(l2)
	}
	res := make([]value.Value, min)
	for i := 0; i < min; i++ {
		res[i] = value.Value{Type: value.TypeTuple, Opaque: []value.Value{l1[i], l2[i]}}
	}
	ptr := new([]value.Value)
	*ptr = res
	m.Push(value.Value{Type: value.TypeList, Opaque: ptr})
	return nil
}

func Repr(m *vm.Machine) error  { return pushString(m, fmt.Sprintf("%#v", m.Pop())) }
func Ascii(m *vm.Machine) error { return Repr(m) }
func Hash(m *vm.Machine) error {
	m.Push(value.Value{Type: value.TypeInt, Data: m.Pop().Data})
	return nil
}
func Id(m *vm.Machine) error { return Hash(m) }

func TypeWord(m *vm.Machine) error {
	v := m.Pop()
	names := map[value.Type]string{value.TypeInt: "int", value.TypeFloat: "float", value.TypeBool: "bool", value.TypeString: "str", value.TypeList: "list", value.TypeDict: "dict"}
	return pushString(m, names[v.Type])
}

func Callable(m *vm.Machine) error {
	name := value.UnpackString(m.Pop().Data, m.Arena)
	res := uint64(0)
	if _, ok := m.FunctionRegistry[name]; ok {
		res = 1
	}
	m.Push(value.Value{Type: value.TypeBool, Data: res})
	return nil
}

func Locals(m *vm.Machine) error {
	res := make(map[string]any)
	for i, name := range m.Frames[m.FP].LocalNames {
		if name != "" {
			res[name] = m.Frames[m.FP].Locals[i]
		}
	}
	m.Push(value.Value{Type: value.TypeDict, Opaque: res})
	return nil
}

func Globals(m *vm.Machine) error { return Locals(m) }

func SliceBuiltin(m *vm.Machine) error {
	step := m.Pop()
	stop := m.Pop()
	start := m.Pop()
	obj := m.Pop()
	if obj.Type == value.TypeString {
		s := value.UnpackString(obj.Data, m.Arena)
		if step.Type == value.TypeInt && int64(step.Data) == -1 {
			runes := []rune(s)
			for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
				runes[i], runes[j] = runes[j], runes[i]
			}
			return pushString(m, string(runes))
		}
		st, sp := 0, len(s)
		if start.Type == value.TypeInt {
			st = int(start.Int())
		}
		if stop.Type == value.TypeInt {
			sp = int(stop.Int())
		}
		return pushString(m, s[st:sp])
	}
	return nil
}

func Bytes(m *vm.Machine) error {
	m.Push(value.Value{Type: value.TypeBytes, Opaque: []byte(value.UnpackString(m.Pop().Data, m.Arena))})
	return nil
}
func ByteArray(m *vm.Machine) error { return Bytes(m) }

func Enumerate(m *vm.Machine) error {
	v := m.Pop()
	if v.Type != value.TypeList {
		return errors.New("TypeError: not iterable")
	}
	l := *(v.Opaque.(*[]value.Value))
	res := make([]value.Value, len(l))
	for i, v := range l {
		res[i] = value.Value{Type: value.TypeTuple, Opaque: []value.Value{{Type: value.TypeInt, Data: uint64(i)}, v}}
	}
	ptr := new([]value.Value)
	*ptr = res
	m.Push(value.Value{Type: value.TypeList, Opaque: ptr})
	return nil
}

func Len(m *vm.Machine) error {
	v := m.Pop()
	var ln int
	switch v.Type {
	case value.TypeString:
		ln = int(uint32(v.Data))
	case value.TypeDict:
		ln = len(v.Opaque.(map[string]any))
	case value.TypeList:
		ln = len(*(v.Opaque.(*[]value.Value)))
	case value.TypeTuple:
		ln = len(v.Opaque.([]value.Value))
	case value.TypeIterator:
		ln = len(*(v.Opaque.(*iteratorState).listPtr))
	default:
		return fmt.Errorf("TypeError: object of type %d has no len()", v.Type)
	}
	m.Push(value.Value{Type: value.TypeInt, Data: uint64(ln)})
	return nil
}

type iteratorState struct {
	listPtr *[]value.Value
	index   int
}

func Iter(m *vm.Machine) error {
	v := m.Pop()
	var lp *[]value.Value
	switch v.Type {
	case value.TypeList:
		lp = v.Opaque.(*[]value.Value)
	case value.TypeTuple:
		l := v.Opaque.([]value.Value)
		lp = &l
	case value.TypeDict:
		d := v.Opaque.(map[string]any)
		l := make([]value.Value, 0, len(d))
		for k := range d {
			pushString(m, k)
			l = append(l, m.Pop())
		}
		lp = &l
	default:
		return fmt.Errorf("TypeError: '%v' object is not iterable", v.Type)
	}
	m.Push(value.Value{Type: value.TypeIterator, Opaque: &iteratorState{listPtr: lp, index: 0}})
	return nil
}

func Next(m *vm.Machine) error {
	v := m.Pop()
	if v.Type != value.TypeIterator {
		return fmt.Errorf("TypeError: '%v' object is not an iterator", v.Type)
	}
	s := v.Opaque.(*iteratorState)
	if s.index >= len(*s.listPtr) {
		return errors.New("stop iteration")
	}
	m.Push((*s.listPtr)[s.index])
	s.index++
	return nil
}

func HasNext(m *vm.Machine) error {
	s := m.Peek().Opaque.(*iteratorState)
	res := uint64(0)
	if s.index < len(*s.listPtr) {
		res = 1
	}
	m.Push(value.Value{Type: value.TypeBool, Data: res})
	return nil
}

func Range(m *vm.Machine) error {
	n := int(m.Pop().Int())
	st, sp := 0, 0
	if n == 1 {
		spVal := m.Pop()
		if spVal.Type != value.TypeInt {
			return errors.New("TypeError")
		}
		sp = int(spVal.Int())
	} else {
		spVal := m.Pop()
		stVal := m.Pop()
		if spVal.Type != value.TypeInt || stVal.Type != value.TypeInt {
			return errors.New("TypeError")
		}
		sp = int(spVal.Int())
		st = int(stVal.Int())
	}
	res := make([]value.Value, 0)
	for i := st; i < sp; i++ {
		res = append(res, value.Value{Type: value.TypeInt, Data: uint64(i)})
	}
	ptr := new([]value.Value)
	*ptr = res
	m.Push(value.Value{Type: value.TypeList, Opaque: ptr})
	return nil
}

func List(m *vm.Machine) error {
	v := m.Pop()
	if v.Type == value.TypeList {
		m.Push(v)
		return nil
	}
	if v.Type == value.TypeIterator {
		s := v.Opaque.(*iteratorState)
		res := make([]value.Value, 0)
		for s.index < len(*s.listPtr) {
			res = append(res, (*s.listPtr)[s.index])
			s.index++
		}
		ptr := new([]value.Value)
		*ptr = res
		m.Push(value.Value{Type: value.TypeList, Opaque: ptr})
		m.Push(value.Value{Type: value.TypeList, Opaque: ptr})
		return nil
	}
	return fmt.Errorf("TypeError: '%v' object is not iterable", v.Type)
}

func Sum(m *vm.Machine) error {
	n := int(m.Pop().Int())
	if n == 0 {
		return errors.New("TypeError: sum() expected at least 1 argument, got 0")
	}

	var start int64 = 0
	if n == 2 {
		start = m.Pop().Int()
	}

	v := m.Pop()
	var l []value.Value
	if v.Type == value.TypeList {
		l = *(v.Opaque.(*[]value.Value))
	} else if v.Type == value.TypeIterator {
		l = (*v.Opaque.(*iteratorState).listPtr)[v.Opaque.(*iteratorState).index:]
	} else {
		return fmt.Errorf("TypeError: '%v' object is not iterable", v.Type)
	}

	var total int64 = start
	for _, x := range l {
		if x.Type == value.TypeInt {
			total += int64(x.Data)
		} else if x.Type == value.TypeFloat {
			total += int64(math.Float64frombits(x.Data))
		}
	}
	m.Push(value.Value{Type: value.TypeInt, Data: uint64(total)})
	return nil
}

func Max(m *vm.Machine) error {
	n := int(m.Pop().Int())
	if n == 0 {
		return errors.New("TypeError: max() expected at least 1 argument, got 0")
	}

	var l []value.Value
	if n == 1 {
		v := m.Pop()
		if v.Type != value.TypeList {
			return fmt.Errorf("TypeError: '%v' object is not iterable", v.Type)
		}
		l = *(v.Opaque.(*[]value.Value))
	} else {
		l = make([]value.Value, n)
		for i := n - 1; i >= 0; i-- {
			l[i] = m.Pop()
		}
	}

	if len(l) == 0 {
		return errors.New("ValueError: max() arg is an empty sequence")
	}
	mx := l[0]
	for _, x := range l {
		if x.Int() > mx.Int() {
			mx = x
		}
	}
	m.Push(mx)
	return nil
}

func Min(m *vm.Machine) error {
	n := int(m.Pop().Int())
	if n == 0 {
		return errors.New("TypeError: min() expected at least 1 argument, got 0")
	}

	var l []value.Value
	if n == 1 {
		v := m.Pop()
		if v.Type != value.TypeList {
			return fmt.Errorf("TypeError: '%v' object is not iterable", v.Type)
		}
		l = *(v.Opaque.(*[]value.Value))
	} else {
		l = make([]value.Value, n)
		for i := n - 1; i >= 0; i-- {
			l[i] = m.Pop()
		}
	}

	if len(l) == 0 {
		return errors.New("ValueError: min() arg is an empty sequence")
	}
	mn := l[0]
	for _, x := range l {
		if x.Int() < mn.Int() {
			mn = x
		}
	}
	m.Push(mn)
	return nil
}

func Map(m *vm.Machine) error {
	v := m.Pop()
	if v.Type != value.TypeList {
		return fmt.Errorf("TypeError: '%v' object is not iterable", v.Type)
	}
	l := *(v.Opaque.(*[]value.Value))
	name := value.UnpackString(m.Pop().Data, m.Arena)
	ip, exists := m.FunctionRegistry[name]
	if !exists {
		return fmt.Errorf("NameError: name '%s' is not defined", name)
	}
	res := make([]value.Value, len(l))
	for i, x := range l {
		res[i], _ = m.Call(ip, x)
	}
	m.Push(value.Value{Type: value.TypeIterator, Opaque: &iteratorState{listPtr: &res, index: 0}})
	return nil
}

func Filter(m *vm.Machine) error {
	v := m.Pop()
	if v.Type != value.TypeList {
		return fmt.Errorf("TypeError: '%v' object is not iterable", v.Type)
	}
	l := *(v.Opaque.(*[]value.Value))
	name := value.UnpackString(m.Pop().Data, m.Arena)
	ip, exists := m.FunctionRegistry[name]
	if !exists {
		return fmt.Errorf("NameError: name '%s' is not defined", name)
	}
	res := make([]value.Value, 0)
	for _, x := range l {
		if r, _ := m.Call(ip, x); r.Data != 0 {
			res = append(res, x)
		}
	}
	m.Push(value.Value{Type: value.TypeIterator, Opaque: &iteratorState{listPtr: &res, index: 0}})
	return nil
}

func Abs(m *vm.Machine) error {
	v := m.Pop()
	if v.Type == value.TypeInt {
		i := int64(v.Data)
		if i < 0 {
			i = -i
		}
		m.Push(value.Value{Type: value.TypeInt, Data: uint64(i)})
		return nil
	} else if v.Type == value.TypeFloat {
		f := math.Float64frombits(v.Data)
		m.Push(value.Value{Type: value.TypeFloat, Data: math.Float64bits(math.Abs(f))})
		return nil
	}
	return errors.New("TypeError: bad operand type for abs()")
}
func Bool(m *vm.Machine) error {
	res := uint64(0)
	if vm.IsTruthy(m.Pop()) {
		res = 1
	}
	m.Push(value.Value{Type: value.TypeBool, Data: res})
	return nil
}
func Int(m *vm.Machine) error {
	v := m.Pop()
	if v.Type == value.TypeInt {
		m.Push(v)
		return nil
	}
	if v.Type == value.TypeFloat {
		m.Push(value.Value{Type: value.TypeInt, Data: uint64(int64(math.Float64frombits(v.Data)))})
		return nil
	}
	i, err := strconv.ParseInt(value.UnpackString(v.Data, m.Arena), 10, 64)
	if err != nil {
		return err
	}
	m.Push(value.Value{Type: value.TypeInt, Data: uint64(i)})
	return nil
}
func Str(m *vm.Machine) error { return pushString(m, m.Pop().Format(m.Arena)) }
func Pow(m *vm.Machine) error {
	eVal := m.Pop()
	bVal := m.Pop()
	if bVal.Type != value.TypeInt && bVal.Type != value.TypeFloat {
		return errors.New("TypeError: unsupported operand type for pow()")
	}
	if eVal.Type != value.TypeInt && eVal.Type != value.TypeFloat {
		return errors.New("TypeError: unsupported operand type for pow()")
	}

	e := float64(eVal.Int())
	if eVal.Type == value.TypeFloat {
		e = math.Float64frombits(eVal.Data)
	}
	b := float64(bVal.Int())
	if bVal.Type == value.TypeFloat {
		b = math.Float64frombits(bVal.Data)
	}

	m.Push(value.Value{Type: value.TypeInt, Data: uint64(int64(math.Pow(b, e)))})
	return nil
}

func All(m *vm.Machine) error {
	v := m.Pop()
	if v.Type != value.TypeList {
		return errors.New("TypeError: not iterable")
	}
	l := *(v.Opaque.(*[]value.Value))
	res := uint64(1)
	for _, x := range l {
		if x.Data == 0 {
			res = 0
			break
		}
	}
	m.Push(value.Value{Type: value.TypeBool, Data: res})
	return nil
}

func Any(m *vm.Machine) error {
	v := m.Pop()
	if v.Type != value.TypeList {
		return errors.New("TypeError: not iterable")
	}
	l := *(v.Opaque.(*[]value.Value))
	res := uint64(0)
	for _, x := range l {
		if x.Data != 0 {
			res = 1
			break
		}
	}
	m.Push(value.Value{Type: value.TypeBool, Data: res})
	return nil
}

func SetItem(m *vm.Machine) error {
	v := m.Pop()
	idxVal := m.Pop()
	obj := m.Pop()
	if obj.Type == value.TypeList {
		if idxVal.Type != value.TypeInt {
			return errors.New("TypeError: list indices must be integers")
		}
		l := *(obj.Opaque.(*[]value.Value))
		idx := int(idxVal.Int())
		if idx < 0 {
			idx += len(l)
		}
		if idx < 0 || idx >= len(l) {
			return errors.New("IndexError: list assignment index out of range")
		}
		l[idx] = v
	} else if obj.Type == value.TypeDict {
		if idxVal.Type != value.TypeString {
			return errors.New("TypeError: dictionary keys must be strings in this implementation")
		}
		key := value.UnpackString(idxVal.Data, m.Arena)
		obj.Opaque.(map[string]any)[key] = v
	} else {
		return fmt.Errorf("TypeError: '%v' object does not support item assignment", obj.Type)
	}
	return nil
}

func MethodCall(m *vm.Machine) error {
	nVal := m.Pop()
	if nVal.Type != value.TypeInt {
		return errors.New("TypeError: method call arg count must be integer")
	}
	n := int(nVal.Int())
	nameVal := m.Pop()
	if nameVal.Type != value.TypeString {
		return errors.New("TypeError: method name must be string")
	}
	name := value.UnpackString(nameVal.Data, m.Arena)
	args := make([]value.Value, n)
	for i := n - 1; i >= 0; i-- {
		args[i] = m.Pop()
	}
	obj := m.Pop()
	switch obj.Type {
	case value.TypeDict:
		d := obj.Opaque.(map[string]any)
		switch name {
		case "items":
			res := make([]value.Value, 0, len(d))
			for k, v := range d {
				pushString(m, k)
				kv := m.Pop()
				pushConverted(m, v)
				vv := m.Pop()
				res = append(res, value.Value{Type: value.TypeTuple, Opaque: []value.Value{kv, vv}})
			}
			m.Push(value.Value{Type: value.TypeIterator, Opaque: &iteratorState{listPtr: &res, index: 0}})
			return nil
		case "keys":
			res := make([]value.Value, 0, len(d))
			for k := range d {
				pushString(m, k)
				res = append(res, m.Pop())
			}
			m.Push(value.Value{Type: value.TypeIterator, Opaque: &iteratorState{listPtr: &res, index: 0}})
			return nil
		case "get":
			k := value.UnpackString(args[0].Data, m.Arena)
			def := value.Value{Type: value.TypeVoid}
			if n == 2 {
				def = args[1]
			}
			if v, ok := d[k]; ok {
				return pushConverted(m, v)
			}
			m.Push(def)
			return nil
		}
	case value.TypeList:
		l := obj.Opaque.(*[]value.Value)
		if name == "append" {
			*l = append(*l, args[0])
			m.Push(value.Value{Type: value.TypeVoid})
			return nil
		}
	case value.TypeString:
		s := value.UnpackString(obj.Data, m.Arena)
		switch name {
		case "upper":
			return pushString(m, strings.ToUpper(s))
		case "lower":
			return pushString(m, strings.ToLower(s))
		case "split":
			sep := " "
			if n == 1 {
				sep = value.UnpackString(args[0].Data, m.Arena)
			}
			parts := strings.Split(s, sep)
			res := make([]value.Value, len(parts))
			for i, p := range parts {
				pushString(m, p)
				res[i] = m.Pop()
			}
			ptr := new([]value.Value)
			*ptr = res
			m.Push(value.Value{Type: value.TypeList, Opaque: ptr})
			return nil
		case "join":
			l := *(args[0].Opaque.(*[]value.Value))
			ss := make([]string, len(l))
			for i, x := range l {
				ss[i] = x.Format(m.Arena)
			}
			return pushString(m, strings.Join(ss, s))
		case "find":
			pushString(m, strconv.Itoa(strings.Index(s, value.UnpackString(args[0].Data, m.Arena))))
			return nil
		case "json":
			m.Push(obj)
			return ParseJSON(m)
		case "format":
			res := s
			for _, arg := range args {
				res = strings.Replace(res, "{}", arg.Format(m.Arena), 1)
			}
			for i, arg := range args {
				res = strings.ReplaceAll(res, fmt.Sprintf("{%d}", i), arg.Format(m.Arena))
			}
			return pushString(m, res)
		}
	}
	return nil
}

func Print(m *vm.Machine) error {
	n := int(m.Pop().Int())
	ss := make([]string, n)
	for i := n - 1; i >= 0; i-- {
		ss[i] = m.Pop().Format(m.Arena)
	}
	fmt.Println(strings.Join(ss, " "))
	m.Push(value.Value{Type: value.TypeVoid})
	return nil
}

func MakeList(m *vm.Machine) error {
	nVal := m.Pop()
	if nVal.Type != value.TypeInt {
		return errors.New("TypeError: list size must be integer")
	}
	n := int(nVal.Int())
	l := make([]value.Value, n)
	for i := n - 1; i >= 0; i-- {
		l[i] = m.Pop()
	}
	ptr := new([]value.Value)
	*ptr = l
	m.Push(value.Value{Type: value.TypeList, Opaque: ptr})
	return nil
}

func GetItem(m *vm.Machine) error {
	idxVal := m.Pop()
	obj := m.Pop()
	if obj.Type == value.TypeList {
		if idxVal.Type != value.TypeInt {
			return errors.New("TypeError: list indices must be integers")
		}
		l := *(obj.Opaque.(*[]value.Value))
		idx := int(idxVal.Int())
		if idx < 0 {
			idx += len(l)
		}
		if idx < 0 || idx >= len(l) {
			return errors.New("IndexError: list index out of range")
		}
		m.Push(l[idx])
		return nil
	} else if obj.Type == value.TypeTuple {
		l := obj.Opaque.([]value.Value)
		if idxVal.Type != value.TypeInt {
			return errors.New("TypeError: tuple indices must be integers")
		}
		idx := int(idxVal.Int())
		if idx < 0 {
			idx += len(l)
		}
		if idx < 0 || idx >= len(l) {
			return errors.New("IndexError: tuple index out of range")
		}
		m.Push(l[idx])
		return nil
	} else if obj.Type == value.TypeDict {
		if idxVal.Type != value.TypeString {
			return errors.New("TypeError: dictionary keys must be strings")
		}
		d := obj.Opaque.(map[string]any)
		key := value.UnpackString(idxVal.Data, m.Arena)
		val, ok := d[key]
		if !ok {
			return fmt.Errorf("KeyError: %s", key)
		}
		return pushConverted(m, val)
	}
	return fmt.Errorf("TypeError: cannot index into object of type %v", obj.Type)
}

func IsInstance(m *vm.Machine) error {
	typeWordOrStr := m.Pop()
	obj := m.Pop()

	typeStr := ""
	if typeWordOrStr.Type == value.TypeString {
		typeStr = value.UnpackString(typeWordOrStr.Data, m.Arena)
	} else if typeWordOrStr.Type == value.TypeVoid {
		// Just false if NoneType checking etc isn't perfectly aligned
	}

	res := uint64(0)
	switch obj.Type {
	case value.TypeInt:
		if typeStr == "int" {
			res = 1
		}
	case value.TypeFloat:
		if typeStr == "float" {
			res = 1
		}
	case value.TypeString:
		if typeStr == "str" {
			res = 1
		}
	case value.TypeBool:
		if typeStr == "bool" {
			res = 1
		}
	case value.TypeList:
		if typeStr == "list" {
			res = 1
		}
	case value.TypeDict:
		if typeStr == "dict" {
			res = 1
		}
	case value.TypeTuple:
		if typeStr == "tuple" {
			res = 1
		}
	case value.TypeSet:
		if typeStr == "set" {
			res = 1
		}
	case value.TypeBytes:
		if typeStr == "bytes" || typeStr == "bytearray" {
			res = 1
		}
	}

	m.Push(value.Value{Type: value.TypeBool, Data: res})
	return nil
}
