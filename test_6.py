def my_func(http_token, fs_token):
    raw_data = '{"data": 42}'
    parsed_data = parse_json(raw_data)
    data_value = int(parsed_data["data"])
    
    numbers = []
    for i in range(1, data_value + 1):
        if i % 5 != 0:
            numbers.append(i)
    
    length_of_numbers = len(numbers)
    
    with scope("FS-ENV", fs_token):
        write_file(str(length_of_numbers), "output6.txt")

my_func('mock', 'mock')
