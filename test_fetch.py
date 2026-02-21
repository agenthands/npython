def my_func(http_token, fs_token):
    with scope("HTTP-ENV", http_token):
        response = fetch("http://127.0.0.1:61690") # Or just ANY fetch for checking
        print(response)

my_func('mock_http', 'mock_fs')
