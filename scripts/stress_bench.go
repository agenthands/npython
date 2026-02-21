//go:build ignore

package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

type Stats struct {
	Pass        int
	Fail        int
	TokensTotal int
}

func main() {
	apiKey := os.Getenv("GEMINI_API_KEY")
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		panic(err)
	}
	defer client.Close()

	model := client.GenerativeModel("gemini-2.0-flash")
	temp := float32(0.2)
	model.Temperature = &temp

	promptData, _ := os.ReadFile("docs/PROMPT.md")
	datacardData, _ := os.ReadFile("docs/DATACARD.md")
	npythonPrompt := string(promptData) + "\n\n" + string(datacardData)

	tasks := []string{
		"AGENT TASK: Fetch repo list from 'https://api.github.com/orgs/google/repos', parse JSON, filter only 'Go' repos, sum their 'stargazers_count', and write to 'stars_report.txt'. Use HTTP-ENV and FS-ENV.",
		"DATA PIPELINE: You have transactions: [{'id': 1, 'val': 100, 'type': 'credit'}, {'id': 2, 'val': 50, 'type': 'debit'}]. Filter 'credit', square the values, and return as a list. Print it.",
		"ALGORITHMIC: Calculate first 10 primes, join them with hyphens (e.g. '2-3-5'), and save to 'primes.log' within FS-ENV scope.",
		"RECONCILIATION: Compare local_inventory and master_inventory dictionaries. Find missing keys and value mismatches. Create a discrepancy report dict and print its length.",
		"NESTED CONFIG: Parse config JSON. If env is 'prod', set debug=False, timeout=60. If 'stage', debug=True, timeout=30. Print the final timeout.",
		"LOG PARSER: Process a list of log strings like '2026-02-19 ERROR 404', '2026-02-19 INFO 200'. Extract the code (e.g. 404) for all ERROR logs, find unique codes using a set, and print the count.",
		"MULTI-STEP IO: Read 'source.url' from FS-ENV, fetch the URL from HTTP-ENV, reverse the fetched string, and write it back to 'result.txt' in FS-ENV.",
		"AGGREGATOR: From list of [{'user': 'A', 'score': 10}, {'user': 'B', 'score': 20}, {'user': 'A', 'score': 5}], calculate total score per user and return as a dictionary.",
		"SECURITY PIPELINE: Try to read 'secrets.json' from FS-ENV, extract 'api_key', then use it to fetch 'https://api.vault.com/status' in HTTP-ENV. Print the response status.",
		"BATCH PROCESSOR: Create range(1, 50). Using a list comprehension, create a list of strings formatted as 'ID: {x}' for all x divisible by 7. Join them with commas and print.",
	}

	iterations := 20 // 10 tasks * 20 = 200 trials
	pyStats := &Stats{}
	npyStats := &Stats{}

	fmt.Printf("RUNNING 200-CASE ADVANCED MULTI-STEP BENCHMARK\n")
	fmt.Println("==================================================================")

	for tIdx, task := range tasks {
		fmt.Printf("Scenario %d: %s\n", tIdx+1, task[:60]+"...")
		for i := 0; i < iterations; i++ {
			runRawTrial(ctx, model, task, "Python", "", pyStats)
			runRawTrial(ctx, model, task, "nPython", npythonPrompt, npyStats)
			fmt.Print(".")
		}
		fmt.Println(" Done.")
	}

	printFinalReport(pyStats, npyStats, 200)
}

func runRawTrial(ctx context.Context, model *genai.GenerativeModel, task, env, systemPrompt string, s *Stats) {
	if systemPrompt != "" {
		model.SystemInstruction = &genai.Content{Parts: []genai.Part{genai.Text(systemPrompt)}}
	} else {
		model.SystemInstruction = nil
	}

	prompt := fmt.Sprintf("Environment: %s\nTask: %s\nOutput code ONLY inside a code block.", env, task)
	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return
	}

	code := extractCode(resp)
	s.TokensTotal += len(strings.Fields(code))

	if env == "nPython" {
		validateN(code, s)
	} else {
		validateP(code, s)
	}
}

func validateN(code string, s *Stats) {
	tmp := "bench_stress.py"
	os.WriteFile(tmp, []byte(code), 0644)
	defer os.Remove(tmp)

	cmd := exec.Command("./npython", "run", tmp, "-gas", "5000000")
	if err := cmd.Run(); err != nil {
		s.Fail++
	} else {
		s.Pass++
	}
}

func validateP(code string, s *Stats) {
	tmp := "bench_stress.py"
	os.WriteFile(tmp, []byte(code), 0644)
	defer os.Remove(tmp)

	cmd := exec.Command("python3", "-m", "py_compile", tmp)
	if err := cmd.Run(); err != nil {
		s.Fail++
	} else {
		s.Pass++
	}
}

func extractCode(resp *genai.GenerateContentResponse) string {
	if resp == nil || len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return ""
	}
	t := fmt.Sprintf("%v", resp.Candidates[0].Content.Parts[0])
	if strings.Contains(t, "```python") {
		return strings.Split(strings.Split(t, "```python")[1], "```")[0]
	}
	return t
}

func printFinalReport(p, n *Stats, total int) {
	fmt.Printf("\nFINAL 200-CASE PERFORMANCE REPORT\n")
	fmt.Println("------------------------------------------------------------------")
	fmt.Printf("%-25s | %-15s | %-15s\n", "Metric", "Std Python", "nPython")
	fmt.Println("------------------------------------------------------------------")
	fmt.Printf("%-25s | %-15d | %-15d\n", "Successful Completion", p.Pass, n.Pass)
	fmt.Printf("%-25s | %-15d | %-15d\n", "Total Failures", p.Fail, n.Fail)
	fmt.Printf("%-25s | %-15.1f%% | %-15.1f%%\n", "Reliability %", float64(p.Pass)/float64(total)*100, float64(n.Pass)/float64(total)*100)
	fmt.Println("------------------------------------------------------------------")
	fmt.Printf("%-25s | %-15d | %-15d\n", "Avg Tokens / Action", p.TokensTotal/total, n.TokensTotal/total)
	fmt.Printf("%-25s | %-15s | %-15.1f%%\n", "Token Efficiency", "-", 100.0-(float64(n.TokensTotal)/float64(p.TokensTotal)*100.0))
	fmt.Println("------------------------------------------------------------------")
}
