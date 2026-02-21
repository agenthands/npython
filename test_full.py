def my_func(http_token, fs_token):
    with scope("HTTP-ENV", http_token):
        response = fetch("http://127.0.0.1:8080")

    # Let's just do parse_json manually exactly like the LLM
    parsed = parse_json('{"data": 42}')
    data = parsed["data"]
    
    with scope("FS-ENV", fs_token):
        write_file(str(data), "output1.txt")

my_func('mock', 'mock')
