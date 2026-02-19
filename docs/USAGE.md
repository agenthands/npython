# nPython Usage Guide

nPython is a secure, high-performance execution environment for AI Agents, allowing them to run a subset of Python on a strictly sandboxed, zero-allocation Virtual Machine.

## 🧠 AI-Native Philosophy

**The Problem:** Traditional Python runtimes grant "ambient authority" to scripts, meaning they can access the filesystem, network, and process table by default. For AI Agents, this is a massive security risk (The Confused Deputy Attack).

**The nPython Solution:**
We enforce a strict **Capability-Based Security** model. Scripts have ZERO authority by default. They must explicitly request access to tools using security scopes.

## 🛡️ Secure Execution

Under the hood, nPython uses a Go-based Virtual Machine. Python source code is compiled into secure bytecode and executed in a sterile environment.

If a script attempts to:
1. Access the network without an active `HTTP-ENV` scope.
2. Write to a file without an active `FS-ENV` scope.
3. Access files outside the workspace root (Path Traversal).

The VM immediately halts execution and returns a security violation error.

## Core Concepts

### 1. Variables and Assignment
Standard Python variable assignment is supported.
```python
x = 10
y = 20
total = x + y
```

### 2. Control Flow
Standard `if/else` and `while` loops are supported.

```python
if score > 90:
    grade = "A"
else:
    grade = "B"

i = 0
while i < 10:
    print(i)
    i = i + 1
```

### 3. Function Definitions
Functions are defined using `def`.
```python
def calculate_tax(price, rate):
    return price * rate / 100
```

### 4. Security Scopes (`with scope`)
Privileged operations are only accessible within a `with scope(name, token):` block.

```python
with scope("HTTP-ENV", http_token):
    html = fetch("https://google.com")

with scope("FS-ENV", fs_token):
    write_file(html, "output.html")
```

## Built-in Functions

| Function | Description |
| :--- | :--- |
| `print(val)` | Print to stdout |
| `fetch(url)` | HTTP GET (Requires `HTTP-ENV`) |
| `write_file(content, path)` | FS Write (Requires `FS-ENV`) |
| `read_file(path)` | FS Read (Requires `FS-ENV`) |
| `parse_json(string)` | Parse JSON into a dictionary |
| `format_string(format, val)` | Simple string formatting (`%s` replacement) |
| `is_empty(val)` | Check if a value is empty (string, map, void) |
