package main_test

import (
	"testing"
	"github.com/agenthands/npython/pkg/compiler/python"
	"github.com/agenthands/npython/pkg/vm"
	"github.com/agenthands/npython/pkg/core/value"
	"github.com/agenthands/npython/pkg/stdlib"
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
				if m.Frames[0].Locals[3].Int() != 1 { t.Errorf("Expected res1 = 1") }
				if m.Frames[0].Locals[4].Int() != 1 { t.Errorf("Expected res2 = 1") }
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
				if lVal != 5 { t.Errorf("Expected len(range(5)) = 5, got %d", lVal) }
				
				sVal := m.Frames[0].Locals[2].Int()
				if sVal != 10 { t.Errorf("Expected sum(range(5)) = 10, got %d", sVal) }
				
				m1Val := m.Frames[0].Locals[3].Int()
				if m1Val != 4 { t.Errorf("Expected max(range(5)) = 4, got %d", m1Val) }
				
				m2Val := m.Frames[0].Locals[4].Int()
				if m2Val != 0 { t.Errorf("Expected min(range(5)) = 0, got %d", m2Val) }
				
				slVal := m.Frames[0].Locals[5].Int()
				if slVal != 5 { t.Errorf("Expected len('hello') = 5, got %d", slVal) }
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
				if m.Frames[0].Locals[1].Int() != 10 { t.Errorf("abs(-10) failed") }
				if m.Frames[0].Locals[2].Data != 1 { t.Errorf("bool(1) failed") }
				if m.Frames[0].Locals[3].Data != 0 { t.Errorf("bool(0) failed") }
				
				s1Val := m.Frames[0].Locals[4]
				if value.UnpackString(s1Val.Data, m.Arena) != "123" { t.Errorf("str(123) failed") }
				
				s2Val := m.Frames[0].Locals[5]
				if value.UnpackString(s2Val.Data, m.Arena) != "true" { t.Errorf("str(true) failed") }
				
				if m.Frames[0].Locals[6].Int() != 456 { t.Errorf("int('456') failed") }
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
				if lVal != 3 { t.Errorf("Expected 3 evens, got %d", lVal) }
				
				pVal := m.Frames[0].Locals[3].Int()
				if pVal != 8 { t.Errorf("Expected pow(2, 3) = 8, got %d", pVal) }
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
				if m.Frames[0].Locals[2].Data != 0 { t.Errorf("all(range(3)) should be false") }
				if m.Frames[0].Locals[3].Data != 1 { t.Errorf("any(range(3)) should be true") }
				if m.Frames[0].Locals[4].Data != 1 { t.Errorf("all([1,2,3]) should be true") }
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
				if xVal != 20 { t.Errorf("Expected items[1] = 20, got %d", xVal) }
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
				if m.Frames[0].Locals[1].Int() != 3 { t.Errorf("divmod q failed") }
				if m.Frames[0].Locals[2].Int() != 1 { t.Errorf("divmod r failed") }
				if value.UnpackString(m.Frames[0].Locals[3].Data, m.Arena) != "0xff" { t.Errorf("hex failed") }
				if value.UnpackString(m.Frames[0].Locals[7].Data, m.Arena) != "int" { t.Errorf("type failed") }
				if m.Frames[0].Locals[8].Data != 1 { t.Errorf("callable failed") }
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
				if val != 99 { t.Errorf("Expected 99, got %d", val) }
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
				if m.Frames[0].Locals[2].Int() != 10 { t.Errorf("next 1 failed") }
				if m.Frames[0].Locals[3].Int() != 20 { t.Errorf("next 2 failed") }
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
			
			machine.RegisterHostFunction("", func(m *vm.Machine) error { return nil }) // 0
			machine.RegisterHostFunction("", func(m *vm.Machine) error { return nil }) // 1
			machine.RegisterHostFunction("", func(m *vm.Machine) error { return nil }) // 2
			machine.RegisterHostFunction("", func(m *vm.Machine) error { return nil }) // 3
			machine.RegisterHostFunction("", func(m *vm.Machine) error { return nil }) // 4
			machine.RegisterHostFunction("", func(m *vm.Machine) error { return nil }) // 5
			machine.RegisterHostFunction("", func(m *vm.Machine) error { return nil }) // 6
			machine.RegisterHostFunction("", func(m *vm.Machine) error { return nil }) // 7
			machine.RegisterHostFunction("", func(m *vm.Machine) error { return nil }) // 8
			machine.RegisterHostFunction("", func(m *vm.Machine) error { return nil }) // 9
			machine.RegisterHostFunction("", func(m *vm.Machine) error { return nil }) // 10
			machine.RegisterHostFunction("", func(m *vm.Machine) error { return nil }) // 11
			machine.RegisterHostFunction("", func(m *vm.Machine) error { return nil }) // 12
			machine.RegisterHostFunction("", func(m *vm.Machine) error { return nil }) // 13
			machine.RegisterHostFunction("", stdlib.Len)      // 14
			machine.RegisterHostFunction("", stdlib.Range)    // 15
			machine.RegisterHostFunction("", stdlib.List)     // 16
			machine.RegisterHostFunction("", stdlib.Sum)      // 17
			machine.RegisterHostFunction("", stdlib.Max)      // 18
			machine.RegisterHostFunction("", stdlib.Min)      // 19
			machine.RegisterHostFunction("", stdlib.Map)      // 20
			machine.RegisterHostFunction("", stdlib.Abs)      // 21
			machine.RegisterHostFunction("", stdlib.Bool)     // 22
			machine.RegisterHostFunction("", stdlib.Int)      // 23
			machine.RegisterHostFunction("", stdlib.Str)      // 24
			machine.RegisterHostFunction("", stdlib.Filter)   // 25
			machine.RegisterHostFunction("", stdlib.Pow)      // 26
			machine.RegisterHostFunction("", stdlib.All)      // 27
			machine.RegisterHostFunction("", stdlib.Any)      // 28
						machine.RegisterHostFunction("", stdlib.MakeList)  // 29
						machine.RegisterHostFunction("", stdlib.GetItem)   // 30
						machine.RegisterHostFunction("", stdlib.SetItem)   // 59
						machine.RegisterHostFunction("", stdlib.DivMod)    // 31
			machine.RegisterHostFunction("", stdlib.Round)    // 32
			machine.RegisterHostFunction("", stdlib.Float)    // 33
			machine.RegisterHostFunction("", stdlib.Bin)      // 34
			machine.RegisterHostFunction("", stdlib.Oct)      // 35
			machine.RegisterHostFunction("", stdlib.Hex)      // 36
			machine.RegisterHostFunction("", stdlib.Chr)      // 37
			machine.RegisterHostFunction("", stdlib.Ord)      // 38
			machine.RegisterHostFunction("", stdlib.Dict)     // 39
			machine.RegisterHostFunction("", stdlib.Tuple)    // 40
			machine.RegisterHostFunction("", stdlib.Set)      // 41
			machine.RegisterHostFunction("", stdlib.Reversed)  // 42
			machine.RegisterHostFunction("", stdlib.Sorted)    // 43
			machine.RegisterHostFunction("", stdlib.Zip)       // 44
			machine.RegisterHostFunction("", stdlib.Enumerate) // 45
			machine.RegisterHostFunction("", stdlib.Repr)      // 46
			machine.RegisterHostFunction("", stdlib.Ascii)     // 47
			machine.RegisterHostFunction("", stdlib.Hash)      // 48
			machine.RegisterHostFunction("", stdlib.Id)        // 49
			machine.RegisterHostFunction("", stdlib.TypeWord)  // 50
			machine.RegisterHostFunction("", stdlib.Callable)  // 51
			machine.RegisterHostFunction("", stdlib.Iter)      // 52
			machine.RegisterHostFunction("", stdlib.Next)      // 53

			err = machine.Run(10000)
			if err != nil {
				t.Fatalf("Runtime error: %v", err)
			}
			
			tt.verify(machine, t)
		})
	}
}
