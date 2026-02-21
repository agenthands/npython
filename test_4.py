def my_func(http_token, fs_token):
    # Dummy variable
    response = ""
    # Mocking what parse_json would return for the json:
    data_dict = parse_json('{"status": "ok", "data": 42, "items": [{"id": 1}, {"id": 2}]}')
    print("Parsed JSON ok")
    status = data_dict["status"]
    print("Status is:")
    print(status)
    print(type(status))
    if status == "ok":
        print("Status is ok matched")
        my_list = []
        my_list.append(data_dict["data"])
        print("Appended ok")
        with scope("FS-ENV", fs_token):
            write_file(str(my_list), "output4_test.txt")

my_func('mock', 'mock')
