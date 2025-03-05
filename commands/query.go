package commands

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var QueryCmd = &cobra.Command{
	Use:   "query <question>",
	Short: "Generate a TypeScript script from an OpenAI query.",
	Args:  cobra.ArbitraryArgs, // Allows spaces in the question
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return errors.New("please provide a question after 'query'")
		}
		question := strings.Join(args, " ")
		return queryQuestion(question)
	},
}

// queryQuestion calls the OpenAI API with the question, extracts the TypeScript code,
// and saves it to ./scripts/<question>.ts.
func queryQuestion(question string) error {
	scriptName := strings.ReplaceAll(strings.ToLower(question), " ", "-") + ".ts"
	err := generateScriptForQuestion(question, scriptName)
	if err != nil {
		return err
	}
	fmt.Printf("Generated script: ./%s/%s\n", scriptsFolder, scriptName)
	return nil
}

// generateScriptForQuestion calls OpenAI, extracts TS code blocks, and writes them to a .ts file.
func generateScriptForQuestion(question, scriptName string) error {
	openAIKey := os.Getenv("OPENAI_API_KEY")
	if openAIKey == "" {
		return errors.New("please set the OPENAI_API_KEY environment variable")
	}

	openAIBase := os.Getenv("OPENAI_API_BASE")
	if openAIBase == "" {
		openAIBase = "https://api.openai.com/v1"
	}

	// If you have a template file (QUERY_PROMPT.md), read it; otherwise, just use the question as the prompt
	promptTemplate, err := os.ReadFile(queryPrompt)
	if err != nil {
		fmt.Printf("Warning: Could not read %s, using question directly.\n", queryPrompt)
		promptTemplate = []byte("{{QUESTION}}")
	}

	prompt := strings.ReplaceAll(string(promptTemplate), "{{QUESTION}}", question)

	fmt.Println("Generating script from OpenAI...")

	body := map[string]interface{}{
		"model":       "gpt-3.5-turbo",
		"max_tokens":  1500,
		"temperature": 0.7,
		"messages": []map[string]string{
			{"role": "system", "content": "You are a helpful assistant who writes scripts for Deno projects."},
			{"role": "user", "content": prompt},
		},
	}

	bodyBytes, _ := json.Marshal(body)
	req, err := http.NewRequest(http.MethodPost, openAIBase+"/chat/completions", bytes.NewBuffer(bodyBytes))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+openAIKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to call OpenAI: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		responseBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("non-200 status from OpenAI: %d\n%s", resp.StatusCode, string(responseBytes))
	}

	// Parse the JSON response
	var responseData struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	err = json.NewDecoder(resp.Body).Decode(&responseData)
	if err != nil {
		return fmt.Errorf("failed to parse OpenAI response: %w", err)
	}

	if len(responseData.Choices) == 0 {
		return errors.New("no choices returned from OpenAI")
	}

	generatedScript := extractTypeScriptCode(responseData.Choices[0].Message.Content)
	if generatedScript == "" {
		return errors.New("failed to extract TypeScript code from the response")
	}

	// Ensure scripts directory
	if err := os.MkdirAll(scriptsFolder, 0755); err != nil {
		return err
	}

	scriptPath := fmt.Sprintf("%s/%s", scriptsFolder, scriptName)
	if err := os.WriteFile(scriptPath, []byte(generatedScript), 0644); err != nil {
		return err
	}

	fmt.Printf("Generated script saved to: %s\n", scriptPath)
	return nil
}
