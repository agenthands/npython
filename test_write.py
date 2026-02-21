def test_fs(fs_token):
    with scope("FS-ENV", fs_token):
        write_file("42", "output1.txt")

test_fs('mock_fs')
