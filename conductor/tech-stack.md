# Tech Stack

- **Language**: Go (Golang)
- **Data Representation**: 16-byte Tagged Union (`value.Value`) to prevent interface-boxing and heap escape.
- **Instruction Format**: `uint32` bytecode (8-bit Opcode, 24-bit Operand).
- **Memory Management**: 
    - Fixed-size arrays for Stack (`StackDepth=128`) and Call Frames (`MaxFrames=32`).
    - `sync.Pool` for Machine reuse.
    - Zero-copy scanning for the Lexer.
- **Security**: 
    - Capability-based access control.
    - HMAC-SHA256 for token verification.
- **Testing**: 
    - 7-step TDD cycle.
    - Benchmarking for 0-allocation verification and throughput (>1M ops/sec).
