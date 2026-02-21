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

func genCode(ctx context.Context, model *genai.GenerativeModel, prompt string) string {
	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		fmt.Printf("GenAI Error: %v\n", err)
		return ""
	}
	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return ""
	}
	txt := fmt.Sprintf("%v", resp.Candidates[0].Content.Parts[0])

	// Extract code block
	if strings.Contains(txt, "```python") {
		parts := strings.Split(txt, "```python")
		if len(parts) > 1 {
			return strings.Split(parts[1], "```")[0]
		}
	} else if strings.Contains(txt, "```") {
		parts := strings.Split(txt, "```")
		if len(parts) > 1 {
			return strings.Split(parts[1], "```")[0]
		}
	}
	return txt
}

type execResult struct {
	Success  bool
	ErrorMsg string
}

func execCode(binPath, code, ext string) execResult {
	tmpFile := "eval_tmp." + ext + ".py"
	os.WriteFile(tmpFile, []byte(code), 0644)
	defer os.Remove(tmpFile)

	var cmd *exec.Cmd
	if ext == "npy" {
		cmd = exec.Command(binPath, "run", tmpFile)
	} else {
		cmd = exec.Command(binPath, tmpFile)
	}

	out, err := cmd.CombinedOutput()
	if err != nil {
		return execResult{Success: false, ErrorMsg: string(out) + "\n" + err.Error()}
	}

	return execResult{Success: true, ErrorMsg: string(out)}
}

func getOutput(file string) string {
	b, err := os.ReadFile(file)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(b))
}

func main() {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		fmt.Println("Missing GEMINI_API_KEY")
		return
	}

	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		panic(err)
	}
	defer client.Close()

	model := client.GenerativeModel("gemini-2.0-flash")
	temp := float32(0.2) // Low temperature for coding
	model.Temperature = &temp

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/index" {
			fmt.Fprint(w, `{"datasets": [{"id": "ds_1", "name": "inventory"}, {"id": "ds_2", "name": "employees"}]}`)
		} else if strings.HasPrefix(r.URL.Path, "/data/ds_2") {
			fmt.Fprint(w, `{"employees": [{"id": 1, "salary": 45000}, {"id": 2, "salary": 60000}, {"id": 3, "salary": 70000}]}`)
		} else {
			w.WriteHeader(404)
		}
	}))
	defer ts.Close()

	promptData, _ := os.ReadFile("docs/PROMPT.md")
	datacardData, _ := os.ReadFile("docs/DATACARD.md")
	npythonPrompt := string(promptData) + "\n\n" + string(datacardData)

	stdPyPrompt := `Output ONLY valid Python 3 code. 
CRITICAL: DO NOT use third-party libraries like requests. You MUST strictly use urllib.request.`

	fmt.Println("RUNNING AGENTIC MULTI-TURN BENCHMARKS")
	fmt.Println("==================================================================")

	iterations := 5
	pySucc := 0
	npySucc := 0

	for i := 1; i <= iterations; i++ {
		fmt.Printf("\n--- ITERATION %d ---\n", i)
		os.Remove("step1.txt")
		os.Remove("step2.txt")

		// STANDARD PYTHON
		pyTurn1 := fmt.Sprintf("TASK 1: Call '%s/index'. Parse the JSON. Find the 'id' of the dataset named 'employees'. Write ONLY this string ID to 'step1.txt'.", ts.URL)

		pyCode1 := genCode(ctx, model, stdPyPrompt+"\nTASK:\n"+pyTurn1)
		execCode("python3", pyCode1, "py")
		out1Py := getOutput("step1.txt")

		var pyRes bool
		if out1Py == "ds_2" {
			pyTurn2 := fmt.Sprintf("PREVIOUS STEP OUTPUT: You found dataset id '%s'.\nTASK 2: Fetch the data from '%s/data/%s'. The JSON has an 'employees' array. Filter the array to ONLY include employees with a 'salary' strictly greater than 50000. Sum those salaries. Write the integer sum to 'step2.txt'.", out1Py, ts.URL, out1Py)
			pyCode2 := genCode(ctx, model, stdPyPrompt+"\nTASK:\n"+pyTurn1+"\n\nMY PREVIOUS CODE:\n```python\n"+pyCode1+"\n```\n\n"+pyTurn2)
			execCode("python3", pyCode2, "py")
			if getOutput("step2.txt") == "130000" {
				pyRes = true
				pySucc++
			}
		} else {
			fmt.Printf("Py Failed Turn 1! CODE:\n%s\n", pyCode1)
		}

		// NPYTHON
		os.Remove("step1.txt")
		os.Remove("step2.txt")

		npyBase := "Write a top-level function STRICTLY named `my_func(http_token, fs_token)` that executes the logic. Then call the function exactly like this at the end:\n\nmy_func(http_token='mock_http', fs_token='mock_fs')\n\nTASK: "

		npyTurn1 := fmt.Sprintf("TASK 1: Fetch '%s/index'. Parse the JSON. Find the 'id' of the dataset named 'employees'. Write ONLY this string ID to 'step1.txt'.", ts.URL)
		npyCode1 := genCode(ctx, model, npythonPrompt+"\n"+npyBase+npyTurn1)
		execCode("./npython", npyCode1, "npy")
		out1Npy := getOutput("step1.txt")

		var npyRes bool
		var npyErrStr string
		var npyCode2 string
		if out1Npy == "ds_2" {
			npyTurn2 := fmt.Sprintf("PREVIOUS STEP OUTPUT: You found dataset id '%s'.\nTASK 2: Fetch the data from '%s/data/%s'. The JSON has an 'employees' array. Filter the array to ONLY include employees with a 'salary' strictly greater than 50000. Sum those salaries. Write the integer sum to 'step2.txt'.", out1Npy, ts.URL, out1Npy)
			npyCode2 = genCode(ctx, model, npythonPrompt+"\n"+npyBase+npyTurn1+"\n\nMY PREVIOUS CODE:\n```python\n"+npyCode1+"\n```\n\n"+npyTurn2)
			res := execCode("./npython", npyCode2, "npy")
			if getOutput("step2.txt") == "130000" {
				npyRes = true
				npySucc++
			} else {
				if !res.Success {
					npyErrStr = res.ErrorMsg
				}
			}
		} else {
			// Let me just grab the execution error if it failed
			res := execCode("./npython", npyCode1, "npy")
			npyErrStr = "Failed Turn 1: wrote '" + out1Npy + "'\nSTDOUT/ERR:\n" + res.ErrorMsg
			if !res.Success {
				npyErrStr += "\nProcess Failed!"
			}
		}

		fmt.Printf("Iter %d | Std Py Multi-Turn Success: %t | nPython Multi-Turn Success: %t\n", i, pyRes, npyRes)
		if !npyRes && npyErrStr != "" {
			fmt.Printf("  nPy Err:\n%s\n", npyErrStr)
			fmt.Printf("  CODE 1:\n%s\n", npyCode1)
			if npyCode2 != "" {
				fmt.Printf("  CODE 2:\n%s\n", npyCode2)
			}
		}
	}

	fmt.Printf("\n--- OVERALL RESULTS ---\n")
	fmt.Printf("Standard Python Agentic Success: %d / %d\n", pySucc, iterations)
	fmt.Printf("nPython Agentic Success: %d / %d\n", npySucc, iterations)
}
