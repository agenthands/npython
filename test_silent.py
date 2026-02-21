def my_func(http_token, fs_token):
    print("starting")
    with scope("HTTP-ENV", http_token):
        response = fetch("http://127.0.0.1:61690") # Or just a Google fetch
        print(response)

my_func('mock', 'mock')
