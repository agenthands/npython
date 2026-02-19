# nPython Language Datacard (v1.0)
Output ONLY valid nPython (Python subset) code.

### **Syntax Overview**
- **Variable Assignment**: `x = 10`
- **Control Flow**: `if/else`, `while`
- **Function Definition**: `def calculate(a, b): return a + b`
- **Security Gating**: `with scope("HTTP-ENV", token): fetch(url)`

### **Built-in Functions**
- `print(val)` (Print to stdout)
- `fetch(url)` (HTTP GET - Requires `HTTP-ENV`)
- `write_file(content, path)` (FS Write - Requires `FS-ENV`)
- `read_file(path)` (FS Read - Requires `FS-ENV`)
- `parse_json(string)` (Parse JSON into a dictionary)
- `format_string(format, val)` (Simple string formatting)
- `is_empty(val)` (Check if a value is empty)

### **The Banned List**
- **NO `import`**: All tools are built-in or provided via security scopes.
- **NO `try/except`**: All errors are fatal and handled by the VM's security layer.
- **NO Ambient Authority**: Never call `fetch` or `write_file` without a `with scope` block.

### **Code Examples**

**1. Secure Data Fetching:**
```python
def save_data(url, path, token):
    with scope("HTTP-ENV", token):
        raw = fetch(url)
    
    with scope("FS-ENV", token):
        write_file(raw, path)
```

**2. Conditional Logic:**
```python
def check_value(val):
    if val > 100:
        print("High")
    else:
        print("Low")
```

### **Token Efficiency vs Standard Python**
| Metric | nPython | Standard Python |
| :--- | :--- | :--- |
| **Logic** | Restricted Subset | Global Language |
| **Structure** | Sterile & Sandboxed | Extensive OS Access |
| **Boilerplate** | Zero (Sandboxed) | Extensive (Imports/Try-Except) |
