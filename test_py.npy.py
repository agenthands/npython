
def my_func(http_token, fs_token):
    with scope("HTTP-ENV", http_token):
        response = fetch("http://127.0.0.1:61858")
    
    parsed = parse_json(response)
    data = parsed["data"]
    
    with scope("FS-ENV", fs_token):
        write_file(str(data), "output_test_run.txt")

my_func('mock_http', 'mock_fs')
