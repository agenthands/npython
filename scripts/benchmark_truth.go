//go:build ignore

package main

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"strings"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

type Result struct {
	Env      string
	Task     string
	Success  bool
	ErrorMsg string
	Code     string
}

func main() {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		fmt.Println("GEMINI_API_KEY required")
		return
	}
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		panic(err)
	}
	defer client.Close()

	model := client.GenerativeModel("gemini-2.0-flash")
	temp := float32(0.3)
	model.Temperature = &temp

	// Start a mock server for our real tasks
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/users" {
			fmt.Fprint(w, `{"users": [{"id": 1, "role": "admin"}, {"id": 2, "role": "user"}, {"id": 3, "role": "admin"}, {"id": 4, "role": "guest"}]}`)
		} else {
			fmt.Fprint(w, `{"status": "ok", "data": 42, "items": [{"id": 1}, {"id": 2}]}`)
		}
	}))
	defer ts.Close()

	tasks := []string{
		fmt.Sprintf("Make an HTTP GET request to '%s'. Parse the JSON response. Extract the 'data' field. Write the integer value to 'output1.txt'.", ts.URL),
		"Calculate the sum of the squares of the first 10 even numbers (0, 2, 4...) and write it to 'output2.txt'.",
		fmt.Sprintf("Fetch JSON from '%s'. It has an 'items' array. Create a new list containing just the 'id' of each item. Write the list length to 'output3.txt'.", ts.URL),
		fmt.Sprintf("Send an HTTP request to '%s'. If the response JSON 'status' is 'ok', append the 'data' value to a new list and write the list to 'output4.txt'.", ts.URL),
		fmt.Sprintf("Fetch JSON from '%s/users'. Filter the 'users' array to only include items where 'role' is 'admin'. Extract the 'id' of these admins and calculate their sum. Write the integer sum to 'output5.txt'.", ts.URL),
		fmt.Sprintf("Fetch JSON from '%s'. Extract the integer 'data' field. Create a list of numbers from 1 up to 'data' (inclusive), but skip ANY number that is divisible by 5. Write the length of this final skipped list to 'output6.txt'.", ts.URL),
	}

	promptData, _ := os.ReadFile("docs/PROMPT.md")
	datacardData, _ := os.ReadFile("docs/DATACARD.md")
	npythonPrompt := string(promptData) + "\n\n" + string(datacardData)

	stdPyPrompt := "You are a standard Python 3 expert. Write Python code to solve the task. " +
		"CRITICAL: You MUST use ONLY the Python Standard Library. DO NOT use third-party libraries like `requests`. For HTTP, use `urllib.request`. " +
		"Write ONLY code inside a ```python block."

	nPypromptBase := "You are an nPython expert. Read the system prompt carefully. Write ONLY code inside a ```python block."

	fmt.Println("RUNNING TRUE EXECUTION BENCHMARKS")
	fmt.Println("==================================================================")

	iterations := 5 // Run multiple times to observe variance/hallucination rate

	pyWins := 0
	npyWins := 0

	var npyErrors []string

	for tIdx, task := range tasks {
		fmt.Printf("\n--- TASK %d: %s\n", tIdx+1, task)
		for i := 0; i < iterations; i++ {
			// Clean output files
			os.Remove("output1.txt")
			os.Remove("output2.txt")
			os.Remove("output3.txt")
			os.Remove("output4.txt")
			os.Remove("output5.txt")
			os.Remove("output6.txt")

			// Eval Std Python
			pyCode := genCode(ctx, model, stdPyPrompt+"\nTASK:\n"+task)
			pyRes := execCode("python3", pyCode, "std")
			pySuccess := checkOutputs(tIdx + 1)

			// Clean output files for nPython
			os.Remove("output1.txt")
			os.Remove("output2.txt")
			os.Remove("output3.txt")
			os.Remove("output4.txt")
			os.Remove("output5.txt")
			os.Remove("output6.txt")

			// Eval nPython
			wrappedTask := "Write a top-level function STRICTLY named `my_func(http_token, fs_token)` that executes the logic. Then call the function exactly like this at the end:\n\nmy_func(http_token='mock_http', fs_token='mock_fs')\n\nTASK: " + task
			npyCode := genCode(ctx, model, npythonPrompt+"\n"+nPypromptBase+"\n"+wrappedTask)
			npyRes := execCode("./npython", npyCode, "npy")

			npySuccess := false
			if npyRes.ErrorMsg == "" {
				npySuccess = checkOutputs(tIdx + 1)
			} else {
				npyErrors = append(npyErrors, npyRes.ErrorMsg)
			}

			if pySuccess {
				pyWins++
			}
			if npySuccess {
				npyWins++
			}

			fmt.Printf("Iter %d | Std Py: %-5v | nPython: %-5v\n", i+1, pySuccess, npySuccess)
			if !npySuccess {
				if npyRes.ErrorMsg != "" {
					fmt.Printf("  nPy Err: %s\n", strings.Split(npyRes.ErrorMsg, "\n")[0])
					fmt.Printf("CODE:\n%s\n", npyCode)
				} else {
					fmt.Printf("  nPy Silently Failed! Output check failed:\n%s\n", npyCode)
				}
			}
			if !pySuccess && pyRes.ErrorMsg != "" {
				firstLine := strings.Split(pyRes.ErrorMsg, "\n")
				if len(firstLine) > 0 {
					fmt.Printf("  StdPy Err: %s\n", firstLine[len(firstLine)-2])
				}
			}
		}
	}

	fmt.Printf("\n--- OVERALL RESULTS ---\n")
	fmt.Printf("Standard Python Success Rate: %d / %d\n", pyWins, len(tasks)*iterations)
	fmt.Printf("nPython Success Rate: %d / %d\n", npyWins, len(tasks)*iterations)
	fmt.Printf("\nCommon nPython Errors for Improvements:\n")
	errMap := make(map[string]int)
	for _, e := range npyErrors {
		// just take the general error
		lines := strings.Split(e, "\n")
		errStr := "Unknown"
		if len(lines) > 0 {
			errStr = lines[0]
		}
		errMap[errStr]++
	}
	for k, v := range errMap {
		fmt.Printf("- %s (x%d)\n", k, v)
	}

}

func genCode(ctx context.Context, model *genai.GenerativeModel, promptStr string) string {
	resp, err := model.GenerateContent(ctx, genai.Text(promptStr))
	if err != nil {
		return ""
	}
	return extractCodeBlock(resp)
}

func extractCodeBlock(resp *genai.GenerateContentResponse) string {
	if resp == nil || len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return ""
	}
	t := fmt.Sprintf("%v", resp.Candidates[0].Content.Parts[0])
	if strings.Contains(t, "```python") {
		return strings.Split(strings.Split(t, "```python")[1], "```")[0]
	}
	if strings.Contains(t, "```") {
		return strings.Split(strings.Split(t, "```")[1], "```")[0]
	}
	return t
}

func execCode(bin, code, ext string) Result {
	tmpName := "eval_tmp." + ext + ".py"
	os.WriteFile(tmpName, []byte(code), 0644)
	defer os.Remove(tmpName)

	var cmd *exec.Cmd
	if bin == "python3" {
		cmd = exec.Command("python3", tmpName)
	} else {
		// nPython needs to have gatekeepers set or we construct a command
		cmd = exec.Command("./npython", "run", tmpName, "-gas", "5000000")
	}

	out, err := cmd.CombinedOutput()
	res := Result{Code: code, Success: err == nil}
	if err != nil {
		res.ErrorMsg = string(out)
	}
	return res
}

func checkOutputs(taskIdx int) bool {
	filename := fmt.Sprintf("output%d.txt", taskIdx)
	b, err := os.ReadFile(filename)
	if err != nil {
		return false
	}
	s := strings.TrimSpace(string(b))

	switch taskIdx {
	case 1:
		return s == "42"
	case 2:
		return s == "1140" // sum of squares of 0,2,4,6,8,10,12,14,16,18 => wait, first 10 even. 0*0 + 2*2 + 4*4 + ... + 18*18. 0+4+16+36+64+100+144+196+256+324 = 1140
	case 3:
		return s == "2"
	case 4:
		// [42] because list str representation in nPython is like [42]
		return strings.Contains(s, "42") && strings.Contains(s, "[")
	case 5:
		return s == "4" // 1 + 3 (admin role ids)
	case 6:
		return s == "34" // 42 items - (42 // 5 = 8 skipped) = 34
	}
	return false
}
