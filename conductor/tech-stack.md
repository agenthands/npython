# Tech Stack: nPython

- **Language Interface**: Python (Subset)
- **Compiler**: `github.com/go-python/gpython/ast` based translator.
- **Execution Engine**: nPython Bytecode VM (Go).
- **Data Representation**: 16-byte Tagged Union (`value.Value`) to prevent interface-boxing and heap escape.
- **Memory Management**: 
    - Fixed-size arrays for Stack and Call Frames.
    - Zero-allocation hot path.
- **Security**: 
    - Capability-based access control via `with scope()` blocks.
    - Sandboxed Filesystem and HTTP clients.
