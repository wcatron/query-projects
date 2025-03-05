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
	"path/filepath"
	"time"
	"bufio"
	"math/rand"

	"github.com/spf13/cobra"
)

var AskCmd = &cobra.Command{
	Use:   "ask <question>",
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

	// Load projects
	projects, err := loadProjects()
	if err != nil {
		return fmt.Errorf("failed to load projects: %w", err)
	}

	// Run the script for a random project
	rand.Seed(time.Now().UnixNano())
	randomIndex := rand.Intn(len(projects.Projects))
	randomProject := projects.Projects[randomIndex]

	cwd, _ := os.Getwd()
	fmt.Printf("Running script for project: %s\n", randomProject.Name)
	result, err := runScriptForProject(filepath.Join(cwd, scriptsFolder, scriptName), randomProject.Path)
	if err != nil {
		return fmt.Errorf("error running script: %w", err)
	}

	fmt.Printf("Result:\n%s\n", result.stdoutText)

	// Prompt user for feedback
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("Did that result look correct? (Enter to continue, 'done' to exit, or type changes): ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		if input == "" {
			// Run for another random project
			randomIndex = rand.Intn(len(projects.Projects))
			randomProject = projects.Projects[randomIndex]
			fmt.Printf("Running script for project: %s\n", randomProject.Name)
			result, err = runScriptForProject(filepath.Join(cwd, scriptsFolder, scriptName), randomProject.Path)
			if err != nil {
				fmt.Printf("Error running script: %v\n", err)
				continue
			}
			fmt.Printf("Result:\n%s\n", result.stdoutText)
		} else if strings.ToLower(input) == "done" {
			break
		} else {
			// Modify the script based on user input
			fmt.Println("Modifying script based on input...")
			err = modifyScriptBasedOnInput(scriptName, input)
			if err != nil {
				fmt.Printf("Error modifying script: %v\n", err)
			} else {
				fmt.Println("Script modified. Running again...")
				result, err = runScriptForProject(filepath.Join(cwd, scriptsFolder, scriptName), randomProject.Path)
				if err != nil {
					fmt.Printf("Error running script: %v\n", err)
				} else {
					fmt.Printf("Result:\n%s\n", result.stdoutText)
				}
			}
		}
	}

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
	// Resolve the path to QUERY_PROMPT.md relative to the executable
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}
	execDir := filepath.Dir(execPath)
	queryPromptPath := filepath.Join(execDir, queryPrompt)

	promptTemplate, err := os.ReadFile(queryPromptPath)
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

	// Log the request and response
	logFile, err := os.OpenFile("openai_requests.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("Failed to open log file: %v\n", err)
	} else {
		defer logFile.Close()
		logEntry := fmt.Sprintf("Request: %s\nResponse Status: %d\n", bodyBytes, resp.StatusCode)
		if resp.StatusCode != http.StatusOK {
			responseBytes, _ := io.ReadAll(resp.Body)
			logEntry += fmt.Sprintf("Response Body: %s\n", responseBytes)
		}
		logFile.WriteString(logEntry + "\n")
	}
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
func modifyScriptBasedOnInput(scriptName, userInput string) error {
	openAIKey := os.Getenv("OPENAI_API_KEY")
	if openAIKey == "" {
		return errors.New("please set the OPENAI_API_KEY environment variable")
	}

	openAIBase := os.Getenv("OPENAI_API_BASE")
	if openAIBase == "" {
		openAIBase = "https://api.openai.com/v1"
	}

	// Read the current script content
	scriptPath := filepath.Join(scriptsFolder, scriptName)
	currentScript, err := os.ReadFile(scriptPath)
	if err != nil {
		return fmt.Errorf("failed to read current script: %w", err)
	}

	// Prepare the prompt with the current script and user input
	prompt := fmt.Sprintf("Here is the current script:\n\n%s\n\nPlease modify it according to the following instructions:\n%s", string(currentScript), userInput)

	body := map[string]interface{}{
		"model":       "gpt-3.5-turbo",
		"max_tokens":  1500,
		"temperature": 0.7,
		"messages": []map[string]string{
			{"role": "system", "content": "You are a helpful assistant who modifies scripts for Deno projects."},
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

	modifiedScript := extractTypeScriptCode(responseData.Choices[0].Message.Content)
	if modifiedScript == "" {
		return errors.New("failed to extract TypeScript code from the response")
	}

	// Write the modified script back to the file
	if err := os.WriteFile(scriptPath, []byte(modifiedScript), 0644); err != nil {
		return err
	}

	fmt.Printf("Modified script saved to: %s\n", scriptPath)
	return nil
}
