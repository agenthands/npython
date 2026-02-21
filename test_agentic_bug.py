def my_func(http_token, fs_token):
    with scope("HTTP-ENV", http_token):
        raw_data = fetch("http://127.0.0.1:50327/index")
    print("Fetched RAW:")
    print(raw_data)
    data = parse_json(raw_data)
    
    print("Testing manual access:")
    datasets = data["datasets"]
    print("Type of dataset list:")
    print(type(datasets))
    
    print("Looping:")
    for ds in datasets:
        print("Item type:")
        print(type(ds))
        name = ds["name"]
        print("Name is:")
        print(name)
        if name == "employees":
            print("FOUND IT!")

my_func('mock', 'mock')
