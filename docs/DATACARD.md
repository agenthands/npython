# nPython Language Datacard (v1.1)
Output ONLY valid nPython (Python subset) code.

### **Language Constraints (CRITICAL)**
- **NO f-strings**: `f"hello {x}"` is NOT supported. Use `"hello {}".format(x)` or string concatenation.
- **No Imports**: `import` is disabled. All tools are built-in.
- **No IO without Scope**: You MUST use `with scope(NAME, token):` to access network/files.

### **Supported Python Subset**
- **Syntax**: `def`, `return`, `if/elif/else`, `while`, `for/in`, `break`, `continue`.
- **Operators**: `+`, `-`, `*`, `/`, `%`, `==`, `!=`, `>`, `<`, `>=`, `<=`, `and`, `or`.
- **Literals**: `10`, `3.14`, `"string"`, `[1, 2]`, `{"k": "v"}`, `True`, `False`, `None`.

### **Built-in Functions & Methods**
| Category | Functions / Methods |
| :--- | :--- |
| **Logic** | `bool`, `all`, `any`, `callable`, `type`, `is_empty` |
| **Math** | `int`, `float`, `abs`, `round`, `sum`, `max`, `min`, `pow`, `divmod` |
| **String** | `str`, `len`, `chr`, `ord`, `hex`, `bin`, `oct`, `.upper()`, `.lower()`, **`.format()`**, **`.json()`** |
| **Collections** | `list`, `dict`, `set`, `tuple`, `range`, `reversed`, `sorted`, `map`, `filter`, `zip`, **`.items()`**, **`.keys()`**, **`.append()`** |
| **IO** | `print`, `fetch`, `write_file`, `read_file` |


### **Code Examples**

**1. Secure Data Processing**
```python
def process_logs(url, token):
    with scope("HTTP-ENV", token):
        data = fetch(url)
    
    logs = parse_json(data)
    errors = filter(lambda x: x["level"] == "error", logs)
    
    if len(errors) > 0:
        print("Found errors: " + str(len(errors)))
```

**2. Algorithmic Logic**
```python
def fib(n):
    if n <= 1: return n
    return fib(n-1) + fib(n-2)

print(fib(10))
```
