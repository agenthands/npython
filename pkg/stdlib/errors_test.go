package stdlib

import (
	"testing"

	"github.com/agenthands/npython/pkg/core/value"
	"github.com/agenthands/npython/pkg/vm"
)

func TestBuiltinErrors(t *testing.T) {
	m := vm.GetMachine()
	defer vm.PutMachine(m)

	tests := []struct {
		name string
		fn   func(*vm.Machine) error
		args []value.Value
	}{
		{"LenErr", Len, []value.Value{{Type: value.TypeInt}}},
		{"RangeErr", Range, []value.Value{{Type: value.TypeString}, {Type: value.TypeInt, Data: 1}}},
		{"SumErr", Sum, []value.Value{{Type: value.TypeInt}}},
		{"MaxErr", Max, []value.Value{{Type: value.TypeInt}}},
		{"MaxEmpty", Max, []value.Value{{Type: value.TypeList, Opaque: &[]value.Value{}}}},
		{"MinErr", Min, []value.Value{{Type: value.TypeInt}}},
		{"MinEmpty", Min, []value.Value{{Type: value.TypeList, Opaque: &[]value.Value{}}}},
		{"MapErr1", Map, []value.Value{{Type: value.TypeInt}, {Type: value.TypeInt}}},
		{"MapErr2", Map, []value.Value{{Type: value.TypeString}, {Type: value.TypeInt}}},
		{"MapErr3", Map, []value.Value{{Type: value.TypeString, Data: 0}, {Type: value.TypeList, Opaque: &[]value.Value{}}}}, // Unknown func
		{"AbsErr", Abs, []value.Value{{Type: value.TypeString}}},
		{"IntErr", Int, []value.Value{{Type: value.TypeVoid}}},
		{"FloatErr", Float, []value.Value{{Type: value.TypeVoid}}},
		{"FloatParseErr", Float, []value.Value{{Type: value.TypeString, Data: uint64(len(m.Arena))}}}, // Needs "abc" in arena
		{"FilterErr1", Filter, []value.Value{{Type: value.TypeInt}, {Type: value.TypeInt}}},
		{"FilterErr2", Filter, []value.Value{{Type: value.TypeString}, {Type: value.TypeInt}}},
		{"PowErr", Pow, []value.Value{{Type: value.TypeString}, {Type: value.TypeInt}}},
		{"GetItemErr1", GetItem, []value.Value{{Type: value.TypeList, Opaque: &[]value.Value{}}, {Type: value.TypeString}}},
		{"GetItemErr2", GetItem, []value.Value{{Type: value.TypeList, Opaque: &[]value.Value{}}, {Type: value.TypeInt, Data: 0}}},
		{"AllErr", All, []value.Value{{Type: value.TypeInt}}},
		{"AnyErr", Any, []value.Value{{Type: value.TypeInt}}},
		{"BinErr", Bin, []value.Value{{Type: value.TypeString}}},
		{"OctErr", Oct, []value.Value{{Type: value.TypeString}}},
		{"HexErr", Hex, []value.Value{{Type: value.TypeString}}},
		{"ChrErr", Chr, []value.Value{{Type: value.TypeString}}},
		{"OrdErr1", Ord, []value.Value{{Type: value.TypeInt}}},
		{"TupleErr", Tuple, []value.Value{{Type: value.TypeInt}}},
		{"SetErr", Set, []value.Value{{Type: value.TypeInt}}},
		{"ReversedErr", Reversed, []value.Value{{Type: value.TypeInt}}},
		{"SortedErr", Sorted, []value.Value{{Type: value.TypeInt}}},
		{"ZipErr", Zip, []value.Value{{Type: value.TypeInt}, {Type: value.TypeInt}}},
		{"EnumerateErr", Enumerate, []value.Value{{Type: value.TypeInt}}},
		{"IterErr", Iter, []value.Value{{Type: value.TypeInt}}},
		{"NextErr", Next, []value.Value{{Type: value.TypeInt}}},
		{"NextEmpty", Next, []value.Value{{Type: value.TypeIterator, Opaque: &iteratorState{listPtr: &[]value.Value{}, index: 0}}}},
		{"ParseJSONErr", ParseJSON, []value.Value{{Type: value.TypeString, Data: 0}}}, // Needs invalid JSON in arena
		{"GetFieldErr", GetField, []value.Value{{Type: value.TypeInt}, {Type: value.TypeString}}},
		{"CheckStatusErr", func(m *vm.Machine) error { sandbox := NewHTTPSandbox(nil); return sandbox.CheckStatus(m) }, []value.Value{{Type: value.TypeInt}}},
		{"SetURLErr", func(m *vm.Machine) error { sandbox := NewHTTPSandbox(nil); return sandbox.SetURL(m) }, []value.Value{{Type: value.TypeString, Data: 0}}},
		{"SetMethodErr", func(m *vm.Machine) error { sandbox := NewHTTPSandbox(nil); return sandbox.SetMethod(m) }, []value.Value{{Type: value.TypeString, Data: 0}}},
		{"SendRequestErr", func(m *vm.Machine) error { sandbox := NewHTTPSandbox(nil); return sandbox.SendRequest(m) }, nil},
		{"SendRequestDomainErr", func(m *vm.Machine) error {
			sandbox := NewHTTPSandbox([]string{"allowed.com"})
			sandbox.WithClient(m)
			m.Arena = append(m.Arena, "http://blocked.com"...)
			m.Push(value.Value{Type: value.TypeString, Data: value.PackString(0, 18)})
			sandbox.SetURL(m)
			return sandbox.SendRequest(m)
		}, nil},
		{"WriteFileEscapeErr", func(m *vm.Machine) error {
			sandbox := NewFSSandbox("/tmp", 100)
			m.Arena = append(m.Arena, "content"...)
			m.Arena = append(m.Arena, "../escape.txt"...)
			m.Push(value.Value{Type: value.TypeString, Data: value.PackString(0, 7)})
			m.Push(value.Value{Type: value.TypeString, Data: value.PackString(7, 13)})
			return sandbox.WriteFile(m)
		}, nil},
		{"WriteFileLargeErr", func(m *vm.Machine) error {
			sandbox := NewFSSandbox("/tmp", 1)
			m.Arena = append(m.Arena, "too large"...)
			m.Arena = append(m.Arena, "ok.txt"...)
			m.Push(value.Value{Type: value.TypeString, Data: value.PackString(0, 9)})
			m.Push(value.Value{Type: value.TypeString, Data: value.PackString(9, 6)})
			return sandbox.WriteFile(m)
		}, nil},
		{"ReadFileEscapeErr", func(m *vm.Machine) error {
			sandbox := NewFSSandbox("/tmp", 100)
			m.Arena = append(m.Arena, "../escape.txt"...)
			m.Push(value.Value{Type: value.TypeString, Data: value.PackString(0, 13)})
			return sandbox.ReadFile(m)
		}, nil},
		{"FetchURLErr", func(m *vm.Machine) error {
			sandbox := NewHTTPSandbox([]string{"allowed.com"})
			m.Arena = append(m.Arena, ":::"...)
			m.Push(value.Value{Type: value.TypeString, Data: value.PackString(0, 3)})
			return sandbox.Fetch(m)
		}, nil},
		{"SendRequestURLErr", func(m *vm.Machine) error {
			sandbox := NewHTTPSandbox([]string{"allowed.com"})
			sandbox.WithClient(m)
			m.Arena = append(m.Arena, ":::"...)
			m.Push(value.Value{Type: value.TypeString, Data: value.PackString(0, 3)})
			sandbox.SetURL(m)
			return sandbox.SendRequest(m)
		}, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m.Reset()
			for _, arg := range tt.args {
				if tt.name == "FloatParseErr" && arg.Type == value.TypeString {
					m.Arena = append(m.Arena, "abc"...)
					arg.Data = value.PackString(0, 3)
				}
				if tt.name == "ParseJSONErr" && arg.Type == value.TypeString {
					m.Arena = append(m.Arena, "!!!"...)
					arg.Data = value.PackString(0, 3)
				}
				m.Push(arg)
			}
			if err := tt.fn(m); err == nil {
				t.Errorf("%s: expected error, got nil", tt.name)
			}
		})
	}
}
