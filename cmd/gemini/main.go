package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/agenthands/nforth/pkg/compiler/emitter"
	"github.com/agenthands/nforth/pkg/compiler/lexer"
	"github.com/agenthands/nforth/pkg/compiler/parser"
	"github.com/agenthands/nforth/pkg/vm"
)

func main() {
	fileFlag := flag.String("file", "", "nFORTH source file to execute")
	gasFlag := flag.Int("gas", 1000000, "Maximum instruction limit")
	flag.Parse()

	if *fileFlag == "" {
		fmt.Println("Usage: gemini -file <source.nf>")
		os.Exit(1)
	}

	// 1. Load Source
	src, err := os.ReadFile(*fileFlag)
	if err != nil {
		fmt.Printf("Error reading file: %v
", err)
		os.Exit(1)
	}

	// 2. Compile
	s := lexer.NewScanner(src)
	p := parser.NewParser(s, src)
	prog, err := p.Parse()
	if err != nil {
		fmt.Printf("Compilation Error: %v
", err)
		os.Exit(1)
	}

	e := emitter.NewEmitter(src)
	bc, err := e.Emit(prog)
	if err != nil {
		fmt.Printf("Emitter Error: %v
", err)
		os.Exit(1)
	}

	// 3. Setup VM
	m := &vm.Machine{}
	m.Code = bc.Instructions
	m.Constants = bc.Constants

	// TODO: Register Stdlib Functions
	// TODO: Setup Security Gatekeeper

	// 4. Run
	fmt.Printf("Executing %s...
", *fileFlag)
	err = m.Run(*gasFlag)
	if err != nil {
		fmt.Printf("Runtime Error: %v
", err)
		os.Exit(1)
	}

	fmt.Println("Execution completed successfully.")
}
