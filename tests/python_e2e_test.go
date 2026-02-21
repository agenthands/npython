package main_test

import (
	"testing"

	"github.com/agenthands/npython/pkg/compiler/python"
	"github.com/agenthands/npython/pkg/core/value"
	"github.com/agenthands/npython/pkg/stdlib"
	"github.com/agenthands/npython/pkg/vm"
)

// TestComprehensive checks edge cases and complex logic in the nPython subset.
func TestComprehensive(t *testing.T) {
	tests := []struct {
		name   string
		src    string
		verify func(m *vm.Machine, t *testing.T)
	}{
		{
			name: "Factorial (While Loop)",
			src: `
n = 5
res = 1
while n > 0:
    res = res * n
    n = n - 1
`,
			verify: func(m *vm.Machine, t *testing.T) {
				res := m.Frames[0].Locals[1].Int()
				if res != 120 {
					t.Errorf("Expected factorial(5) = 120, got %d", res)
				}
			},
		},
		{
			name: "Nested If/Else",
			src: `
val = 50
res = 0
if val > 10:
    if val > 40:
        res = 2
    else:
        res = 1
else:
    res = 0
`,
			verify: func(m *vm.Machine, t *testing.T) {
				res := m.Frames[0].Locals[1].Int()
				if res != 2 {
					t.Errorf("Expected res = 2, got %d", res)
				}
			},
		},
		{
			name: "Fibonacci (Iterative)",
			src: `
n = 10
a = 0
b = 1
temp = 0
i = 0
while i < n:
    temp = a
    a = b
    b = temp + b
    i = i + 1
`,
			verify: func(m *vm.Machine, t *testing.T) {
				res := m.Frames[0].Locals[1].Int() // a is local 1
				if res != 55 {
					t.Errorf("Expected fib(10) = 55, got %d", res)
				}
			},
		},
		{
			name: "String Concatenation",
			src: `
s1 = "Hello"
s2 = " "
s3 = "World"
res = s1 + s2 + s3
`,
			verify: func(m *vm.Machine, t *testing.T) {
				resVal := m.Frames[0].Locals[3]
				res := value.UnpackString(resVal.Data, m.Arena)
				if res != "Hello World" {
					t.Errorf("Expected 'Hello World', got '%s'", res)
				}
			},
		},
		{
			name: "Comparison Logic",
			src: `
a = 10
b = 20
c = 10
res1 = 0
res2 = 0
if a == c:
    res1 = 1
if a != b:
    res2 = 1
`,
			verify: func(m *vm.Machine, t *testing.T) {
				if m.Frames[0].Locals[3].Int() != 1 {
					t.Errorf("Expected res1 = 1")
				}
				if m.Frames[0].Locals[4].Int() != 1 {
					t.Errorf("Expected res2 = 1")
				}
			},
		},
		{
			name: "Python Built-ins (len, range, sum, max, min)",
			src: `
items = range(5)
l = len(items)
s = sum(items)
m1 = max(items)
m2 = min(items)
sl = len("hello")
`,
			verify: func(m *vm.Machine, t *testing.T) {
				lVal := m.Frames[0].Locals[1].Int()
				if lVal != 5 {
					t.Errorf("Expected len(range(5)) = 5, got %d", lVal)
				}

				sVal := m.Frames[0].Locals[2].Int()
				if sVal != 10 {
					t.Errorf("Expected sum(range(5)) = 10, got %d", sVal)
				}

				m1Val := m.Frames[0].Locals[3].Int()
				if m1Val != 4 {
					t.Errorf("Expected max(range(5)) = 4, got %d", m1Val)
				}

				m2Val := m.Frames[0].Locals[4].Int()
				if m2Val != 0 {
					t.Errorf("Expected min(range(5)) = 0, got %d", m2Val)
				}

				slVal := m.Frames[0].Locals[5].Int()
				if slVal != 5 {
					t.Errorf("Expected len('hello') = 5, got %d", slVal)
				}
			},
		},
		{
			name: "Function Definition and Map",
			src: `
def double(x):
    return x * 2

items = range(3)
res = map(double, items)
s = sum(res)
`,
			verify: func(m *vm.Machine, t *testing.T) {
				sVal := m.Frames[0].Locals[2].Int()
				if sVal != 6 {
					t.Errorf("Expected sum(map(double, items)) = 6, got %d", sVal)
				}
			},
		},
		{
			name: "Conversions and Math (abs, bool, str, int)",
			src: `
a = -10
res_abs = abs(a)
b1 = bool(1)
b2 = bool(0)
s1 = str(123)
s2 = str(b1)
i1 = int("456")
`,
			verify: func(m *vm.Machine, t *testing.T) {
				if m.Frames[0].Locals[1].Int() != 10 {
					t.Errorf("abs(-10) failed")
				}
				if m.Frames[0].Locals[2].Data != 1 {
					t.Errorf("bool(1) failed")
				}
				if m.Frames[0].Locals[3].Data != 0 {
					t.Errorf("bool(0) failed")
				}

				s1Val := m.Frames[0].Locals[4]
				if value.UnpackString(s1Val.Data, m.Arena) != "123" {
					t.Errorf("str(123) failed")
				}

				s2Val := m.Frames[0].Locals[5]
				if value.UnpackString(s2Val.Data, m.Arena) != "True" {
					t.Errorf("str(True) failed")
				}

				if m.Frames[0].Locals[6].Int() != 456 {
					t.Errorf("int('456') failed")
				}
			},
		},
		{
			name: "Filter and Pow",
			src: `
def is_even(x):
    half = x / 2
    back = half * 2
    return x == back

items = range(6)
evens = filter(is_even, items)
l = len(evens)
p = pow(2, 3)
`,
			verify: func(m *vm.Machine, t *testing.T) {
				lVal := m.Frames[0].Locals[2].Int()
				if lVal != 3 {
					t.Errorf("Expected 3 evens, got %d", lVal)
				}

				pVal := m.Frames[0].Locals[3].Int()
				if pVal != 8 {
					t.Errorf("Expected pow(2, 3) = 8, got %d", pVal)
				}
			},
		},
		{
			name: "All, Any and List Literals",
			src: `
list1 = range(3) # [0, 1, 2]
list2 = [1, 2, 3]
a1 = all(list1) # false
a2 = any(list1) # true
a3 = all(list2) # true
`,
			verify: func(m *vm.Machine, t *testing.T) {
				if m.Frames[0].Locals[2].Data != 0 {
					t.Errorf("all(range(3)) should be false")
				}
				if m.Frames[0].Locals[3].Data != 1 {
					t.Errorf("any(range(3)) should be true")
				}
				if m.Frames[0].Locals[4].Data != 1 {
					t.Errorf("all([1,2,3]) should be true")
				}
			},
		},
		{
			name: "Indexing",
			src: `
items = [10, 20, 30]
x = items[1]
`,
			verify: func(m *vm.Machine, t *testing.T) {
				xVal := m.Frames[0].Locals[1].Int()
				if xVal != 20 {
					t.Errorf("Expected items[1] = 20, got %d", xVal)
				}
			},
		},
		{
			name: "Advanced Python Built-ins",
			src: `
def double(x): return x * 2
dm = divmod(10, 3)
q = dm[0]
r = dm[1]
h = hex(255)
f = float("3.14")
c = chr(65)
o = ord("A")
t1 = type(10)
is_f = callable("double")
`,
			verify: func(m *vm.Machine, t *testing.T) {
				if m.Frames[0].Locals[1].Int() != 3 {
					t.Errorf("divmod q failed")
				}
				if m.Frames[0].Locals[2].Int() != 1 {
					t.Errorf("divmod r failed")
				}
				if value.UnpackString(m.Frames[0].Locals[3].Data, m.Arena) != "0xff" {
					t.Errorf("hex failed")
				}
				if value.UnpackString(m.Frames[0].Locals[7].Data, m.Arena) != "int" {
					t.Errorf("type failed")
				}
				if m.Frames[0].Locals[8].Data != 1 {
					t.Errorf("callable failed")
				}
			},
		},
		{
			name: "List Modification",
			src: `
items = [10, 20, 30]
items[1] = 99
val = items[1]
`,
			verify: func(m *vm.Machine, t *testing.T) {
				val := m.Frames[0].Locals[1].Int() // items=0, val=1
				if val != 99 {
					t.Errorf("Expected 99, got %d", val)
				}
			},
		},
		{
			name: "Iterators (iter and next)",
			src: `
items = [10, 20]
it = iter(items)
v1 = next(it)
v2 = next(it)
`,
			verify: func(m *vm.Machine, t *testing.T) {
				if m.Frames[0].Locals[2].Int() != 10 {
					t.Errorf("next 1 failed")
				}
				if m.Frames[0].Locals[3].Int() != 20 {
					t.Errorf("next 2 failed")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			compiler := python.NewCompiler()
			bytecode, err := compiler.Compile(tt.src)
			if err != nil {
				t.Fatalf("Compilation failed: %v", err)
			}

			machine := vm.GetMachine()
			defer vm.PutMachine(machine)

			machine.Code = bytecode.Instructions
			machine.Constants = bytecode.Constants
			machine.Arena = bytecode.Arena

			for name, ip := range bytecode.Functions {
				machine.FunctionRegistry[name] = ip
			}

			machine.HostRegistry = make([]vm.HostFunctionEntry, 100)
			for idx := range machine.HostRegistry {
				machine.HostRegistry[idx] = vm.HostFunctionEntry{Fn: func(m *vm.Machine) error { return nil }}
			}
			machine.HostRegistry[14] = vm.HostFunctionEntry{Fn: stdlib.Len}
			machine.HostRegistry[15] = vm.HostFunctionEntry{Fn: stdlib.Range}
			machine.HostRegistry[16] = vm.HostFunctionEntry{Fn: stdlib.List}
			machine.HostRegistry[17] = vm.HostFunctionEntry{Fn: stdlib.Sum}
			machine.HostRegistry[18] = vm.HostFunctionEntry{Fn: stdlib.Max}
			machine.HostRegistry[19] = vm.HostFunctionEntry{Fn: stdlib.Min}
			machine.HostRegistry[20] = vm.HostFunctionEntry{Fn: stdlib.Map}
			machine.HostRegistry[21] = vm.HostFunctionEntry{Fn: stdlib.Abs}
			machine.HostRegistry[22] = vm.HostFunctionEntry{Fn: stdlib.Bool}
			machine.HostRegistry[23] = vm.HostFunctionEntry{Fn: stdlib.Int}
			machine.HostRegistry[24] = vm.HostFunctionEntry{Fn: stdlib.Str}
			machine.HostRegistry[25] = vm.HostFunctionEntry{Fn: stdlib.Filter}
			machine.HostRegistry[26] = vm.HostFunctionEntry{Fn: stdlib.Pow}
			machine.HostRegistry[27] = vm.HostFunctionEntry{Fn: stdlib.All}
			machine.HostRegistry[28] = vm.HostFunctionEntry{Fn: stdlib.Any}
			machine.HostRegistry[29] = vm.HostFunctionEntry{Fn: stdlib.MakeList}
			machine.HostRegistry[30] = vm.HostFunctionEntry{Fn: stdlib.GetItem}
			machine.HostRegistry[31] = vm.HostFunctionEntry{Fn: stdlib.SetItem}
			machine.HostRegistry[32] = vm.HostFunctionEntry{Fn: stdlib.DivMod}
			machine.HostRegistry[33] = vm.HostFunctionEntry{Fn: stdlib.Round}
			machine.HostRegistry[34] = vm.HostFunctionEntry{Fn: stdlib.Float}
			machine.HostRegistry[35] = vm.HostFunctionEntry{Fn: stdlib.Bin}
			machine.HostRegistry[36] = vm.HostFunctionEntry{Fn: stdlib.Oct}
			machine.HostRegistry[37] = vm.HostFunctionEntry{Fn: stdlib.Hex}
			machine.HostRegistry[38] = vm.HostFunctionEntry{Fn: stdlib.Chr}
			machine.HostRegistry[39] = vm.HostFunctionEntry{Fn: stdlib.Ord}
			machine.HostRegistry[40] = vm.HostFunctionEntry{Fn: stdlib.Dict}
			machine.HostRegistry[41] = vm.HostFunctionEntry{Fn: stdlib.Tuple}
			machine.HostRegistry[42] = vm.HostFunctionEntry{Fn: stdlib.Set}
			machine.HostRegistry[43] = vm.HostFunctionEntry{Fn: stdlib.Reversed}
			machine.HostRegistry[44] = vm.HostFunctionEntry{Fn: stdlib.Sorted}
			machine.HostRegistry[45] = vm.HostFunctionEntry{Fn: stdlib.Zip}
			machine.HostRegistry[46] = vm.HostFunctionEntry{Fn: stdlib.Enumerate}
			machine.HostRegistry[47] = vm.HostFunctionEntry{Fn: stdlib.Repr}
			machine.HostRegistry[48] = vm.HostFunctionEntry{Fn: stdlib.Ascii}
			machine.HostRegistry[49] = vm.HostFunctionEntry{Fn: stdlib.Hash}
			machine.HostRegistry[50] = vm.HostFunctionEntry{Fn: stdlib.Id}
			machine.HostRegistry[51] = vm.HostFunctionEntry{Fn: stdlib.TypeWord}
			machine.HostRegistry[52] = vm.HostFunctionEntry{Fn: stdlib.Callable}
			machine.HostRegistry[53] = vm.HostFunctionEntry{Fn: stdlib.Iter}
			machine.HostRegistry[54] = vm.HostFunctionEntry{Fn: stdlib.Next}
			machine.HostRegistry[55] = vm.HostFunctionEntry{Fn: stdlib.Locals}
			machine.HostRegistry[56] = vm.HostFunctionEntry{Fn: stdlib.Globals}
			machine.HostRegistry[57] = vm.HostFunctionEntry{Fn: stdlib.SliceBuiltin}
			machine.HostRegistry[58] = vm.HostFunctionEntry{Fn: stdlib.Bytes}
			machine.HostRegistry[59] = vm.HostFunctionEntry{Fn: stdlib.ByteArray}
			machine.HostRegistry[60] = vm.HostFunctionEntry{Fn: stdlib.HasNext}
			machine.HostRegistry[63] = vm.HostFunctionEntry{Fn: stdlib.IsInstance}

			err = machine.Run(10000)
			if err != nil {
				t.Fatalf("Runtime error: %v", err)
			}

			tt.verify(machine, t)
		})
	}
}
