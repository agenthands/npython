package stdlib

import (
	"math"
	"strings"
	"testing"

	"github.com/agenthands/npython/pkg/core/value"
	"github.com/agenthands/npython/pkg/vm"
)

func TestBuiltins(t *testing.T) {
	m := vm.GetMachine()
	defer vm.PutMachine(m)

	t.Run("DivMod", func(t *testing.T) {
		m.Reset()
		m.Push(value.Value{Type: value.TypeInt, Data: 10})
		m.Push(value.Value{Type: value.TypeInt, Data: 3})
		if err := DivMod(m); err != nil {
			t.Fatal(err)
		}
		res := m.Pop()
		if res.Type != value.TypeTuple {
			t.Errorf("expected tuple")
		}
		tuple := res.Opaque.([]value.Value)
		if tuple[0].Int() != 3 || tuple[1].Int() != 1 {
			t.Errorf("got %v, %v", tuple[0].Int(), tuple[1].Int())
		}
	})

	t.Run("Round", func(t *testing.T) {
		m.Reset()
		m.Push(value.Value{Type: value.TypeFloat, Data: 0x40091eb851eb851f}) // 3.14
		m.Push(value.Value{Type: value.TypeInt, Data: 1})
		if err := Round(m); err != nil {
			t.Fatal(err)
		}
		if m.Pop().Int() != 3 {
			t.Errorf("round(3.14) failed")
		}
	})

	t.Run("Float", func(t *testing.T) {
		m.Reset()
		m.Push(value.Value{Type: value.TypeInt, Data: 42})
		if err := Float(m); err != nil {
			t.Fatal(err)
		}
		if m.Pop().Type != value.TypeFloat {
			t.Errorf("expected float")
		}
	})

	t.Run("BinOctHex", func(t *testing.T) {
		m.Reset()
		m.Push(value.Value{Type: value.TypeInt, Data: 255})
		if err := Hex(m); err != nil {
			t.Fatal(err)
		}
		h := value.UnpackString(m.Pop().Data, m.Arena)
		if h != "0xff" {
			t.Errorf("hex(255) = %s", h)
		}

		m.Push(value.Value{Type: value.TypeInt, Data: 7})
		Bin(m)
		if value.UnpackString(m.Pop().Data, m.Arena) != "0b111" {
			t.Errorf("bin(7) failed")
		}

		m.Push(value.Value{Type: value.TypeInt, Data: 10})
		Oct(m)
		if value.UnpackString(m.Pop().Data, m.Arena) != "0o12" {
			t.Errorf("oct(10) failed")
		}
	})

	t.Run("Collections", func(t *testing.T) {
		m.Reset()
		m.Push(value.Value{Type: value.TypeDict, Opaque: map[string]any{}})
		Dict(m)
		if m.Pop().Type != value.TypeDict {
			t.Errorf("expected dict")
		}

		m.Push(value.Value{Type: value.TypeInt, Data: 3})
		m.Push(value.Value{Type: value.TypeInt, Data: 1})
		Range(m)
		m.Push(m.Stack[m.SP-1])
		Tuple(m)
		if m.Pop().Type != value.TypeTuple {
			t.Errorf("expected tuple")
		}

		m.Push(m.Stack[m.SP-1])
		Set(m)
		if m.Pop().Type != value.TypeSet {
			t.Errorf("expected set")
		}
	})

	t.Run("ReversedSorted", func(t *testing.T) {
		m.Reset()
		// [2, 1, 0]
		list := []value.Value{
			{Type: value.TypeInt, Data: 2},
			{Type: value.TypeInt, Data: 1},
			{Type: value.TypeInt, Data: 0},
		}
		m.Push(value.Value{Type: value.TypeList, Opaque: &list})
		m.Push(m.Stack[m.SP-1])
		Sorted(m)
		sorted := *(m.Pop().Opaque.(*[]value.Value))
		if sorted[0].Int() != 0 {
			t.Errorf("sorted failed")
		}

		Reversed(m)
		rev := *(m.Pop().Opaque.(*[]value.Value))
		if rev[0].Int() != 0 {
			t.Errorf("reversed failed")
		}
	})

	t.Run("ZipEnumerate", func(t *testing.T) {
		m.Reset()
		l1 := []value.Value{{Type: value.TypeInt, Data: 1}}
		l2 := []value.Value{{Type: value.TypeString, Data: 0}}
		m.Push(value.Value{Type: value.TypeList, Opaque: &l1})
		m.Push(value.Value{Type: value.TypeList, Opaque: &l2})
		Zip(m)
		res := *(m.Pop().Opaque.(*[]value.Value))
		if len(res) != 1 || res[0].Type != value.TypeTuple {
			t.Errorf("zip failed")
		}

		m.Push(value.Value{Type: value.TypeList, Opaque: &l1})
		Enumerate(m)
		enum := *(m.Pop().Opaque.(*[]value.Value))
		if len(enum) != 1 || enum[0].Type != value.TypeTuple {
			t.Errorf("enumerate failed")
		}
	})

	t.Run("Introspection", func(t *testing.T) {
		m.Reset()
		m.Push(value.Value{Type: value.TypeInt, Data: 123})
		Hash(m)
		if m.Pop().Int() != 123 {
			t.Errorf("hash failed")
		}

		m.Push(value.Value{Type: value.TypeInt, Data: 123})
		Id(m)
		if m.Pop().Int() != 123 {
			t.Errorf("id failed")
		}

		m.Push(value.Value{Type: value.TypeInt, Data: 123})
		Repr(m)
		if value.UnpackString(m.Pop().Data, m.Arena) == "" {
			t.Errorf("repr failed")
		}

		m.Push(value.Value{Type: value.TypeInt, Data: 123})
		Ascii(m)
		if value.UnpackString(m.Pop().Data, m.Arena) == "" {
			t.Errorf("ascii failed")
		}

		Locals(m)
		if m.Pop().Type != value.TypeDict {
			t.Errorf("locals failed")
		}

		Globals(m)
		if m.Pop().Type != value.TypeDict {
			t.Errorf("globals failed")
		}

		m.Push(value.Value{Type: value.TypeString, Data: value.PackString(0, 0)})
		m.Push(value.Value{Type: value.TypeInt, Data: 0})
		m.Push(value.Value{Type: value.TypeInt, Data: 0})
		m.Push(value.Value{Type: value.TypeInt, Data: 1})
		SliceBuiltin(m)
		m.Pop()
	})

	t.Run("Bytes", func(t *testing.T) {
		m.Reset()
		m.Push(value.Value{Type: value.TypeString, Data: value.PackString(0, 0)})
		Bytes(m)
		if m.Pop().Type != value.TypeBytes {
			t.Errorf("bytes failed")
		}

		m.Push(value.Value{Type: value.TypeString, Data: value.PackString(0, 0)})
		ByteArray(m)
		if m.Pop().Type != value.TypeBytes {
			t.Errorf("bytearray failed")
		}
	})

	t.Run("TruthyAndBool", func(t *testing.T) {
		m.Reset()
		// Test Bool built-in
		m.Push(value.Value{Type: value.TypeInt, Data: 10})
		Bool(m)
		if m.Pop().Data != 1 {
			t.Errorf("bool(10) failed")
		}
	})

	t.Run("LenAllTypes", func(t *testing.T) {
		m.Reset()
		// String
		m.Push(value.Value{Type: value.TypeString, Data: 5})
		Len(m)
		if m.Pop().Int() != 5 {
			t.Errorf("len(str) failed")
		}
		// List
		list := []value.Value{{}, {}}
		m.Push(value.Value{Type: value.TypeList, Opaque: &list})
		Len(m)
		if m.Pop().Int() != 2 {
			t.Errorf("len(list) failed")
		}
		// Dict
		m.Push(value.Value{Type: value.TypeDict, Opaque: map[string]any{"a": 1}})
		Len(m)
		if m.Pop().Int() != 1 {
			t.Errorf("len(dict) failed")
		}
		// Error
		m.Push(value.Value{Type: value.TypeInt})
		if err := Len(m); err == nil {
			t.Errorf("expected error for len(int)")
		}
	})

	t.Run("ListIdentity", func(t *testing.T) {
		m.Reset()
		list := []value.Value{}
		l := value.Value{Type: value.TypeList, Opaque: &list}
		m.Push(l)
		List(m)
		if m.Pop().Type != value.TypeList {
			t.Errorf("list identity failed")
		}

		m.Push(value.Value{Type: value.TypeInt})
		if err := List(m); err == nil {
			t.Errorf("expected error for list(int)")
		}
	})

	t.Run("StrAllTypes", func(t *testing.T) {
		m.Reset()
		m.Push(value.Value{Type: value.TypeBool, Data: 1})
		Str(m)
		if value.UnpackString(m.Pop().Data, m.Arena) != "True" {
			t.Errorf("str(bool) failed")
		}

		m.Push(value.Value{Type: value.TypeVoid})
		Str(m)
		m.Pop()
	})

	t.Run("IntErrors", func(t *testing.T) {
		m.Reset()
		m.Push(value.Value{Type: value.TypeString, Data: uint64(len(m.Arena))})
		m.Arena = append(m.Arena, "abc"...)
		m.Push(value.Value{Type: value.TypeString, Data: value.PackString(uint32(len(m.Arena)-3), 3)})
		if err := Int(m); err == nil {
			t.Errorf("expected error for int('abc')")
		}

		m.Push(value.Value{Type: value.TypeVoid})
		if err := Int(m); err == nil {
			t.Errorf("expected error for int(void)")
		}
	})

	t.Run("PushConverted", func(t *testing.T) {
		m.Reset()
		pushConverted(m, "hello")
		if m.Pop().Type != value.TypeString {
			t.Errorf("push string failed")
		}
		pushConverted(m, float64(3.14))
		if m.Pop().Type != value.TypeInt {
			t.Errorf("push float64 failed")
		} // Currently pushes as Int
		pushConverted(m, true)
		if m.Pop().Type != value.TypeBool {
			t.Errorf("push bool failed")
		}
		pushConverted(m, struct{ A int }{1})
		if m.Pop().Type != value.TypeString {
			t.Errorf("push default failed")
		}
	})

	t.Run("Callable", func(t *testing.T) {
		m.Reset()
		m.FunctionRegistry["f"] = 0
		m.Push(value.Value{Type: value.TypeString, Data: uint64(len(m.Arena))})
		m.Arena = append(m.Arena, "f"...)
		m.Push(value.Value{Type: value.TypeString, Data: value.PackString(uint32(len(m.Arena)-1), 1)})
		if err := Callable(m); err != nil {
			t.Fatal(err)
		}
		if m.Pop().Data != 1 {
			t.Errorf("callable failed")
		}
	})

	t.Run("Errors", func(t *testing.T) {
		m.Reset()
		m.Push(value.Value{Type: value.TypeString, Data: 0})
		m.Push(value.Value{Type: value.TypeString, Data: 0})
		if err := DivMod(m); err == nil {
			t.Errorf("expected error for divmod(str, str)")
		}

		m.Push(value.Value{Type: value.TypeInt, Data: 1})
		m.Push(value.Value{Type: value.TypeInt, Data: 0})
		if err := DivMod(m); err == nil || !strings.Contains(err.Error(), "ZeroDivisionError") {
			t.Errorf("expected zero division error, got %v", err)
		}

		m.Push(value.Value{Type: value.TypeString, Data: 0})
		m.Push(value.Value{Type: value.TypeInt, Data: 1})
		if err := Round(m); err == nil {
			t.Errorf("expected error for round(str)")
		}

		m.Push(value.Value{Type: value.TypeVoid})
		if err := Float(m); err == nil {
			t.Errorf("expected error for float(void)")
		}

		m.Push(value.Value{Type: value.TypeString, Data: value.PackString(0, 0)})
		if err := Ord(m); err == nil {
			t.Errorf("expected error for ord('')")
		}

		m.Push(value.Value{Type: value.TypeInt, Data: 0})
		if err := Iter(m); err == nil {
			t.Errorf("expected error for iter(int)")
		}

		m.Push(value.Value{Type: value.TypeInt, Data: 0})
		if err := Next(m); err == nil {
			t.Errorf("expected error for next(int)")
		}

		emptyIterList := make([]value.Value, 0)
		m.Push(value.Value{Type: value.TypeIterator, Opaque: &iteratorState{listPtr: &emptyIterList, index: 0}})
		if err := Next(m); err == nil || err.Error() != "stop iteration" {
			t.Errorf("expected stop iteration")
		}
	})

	t.Run("JSONAndStrings", func(t *testing.T) {
		m.Reset()
		// ParseJSON
		jsonStr := `{"a": 1}`
		offset := uint32(len(m.Arena))
		m.Arena = append(m.Arena, []byte(jsonStr)...)
		m.Push(value.Value{Type: value.TypeString, Data: value.PackString(offset, uint32(len(jsonStr)))})
		if err := ParseJSON(m); err != nil {
			t.Fatal(err)
		}
		dict := m.Pop()
		if dict.Type != value.TypeDict {
			t.Errorf("expected dict")
		}

		// GetField
		m.Push(dict)
		key := "a"
		offset = uint32(len(m.Arena))
		m.Arena = append(m.Arena, []byte(key)...)
		m.Push(value.Value{Type: value.TypeString, Data: value.PackString(offset, uint32(len(key)))})
		if err := GetField(m); err != nil {
			t.Fatal(err)
		}
		if m.Pop().Int() != 1 {
			t.Errorf("getfield failed")
		}

		// FormatString
		m.Reset()
		m.Push(value.Value{Type: value.TypeString, Data: uint64(len(m.Arena))})
		m.Arena = append(m.Arena, "val: %s"...)
		m.Push(value.Value{Type: value.TypeString, Data: value.PackString(uint32(len(m.Arena)-7), 7)})
		m.Push(value.Value{Type: value.TypeInt, Data: 42})
		if err := FormatString(m); err != nil {
			t.Fatal(err)
		}
		res := value.UnpackString(m.Pop().Data, m.Arena)
		if res != "val: 42" {
			t.Errorf("format failed: %s", res)
		}

		m.Push(value.Value{Type: value.TypeString, Data: uint64(len(m.Arena))})
		m.Arena = append(m.Arena, "%s"...)
		m.Push(value.Value{Type: value.TypeString, Data: value.PackString(uint32(len(m.Arena)-2), 2)})
		m.Push(value.Value{Type: value.TypeString, Data: value.PackString(0, 3)}) // "123"
		FormatString(m)
		m.Pop()

		m.Push(value.Value{Type: value.TypeString, Data: value.PackString(uint32(len(m.Arena)-2), 2)})
		m.Push(value.Value{Type: value.TypeBool, Data: 1})
		FormatString(m)
		m.Pop()

		m.Push(value.Value{Type: value.TypeString, Data: value.PackString(uint32(len(m.Arena)-2), 2)})
		m.Push(value.Value{Type: value.TypeVoid})
		FormatString(m)
		m.Pop()

		// GetField missing key
		m.Reset()
		m.Push(value.Value{Type: value.TypeDict, Opaque: map[string]any{"a": 1}})
		key = "missing"
		offset = uint32(len(m.Arena))
		m.Arena = append(m.Arena, []byte(key)...)
		m.Push(value.Value{Type: value.TypeString, Data: value.PackString(offset, uint32(len(key)))})
		GetField(m)
		if m.Pop().Type != value.TypeVoid {
			t.Errorf("expected void")
		}

		// IsEmpty
		m.Push(value.Value{Type: value.TypeString, Data: 0})
		IsEmpty(m)
		if m.Pop().Data != 1 {
			t.Errorf("isempty str failed")
		}

		// ParseJSONKey
		m.Reset()
		jsonStr = `{"key": "val"}`
		offset = uint32(len(m.Arena))
		m.Arena = append(m.Arena, []byte(jsonStr)...)
		m.Push(value.Value{Type: value.TypeString, Data: value.PackString(offset, uint32(len(jsonStr)))})
		key = "key"
		offset = uint32(len(m.Arena))
		m.Arena = append(m.Arena, []byte(key)...)
		m.Push(value.Value{Type: value.TypeString, Data: value.PackString(offset, uint32(len(key)))})
		if err := ParseJSONKey(m); err != nil {
			t.Fatal(err)
		}
		if value.UnpackString(m.Pop().Data, m.Arena) != "val" {
			t.Errorf("parsejsonkey failed")
		}

		// Missing key
		m.Push(value.Value{Type: value.TypeString, Data: value.PackString(offset-uint32(len(jsonStr)), uint32(len(jsonStr)))})
		key = "missing"
		offset = uint32(len(m.Arena))
		m.Arena = append(m.Arena, []byte(key)...)
		m.Push(value.Value{Type: value.TypeString, Data: value.PackString(offset, uint32(len(key)))})
		if err := ParseJSONKey(m); err != nil {
			t.Fatal(err)
		}
		if m.Pop().Type != value.TypeVoid {
			t.Errorf("expected void for missing key")
		}

		// Invalid JSON
		m.Push(value.Value{Type: value.TypeString, Data: uint64(len(m.Arena))})
		m.Arena = append(m.Arena, "invalid"...)
		m.Push(value.Value{Type: value.TypeString, Data: value.PackString(uint32(len(m.Arena)-7), 7)})
		m.Push(value.Value{Type: value.TypeString, Data: 0})
		if err := ParseJSONKey(m); err == nil {
			t.Errorf("expected error for invalid json")
		}
	})

	t.Run("GetItem", func(t *testing.T) {
		m.Reset()
		list := []value.Value{{Type: value.TypeInt, Data: 42}}
		m.Push(value.Value{Type: value.TypeList, Opaque: &list})
		m.Push(value.Value{Type: value.TypeInt, Data: 0})
		if err := GetItem(m); err != nil {
			t.Fatal(err)
		}
		if m.Pop().Int() != 42 {
			t.Errorf("getitem list failed")
		}

		m.Push(value.Value{Type: value.TypeDict, Opaque: map[string]any{"k": "v"}})
		m.Push(value.Value{Type: value.TypeString, Data: value.PackString(uint32(len(m.Arena)), 1)})
		m.Arena = append(m.Arena, 'k')
		if err := GetItem(m); err != nil {
			t.Fatal(err)
		}
		if m.Pop().Type != value.TypeString {
			t.Errorf("getitem dict failed")
		}
	})

	t.Run("ChrOrd", func(t *testing.T) {
		m.Reset()
		m.Push(value.Value{Type: value.TypeInt, Data: 65})
		if err := Chr(m); err != nil {
			t.Fatal(err)
		}
		sVal := m.Pop()
		s := value.UnpackString(sVal.Data, m.Arena)
		if s != "A" {
			t.Errorf("chr(65) = %s", s)
		}

		m.Push(sVal) // Uses valid offset into arena
		if err := Ord(m); err != nil {
			t.Fatal(err)
		}
		if m.Pop().Int() != 65 {
			t.Errorf("ord('A') failed")
		}
	})

	t.Run("LenListRange", func(t *testing.T) {
		m.Reset()
		m.Push(value.Value{Type: value.TypeInt, Data: 5})
		m.Push(value.Value{Type: value.TypeInt, Data: 1})
		if err := Range(m); err != nil {
			t.Fatal(err)
		}
		if err := Len(m); err != nil {
			t.Fatal(err)
		}
		if m.Pop().Int() != 5 {
			t.Errorf("len(range(5)) failed")
		}
	})

	t.Run("SumMaxMin", func(t *testing.T) {
		m.Reset()
		m.Push(value.Value{Type: value.TypeInt, Data: 3})
		m.Push(value.Value{Type: value.TypeInt, Data: 1})
		Range(m)                // [0, 1, 2]
		m.Push(m.Stack[m.SP-1]) // Dup
		m.Push(value.Value{Type: value.TypeInt, Data: 1})
		if err := Sum(m); err != nil {
			t.Fatal(err)
		}
		if m.Pop().Int() != 3 {
			t.Errorf("sum([0,1,2]) failed")
		}

		m.Push(value.Value{Type: value.TypeInt, Data: 1})
		if err := Max(m); err != nil {
			t.Fatal(err)
		}
		if m.Pop().Int() != 2 {
			t.Errorf("max([0,1,2]) failed")
		}
	})

	t.Run("AllAny", func(t *testing.T) {
		m.Reset()
		m.Push(value.Value{Type: value.TypeInt, Data: 2})
		m.Push(value.Value{Type: value.TypeInt, Data: 1})
		Range(m) // [0, 1]
		m.Push(m.Stack[m.SP-1])
		All(m)
		if m.Pop().Data != 0 {
			t.Errorf("all([0,1]) should be false")
		}
		Any(m)
		if m.Pop().Data != 1 {
			t.Errorf("any([0,1]) should be true")
		}
	})

	t.Run("Conversions", func(t *testing.T) {
		m.Reset()
		// Int identity
		m.Push(value.Value{Type: value.TypeInt, Data: 42})
		Int(m)
		if m.Pop().Int() != 42 {
			t.Errorf("int identity failed")
		}

		m.Push(value.Value{Type: value.TypeString, Data: uint64(len(m.Arena))})
		m.Arena = append(m.Arena, "123"...)
		m.Push(value.Value{Type: value.TypeString, Data: value.PackString(uint32(len(m.Arena)-3), 3)})
		if err := Int(m); err != nil {
			t.Fatal(err)
		}
		if m.Pop().Int() != 123 {
			t.Errorf("int('123') failed")
		}

		// Str identity
		m.Push(value.Value{Type: value.TypeString, Data: value.PackString(0, 3)})
		Str(m)
		if m.Pop().Type != value.TypeString {
			t.Errorf("str identity failed")
		}

		m.Push(value.Value{Type: value.TypeInt, Data: 456})
		Str(m)
		if value.UnpackString(m.Pop().Data, m.Arena) != "456" {
			t.Errorf("str(456) failed")
		}
	})

	t.Run("Functional", func(t *testing.T) {
		m.Reset()
		// Mock a function: def double(x): return x * 2
		m.FunctionRegistry["double"] = 0
		m.Code = []uint32{
			(uint32(vm.OP_PUSH_L) << 24), // Push local 0
			(uint32(vm.OP_PUSH_C) << 24), // Push 2
			(uint32(vm.OP_MUL) << 24),    // Multiply
			(uint32(vm.OP_RET) << 24),    // Return result
		}
		m.Constants = []value.Value{{Type: value.TypeInt, Data: 2}}

		// Map
		l1 := []value.Value{{Type: value.TypeInt, Data: 1}, {Type: value.TypeInt, Data: 2}}
		funcName := "double"
		offset := uint32(len(m.Arena))
		m.Arena = append(m.Arena, []byte(funcName)...)
		m.Push(value.Value{Type: value.TypeString, Data: value.PackString(offset, uint32(len(funcName)))})
		m.Push(value.Value{Type: value.TypeList, Opaque: &l1})

		if err := Map(m); err != nil {
			t.Fatal(err)
		}
		resVal := m.Pop()
		if resVal.Type != value.TypeIterator {
			t.Errorf("expected iterator")
		}
		res := *(resVal.Opaque.(*iteratorState).listPtr)
		if len(res) != 2 || res[0].Int() != 2 || res[1].Int() != 4 {
			t.Errorf("map double failed, got %v", res)
		}

		// Filter
		m.Reset()
		m.FunctionRegistry["is_one"] = 0
		m.Code = []uint32{
			(uint32(vm.OP_PUSH_L) << 24), // Push Arg 0
			(uint32(vm.OP_PUSH_C) << 24), // Push 1
			(uint32(vm.OP_EQ) << 24),     // Equal?
			(uint32(vm.OP_RET) << 24),    // Return bool
		}
		m.Constants = []value.Value{{Type: value.TypeInt, Data: 1}}

		funcName = "is_one"
		offset = uint32(len(m.Arena))
		m.Arena = append(m.Arena, []byte(funcName)...)
		m.Push(value.Value{Type: value.TypeString, Data: value.PackString(offset, uint32(len(funcName)))})
		m.Push(value.Value{Type: value.TypeList, Opaque: &l1})

		if err := Filter(m); err != nil {
			t.Fatal(err)
		}
		resVal = m.Pop()
		if resVal.Type != value.TypeIterator {
			t.Errorf("expected iterator")
		}
		res = *(resVal.Opaque.(*iteratorState).listPtr)
		if len(res) != 1 || res[0].Int() != 1 {
			t.Errorf("filter is_one failed, got %v", res)
		}
	})

	t.Run("Math", func(t *testing.T) {
		m.Reset()
		m.Push(value.Value{Type: value.TypeInt, Data: 2})
		m.Push(value.Value{Type: value.TypeInt, Data: 3})
		Pow(m)
		if m.Pop().Int() != 8 {
			t.Errorf("pow(2,3) failed")
		}

		var v5 value.Value
		v5.SetInt(-5)
		m.Push(v5)
		Abs(m)
		if m.Pop().Int() != 5 {
			t.Errorf("abs(-5) failed")
		}
	})

	t.Run("Types", func(t *testing.T) {
		m.Reset()
		types := []value.Type{value.TypeInt, value.TypeFloat, value.TypeBool, value.TypeString, value.TypeList, value.TypeTuple, value.TypeDict, value.TypeSet, value.TypeVoid}
		for _, ty := range types {
			m.Push(value.Value{Type: ty})
			TypeWord(m)
			m.Pop()
		}
	})

	t.Run("FloatAndRoundExtended", func(t *testing.T) {
		m.Reset()
		// Float identity
		m.Push(value.Value{Type: value.TypeFloat, Data: math.Float64bits(1.1)})
		Float(m)
		if m.Pop().Type != value.TypeFloat {
			t.Errorf("float identity failed")
		}

		// Round Float
		m.Push(value.Value{Type: value.TypeFloat, Data: math.Float64bits(1.1)})
		m.Push(value.Value{Type: value.TypeInt, Data: 1})
		Round(m)
		if m.Pop().Int() != 1 {
			t.Errorf("round float failed")
		}

		// Round Int
		m.Push(value.Value{Type: value.TypeInt, Data: 10})
		m.Push(value.Value{Type: value.TypeInt, Data: 1})
		Round(m)
		if m.Pop().Int() != 10 {
			t.Errorf("round int failed")
		}
	})

	t.Run("BoolExtended", func(t *testing.T) {
		m.Reset()
		// Float
		m.Push(value.Value{Type: value.TypeFloat, Data: math.Float64bits(1.1)})
		Bool(m)
		if m.Pop().Data != 1 {
			t.Errorf("bool(1.1) failed")
		}
		m.Push(value.Value{Type: value.TypeFloat, Data: 0})
		Bool(m)
		if m.Pop().Data != 0 {
			t.Errorf("bool(0.0) failed")
		}

		// Set
		m.Push(value.Value{Type: value.TypeSet, Opaque: map[any]struct{}{1: {}}})
		Bool(m)
		if m.Pop().Data != 1 {
			t.Errorf("bool(set) failed")
		}
	})

	t.Run("StrExtended", func(t *testing.T) {
		m.Reset()
		// Float
		m.Push(value.Value{Type: value.TypeFloat, Data: math.Float64bits(1.1)})
		Str(m)
		if value.UnpackString(m.Pop().Data, m.Arena) == "" {
			t.Errorf("str(float) failed")
		}

		// Dict
		m.Push(value.Value{Type: value.TypeDict, Opaque: map[string]any{"a": 1}})
		Str(m)
		if value.UnpackString(m.Pop().Data, m.Arena) == "" {
			t.Errorf("str(dict) failed")
		}
	})

	t.Run("IsEmptyExtended", func(t *testing.T) {
		m.Reset()
		// Void
		m.Push(value.Value{Type: value.TypeVoid})
		IsEmpty(m)
		if m.Pop().Data != 1 {
			t.Errorf("isempty void failed")
		}

		// List
		m.Push(value.Value{Type: value.TypeList, Opaque: []value.Value{}})
		IsEmpty(m)
		if m.Pop().Data != 1 {
			t.Errorf("isempty list failed")
		}

		// Map
		m.Push(value.Value{Type: value.TypeDict, Opaque: map[string]any{}})
		IsEmpty(m)
		if m.Pop().Data != 1 {
			t.Errorf("isempty map failed")
		}
	})

	t.Run("MakeList", func(t *testing.T) {
		m.Reset()
		m.Push(value.Value{Type: value.TypeInt, Data: 10})
		m.Push(value.Value{Type: value.TypeInt, Data: 20})
		m.Push(value.Value{Type: value.TypeInt, Data: 2}) // Count
		MakeList(m)
		res := *(m.Pop().Opaque.(*[]value.Value))
		if len(res) != 2 || res[0].Int() != 10 || res[1].Int() != 20 {
			t.Errorf("makelist failed")
		}
	})

	t.Run("EmptyCollections", func(t *testing.T) {
		m.Reset()
		emptySlice := []value.Value{}
		// Sum empty
		m.Push(value.Value{Type: value.TypeList, Opaque: &emptySlice})
		m.Push(value.Value{Type: value.TypeInt, Data: 1})
		Sum(m)
		if m.Pop().Int() != 0 {
			t.Errorf("sum empty failed")
		}
		// All empty
		m.Push(value.Value{Type: value.TypeList, Opaque: &emptySlice})
		All(m)
		if m.Pop().Data != 1 {
			t.Errorf("all empty failed")
		}
		// Any empty
		m.Push(value.Value{Type: value.TypeList, Opaque: &emptySlice})
		Any(m)
		if m.Pop().Data != 0 {
			t.Errorf("any empty failed")
		}
	})
}
