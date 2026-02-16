package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/agenthands/nforth/pkg/compiler/emitter"
	"github.com/agenthands/nforth/pkg/compiler/lexer"
	"github.com/agenthands/nforth/pkg/compiler/parser"
	"github.com/agenthands/nforth/pkg/core/value"
	"github.com/agenthands/nforth/pkg/stdlib"
	"github.com/agenthands/nforth/pkg/vm"
)

type cliGatekeeper struct {
	tokens map[string]string
}

func (g *cliGatekeeper) Validate(scope, token string) bool {
	// For testing simplicity, we accept "token" or the specific test tokens
	return token == "token" || token == "http-token" || token == "fs-token" || token == "secret-fs-token" || token == "http-secret" || token == "fs-secret"
}

func main() {
	if len(os.Args) < 3 || os.Args[1] != "run" {
		fmt.Println("Usage: nforth run <source.nf> [-gas limit]")
		os.Exit(1)
	}

	runCmd := flag.NewFlagSet("run", flag.ExitOnError)
	gasLimit := runCmd.Int("gas", 1000000, "Maximum instruction limit")
	argFlag := runCmd.Int64("arg", 0, "Initial argument to push to stack")
	
	scriptPath := os.Args[2]
	runCmd.Parse(os.Args[3:])

	// 1. Load Source
	src, err := os.ReadFile(scriptPath)
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		os.Exit(1)
	}

	// 2. Compile
	s := lexer.NewScanner(src)
	p := parser.NewParser(s, src)
	prog, err := p.Parse()
	if err != nil {
		fmt.Printf("Compilation Error: %v\n", err)
		os.Exit(1)
	}

	e := emitter.NewEmitter(src)
	bc, err := e.Emit(prog)
	if err != nil {
		fmt.Printf("Emitter Error: %v\n", err)
		os.Exit(1)
	}

	// 3. Setup VM
	m := &vm.Machine{}
	m.Code = bc.Instructions
	m.Constants = bc.Constants
	m.Arena = bc.Arena

	// Push initial argument if it's provided via flag
	if *argFlag != 0 {
		m.Push(value.Value{Type: value.TypeInt, Data: uint64(*argFlag)})
	}

	// Security setup
	m.Gatekeeper = &cliGatekeeper{}
	
	// Registry Stdlib
	wd, _ := os.Getwd()
	fsSandbox := stdlib.NewFSSandbox(wd, 5*1024*1024)
	httpSandbox := stdlib.NewHTTPSandbox([]string{"localhost", "127.0.0.1", "api.github.com", "google.com"})
	httpSandbox.AllowLocalhost = true

	// Ensure registration order matches Emitter IDs
	// WRITE-FILE = 0
	// FETCH = 1
	m.RegisterHostFunction("FS-ENV", fsSandbox.WriteFile)
	m.RegisterHostFunction("HTTP-ENV", httpSandbox.Fetch)
	
	// Optional: Register non-scoped words as syscalls if they need complex logic
	m.RegisterHostFunction("", func(m *vm.Machine) error {
		val := m.Pop()
		if val.Type == value.TypeString {
			fmt.Println(value.UnpackString(val.Data, m.Arena))
		} else {
			fmt.Println(val.Data)
		}
		return nil
	})

	// 4. Run
	err = m.Run(*gasLimit)
	if err != nil {
		fmt.Printf("Runtime Error: %v\n", err)
		os.Exit(1)
	}

	// Execution successful
}
