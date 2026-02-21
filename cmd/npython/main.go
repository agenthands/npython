package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/agenthands/npython/pkg/compiler/emitter"
	"github.com/agenthands/npython/pkg/compiler/lexer"
	"github.com/agenthands/npython/pkg/compiler/parser"
	"github.com/agenthands/npython/pkg/compiler/python"
	"github.com/agenthands/npython/pkg/stdlib"
	"github.com/agenthands/npython/pkg/vm"
)

type cliGatekeeper struct{}

func (g *cliGatekeeper) Validate(scope, token string) bool {
	return token != "" // Simple validation for benchmark
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: npython [run|query] ...")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "run":
		runScript()
	case "query":
		runQuery()
	default:
		fmt.Println("Unknown command:", os.Args[1])
		os.Exit(1)
	}
}

func runScript() {
	runCmd := flag.NewFlagSet("run", flag.ExitOnError)
	gasLimit := runCmd.Int("gas", 1000000, "Maximum instruction limit")

	if len(os.Args) < 3 {
		fmt.Println("Usage: npython run <source.py> [-gas limit]")
		os.Exit(1)
	}
	scriptPath := os.Args[2]
	runCmd.Parse(os.Args[3:])

	src, err := os.ReadFile(scriptPath)
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		os.Exit(1)
	}

	execute(string(src), filepath.Ext(scriptPath) == ".py", *gasLimit)
}

func runQuery() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: npython query <url> [token]")
		os.Exit(1)
	}
	url := os.Args[2]
	token := "token"
	if len(os.Args) > 3 {
		token = os.Args[3]
	}

	src := fmt.Sprintf(`
with scope("HTTP-ENV", "%s"):
    print(fetch("%s"))
`, token, url)

	execute(src, true, 1000000)
}

func execute(src string, isPython bool, gasLimit int) {
	var bc *vm.Bytecode
	var err error
	if isPython {
		c := python.NewCompiler()
		bc, err = c.Compile(src)
	} else {
		srcBytes := []byte(src)
		s := lexer.NewScanner(srcBytes)
		p := parser.NewParser(s, srcBytes)
		prog, err := p.Parse()
		if err != nil {
			fmt.Printf("Compilation Error: %v\n", err)
			os.Exit(1)
		}
		e := emitter.NewEmitter(srcBytes)
		bc, err = e.Emit(prog)
	}

	if err != nil {
		fmt.Printf("Compilation Error: %v\n", err)
		os.Exit(1)
	}

	m := vm.GetMachine()
	defer vm.PutMachine(m)

	m.Code = bc.Instructions
	m.Constants = bc.Constants
	m.Arena = bc.Arena
	m.Gatekeeper = &cliGatekeeper{}

	wd, _ := os.Getwd()
	fsSandbox := stdlib.NewFSSandbox(wd, 5*1024*1024)
	httpSandbox := stdlib.NewHTTPSandbox([]string{"localhost", "127.0.0.1", "api.github.com", "google.com"})
	httpSandbox.AllowLocalhost = true

	// Host Registry
	m.HostRegistry = make([]vm.HostFunctionEntry, 0, 64)
	m.RegisterHostFunction("FS-ENV", fsSandbox.WriteFile)       // 0
	m.RegisterHostFunction("HTTP-ENV", httpSandbox.Fetch)       // 1
	m.RegisterHostFunction("", stdlib.Print)                    // 2
	m.RegisterHostFunction("", stdlib.ParseJSON)                // 3
	m.RegisterHostFunction("", stdlib.GetField)                 // 4
	m.RegisterHostFunction("HTTP-ENV", httpSandbox.SendRequest) // 5
	m.RegisterHostFunction("", httpSandbox.CheckStatus)         // 6
	m.RegisterHostFunction("", stdlib.ParseJSONKey)             // 7
	m.RegisterHostFunction("", stdlib.ParseJSONKey)             // 8
	m.RegisterHostFunction("", stdlib.FormatString)             // 9
	m.RegisterHostFunction("", stdlib.IsEmpty)                  // 10
	m.RegisterHostFunction("", httpSandbox.WithClient)          // 11
	m.RegisterHostFunction("", httpSandbox.SetURL)              // 12
	m.RegisterHostFunction("", httpSandbox.SetMethod)           // 13
	m.RegisterHostFunction("", stdlib.Len)                      // 14
	m.RegisterHostFunction("", stdlib.Range)                    // 15
	m.RegisterHostFunction("", stdlib.List)                     // 16
	m.RegisterHostFunction("", stdlib.Sum)                      // 17
	m.RegisterHostFunction("", stdlib.Max)                      // 18
	m.RegisterHostFunction("", stdlib.Min)                      // 19
	m.RegisterHostFunction("", stdlib.Map)                      // 20
	m.RegisterHostFunction("", stdlib.Abs)                      // 21
	m.RegisterHostFunction("", stdlib.Bool)                     // 22
	m.RegisterHostFunction("", stdlib.Int)                      // 23
	m.RegisterHostFunction("", stdlib.Str)                      // 24
	m.RegisterHostFunction("", stdlib.Filter)                   // 25
	m.RegisterHostFunction("", stdlib.Pow)                      // 26
	m.RegisterHostFunction("", stdlib.All)                      // 27
	m.RegisterHostFunction("", stdlib.Any)                      // 28
	m.RegisterHostFunction("", stdlib.MakeList)                 // 29
	m.RegisterHostFunction("", stdlib.GetItem)                  // 30
	m.RegisterHostFunction("", stdlib.SetItem)                  // 31
	m.RegisterHostFunction("", stdlib.DivMod)                   // 32
	m.RegisterHostFunction("", stdlib.Round)                    // 33
	m.RegisterHostFunction("", stdlib.Float)                    // 34
	m.RegisterHostFunction("", stdlib.Bin)                      // 35
	m.RegisterHostFunction("", stdlib.Oct)                      // 36
	m.RegisterHostFunction("", stdlib.Hex)                      // 37
	m.RegisterHostFunction("", stdlib.Chr)                      // 38
	m.RegisterHostFunction("", stdlib.Ord)                      // 39
	m.RegisterHostFunction("", stdlib.Dict)                     // 40
	m.RegisterHostFunction("", stdlib.Tuple)                    // 41
	m.RegisterHostFunction("", stdlib.Set)                      // 42
	m.RegisterHostFunction("", stdlib.Reversed)                 // 43
	m.RegisterHostFunction("", stdlib.Sorted)                   // 44
	m.RegisterHostFunction("", stdlib.Zip)                      // 45
	m.RegisterHostFunction("", stdlib.Enumerate)                // 46
	m.RegisterHostFunction("", stdlib.Repr)                     // 47
	m.RegisterHostFunction("", stdlib.Ascii)                    // 48
	m.RegisterHostFunction("", stdlib.Hash)                     // 49
	m.RegisterHostFunction("", stdlib.Id)                       // 50
	m.RegisterHostFunction("", stdlib.TypeWord)                 // 51
	m.RegisterHostFunction("", stdlib.Callable)                 // 52
	m.RegisterHostFunction("", stdlib.Iter)                     // 53
	m.RegisterHostFunction("", stdlib.Next)                     // 54
	m.RegisterHostFunction("", stdlib.Locals)                   // 55
	m.RegisterHostFunction("", stdlib.Globals)                  // 56
	m.RegisterHostFunction("", stdlib.SliceBuiltin)             // 57
	m.RegisterHostFunction("", stdlib.Bytes)                    // 58
	m.RegisterHostFunction("", stdlib.ByteArray)                // 59
	m.RegisterHostFunction("", stdlib.HasNext)                  // 60
	m.RegisterHostFunction("", stdlib.MakeTuple)                // 61
	m.RegisterHostFunction("", stdlib.MethodCall)               // 62
	m.RegisterHostFunction("", stdlib.IsInstance)               // 63

	for name, ip := range bc.Functions {
		m.FunctionRegistry[name] = ip
	}

	err = m.Run(gasLimit)
	if err != nil {
		fmt.Printf("Runtime Error: %v\n", err)
		os.Exit(1)
	}
}
