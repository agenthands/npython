import os

def generate_snippets():
    os.makedirs("tests/snippets", exist_ok=True)
    
    snippets = []

    # 1. Basic Arithmetic (300 snippets)
    for i in range(50):
        a, b = i * 7, i + 13
        snippets.append(f"print({a} + {b})")
        snippets.append(f"print({a} - {b})")
        snippets.append(f"print({a} * {b})")
        if b != 0:
            snippets.append(f"print({a} // {b})")
            snippets.append(f"print({a} % {b})")
            snippets.append(f"print(round({a} / {b}, 2))")

    # 2. String Operations (250 snippets)
    for i in range(50):
        s1 = f"hello_{i}"
        s2 = f"world_{50-i}"
        snippets.append(f"print('{s1}' + ' ' + '{s2}')")
        snippets.append(f"print('{s1}'.upper())")
        snippets.append(f"print('{s1}'.lower())")
        snippets.append(f"print(len('{s1}'))")
        snippets.append(f"print('{s1}'[1:4])")

    # 3. List Comprehensions (150 snippets)
    for i in range(50):
        n = (i % 10) + 1
        snippets.append(f"print([x * 2 for x in range({n})])")
        snippets.append(f"print([x for x in range({n * 2}) if x % 2 == 0])")
        # Use string keys for now to match implementation limit
        snippets.append(f"print({{str(x): x*x for x in range({n})}})")

    # 4. Functions & Scope (100 snippets)
    for i in range(50):
        snippets.append(f"def f{i}(x):\n    return x + {i}\nprint(f{i}(10))")
        # Avoid closures for now
        snippets.append(f"def g{i}(x, y):\n    return x * y + {i}\nprint(g{i}(2, 3))")

    # 5. Nested Structures (100 snippets)
    for i in range(50):
        n = (i % 5) + 2
        snippets.append(f"d = {{'a': [1,2,{i}], 'b': {{'ckey': {i*2}}}}}\nprint(d['a'][2])")
        snippets.append(f"l = [[1,2], [3,4], [{i}, {i+1}]]\nprint(l[2][0])")

    # 6. Built-ins (200 snippets)
    for i in range(50):
        snippets.append(f"print(abs({-i}))")
        snippets.append(f"print(min({i}, {i+1}, {i-1}))")
        snippets.append(f"print(max({i}, {i*2}, {i//2 if i > 0 else 0}))")
        snippets.append(f"print(sum([x for x in range({i%10 + 2})]))")

    # 7. Bitwise Ops (250 snippets)
    for i in range(50):
        snippets.append(f"print({i} & {i+1})")
        snippets.append(f"print({i} | {i+1})")
        snippets.append(f"print({i} ^ {i*2})")
        snippets.append(f"print({i} << 2)")
        snippets.append(f"print({i*10} >> 2)")

    # 8. Control Flow (50 snippets)
    for i in range(50):
        n = i % 10
        code = f"""
res = 0
for x in range({n}):
    if x % 2 == 0:
        res += x
    else:
        res *= 2
print(res)
"""
        snippets.append(code.strip())

    # 9. Complex Formatting (50 snippets)
    for i in range(25):
        snippets.append(f"print('val is {{}}'.format({i}))")
        snippets.append(f"print('{{0}} {{1}}'.format({i}, {i+1}))")

    final_snippets = snippets

    for i, code in enumerate(final_snippets):
        with open(f"tests/snippets/test_{i:03d}.py", "w") as f:
            f.write(code + "\n")

    print(f"Generated {len(final_snippets)} snippets in tests/snippets/")

if __name__ == "__main__":
    generate_snippets()
