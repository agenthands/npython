package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
)

func main() {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"status": "ok", "data": 42, "items": [{"id": 1}, {"id": 2}]}`)
	}))
	defer ts.Close()

	code := fmt.Sprintf(`
def my_func(http_token, fs_token):
    with scope("HTTP-ENV", http_token):
        response = fetch("%s")
    
    parsed = parse_json(response)
    data = parsed["data"]
    
    with scope("FS-ENV", fs_token):
        write_file(str(data), "output_test_run.txt")

my_func('mock_http', 'mock_fs')
`, ts.URL)

	os.WriteFile("test_py.npy.py", []byte(code), 0644)
	cmd := exec.Command("./npython", "run", "test_py.npy.py", "-gas", "5000000")
	out, err := cmd.CombinedOutput()
	fmt.Printf("CMD Output:\n%s\nErr: %v\n", string(out), err)
}
