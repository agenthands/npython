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

func main() {
	apiKey := os.Getenv("GEMINI_API_KEY")
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil { panic(err) }
	defer client.Close()

	model := client.GenerativeModel("gemini-2.0-flash")
	temp := float32(0.7)
	model.Temperature = &temp

	promptData, _ := os.ReadFile("docs/PROMPT.md")
	datacardData, _ := os.ReadFile("docs/DATACARD.md")
	npythonPrompt := string(promptData) + "\n\n" + string(datacardData)
	model.SystemInstruction = &genai.Content{Parts: []genai.Part{genai.Text(npythonPrompt)}}

	tasks := []string{
		"AGENT TASK: Fetch repository list from 'https://api.github.com/orgs/google/repos', parse the JSON, filter only those where 'language' is 'Go', calculate the total 'stargazers_count' for these repos, and write the final sum string to a file named 'stars_report.txt'. Use scope HTTP-ENV and FS-ENV.",
		"DATA TASK: You have a list of dictionaries representing transactions: [{'id': 1, 'amount': 100, 'type': 'credit'}, {'id': 2, 'amount': 50, 'type': 'debit'}]. Create a script that filters for 'credit' types, calculates the square of each amount, and returns a new list of these squared values using a list comprehension. Print the result.",
		"MATH TASK: Write a function that calculates the first 10 prime numbers. Format them as a single string separated by hyphens (e.g. '2-3-5...'), and use a scoped block to save this string to 'primes.log'.",
		"RECONCILIATION TASK: Compare two dictionaries: 'local_inventory' and 'master_inventory'. Find keys that exist in local but not in master, and keys where the values (counts) differ. Create a 'discrepancy_report' dictionary with these findings and print its length.",
		"CONFIG TASK: Parse a nested JSON configuration string. If the 'environment' key is 'production', set 'debug' to False and 'timeout' to 60. If 'staging', set 'debug' to True and 'timeout' to 30. Return the updated dictionary and print the 'timeout' value.",
	}

	fmt.Println("DIAGNOSING NPYTHON FAILURES...")
	logFile, _ := os.Create("hallucinations.log")
	defer logFile.Close()

	iterations := 10
	for _, task := range tasks {
		for i := 0; i < iterations; i++ {
			resp, _ := model.GenerateContent(ctx, genai.Text(task))
			code := extractCode(resp)
			
					tmp := "diag_tmp.py"
					os.WriteFile(tmp, []byte(code), 0644)
					cmd := exec.Command("./npython", "run", tmp, "-gas", "5000000")
					out, err := cmd.CombinedOutput()
					os.Remove(tmp)
			if err != nil {
				logFile.WriteString(fmt.Sprintf("TASK: %s\nCODE:\n%s\nERROR: %s\n-------------------\n", task, code, string(out)))
			}
		}
	}
	fmt.Println("Failures logged to hallucinations.log")
}

func extractCode(resp *genai.GenerateContentResponse) string {
	if resp == nil || len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 { return "" }
	t := fmt.Sprintf("%v", resp.Candidates[0].Content.Parts[0])
	if strings.Contains(t, "```python") {
		return strings.Split(strings.Split(t, "```python")[1], "```")[0]
	}
	return t
}
