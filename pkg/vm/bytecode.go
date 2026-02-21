package vm

import "github.com/agenthands/npython/pkg/core/value"

// Bytecode represents the compiled output of a program.
type Bytecode struct {
	Instructions []uint32
	Constants    []value.Value
	Arena        []byte
	Functions    map[string]int
}
