package commands

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/wcatron/query-projects/internal/projects"
	"github.com/wcatron/query-projects/internal/scripts"
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
		return CMD_ask(question)
	},
}

// CMD_ask calls the OpenAI API with the question, extracts the TypeScript code,
// and saves it to ./scripts/<question>.ts.
func CMD_ask(question string) error {
	scriptName := strings.ReplaceAll(strings.ToLower(question), " ", "-") + ".ts"
	err := generateScriptForQuestion(question, scriptName)
	if err != nil {
		return err
	}
	fmt.Printf("Generated script: ./%s/%s\n", projects.ScriptsFolder, scriptName)

	// Load projectsList
	projectsList, err := projects.LoadProjects()
	if err != nil {
		return fmt.Errorf("failed to load projects: %w", err)
	}

	// Run the script for a random project
	rand.Seed(time.Now().UnixNano())
	randomIndex := rand.Intn(len(projectsList.Projects))
	randomProject := projectsList.Projects[randomIndex]

	cwd, _ := os.Getwd()
	fmt.Printf("Running script for project: %s\n", randomProject.Name)
	scriptInfo, err := getScriptInfo(filepath.Join(cwd, projects.ScriptsFolder, scriptName), *projectsList)
	if err != nil {
		return fmt.Errorf("failed to get script info: %w", err)
	}
	result, err := scripts.RunScriptForProject(projectsList, scriptInfo, randomProject.Path, []string{}, true)
	if err != nil {
		return fmt.Errorf("error running script: %w", err)
	}

	fmt.Printf("Result:\n%s\n", result.StdoutText)

	// Prompt user for feedback
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("Did that result look correct? (Enter to continue, 'done' to exit, or type changes): ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		if input == "" {
			// Run again with the same script
		} else if strings.ToLower(input) == "done" {
			break
		} else {
			// Modify the script based on user input
			fmt.Println("Modifying script based on input...")
			err = modifyScriptBasedOnInput(scriptName, input)
			if err != nil {
				fmt.Printf("Error modifying script: %v\n", err)
			} else {
				fmt.Println("Script modified.")
			}
		}
		// Run for another random project
		randomIndex = rand.Intn(len(projectsList.Projects))
		randomProject = projectsList.Projects[randomIndex]
		fmt.Printf("Running next script for project: %s\n", randomProject.Name)
		scriptInfo, err = getScriptInfo(filepath.Join(cwd, projects.ScriptsFolder, scriptName), *projectsList)
		if err != nil {
			fmt.Printf("Failed to get script info: %v\n", err)
			continue
		}
		result, err = scripts.RunScriptForProject(projectsList, scriptInfo, randomProject.Path, []string{}, true)
		if err != nil {
			fmt.Printf("Error running script: %v\n", err)
		} else {
			fmt.Printf("Result:\n%s\n", result.StdoutText)
		}
	}

	return nil
}

func callOpenAI(prompt string) (string, error) {
	openAIKey := os.Getenv("OPENAI_API_KEY")
	if openAIKey == "" {
		return "", errors.New("please set the OPENAI_API_KEY environment variable")
	}

	openAIBase := os.Getenv("OPENAI_API_BASE")
	if openAIBase == "" {
		openAIBase = "https://api.openai.com/v1"
	}

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
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+openAIKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to call OpenAI: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		responseBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("non-200 status from OpenAI: %d\n%s", resp.StatusCode, string(responseBytes))
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
		return "", fmt.Errorf("failed to parse OpenAI response: %w", err)
	}

	if len(responseData.Choices) == 0 {
		return "", errors.New("no choices returned from OpenAI")
	}

	return responseData.Choices[0].Message.Content, nil
}

func generateScriptForQuestion(question, scriptName string) error {
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}
	execDir := filepath.Dir(execPath)
	QueryPromptPath := filepath.Join(execDir, projects.QueryPrompt)

	promptTemplate, err := os.ReadFile(QueryPromptPath)
	if err != nil {
		fmt.Printf("Warning: Could not read %s, using question directly.\n", projects.QueryPrompt)
		promptTemplate = []byte("{{QUESTION}}")
	}

	prompt := strings.ReplaceAll(string(promptTemplate), "{{QUESTION}}", question)

	fmt.Println("Generating script from OpenAI...")

	responseContent, err := callOpenAI(prompt)
	if err != nil {
		return err
	}

	generatedScript := projects.ExtractTypeScriptCode(responseContent)
	if generatedScript == "" {
		return errors.New("failed to extract TypeScript code from the response")
	}

	if err := os.MkdirAll(projects.ScriptsFolder, 0755); err != nil {
		return err
	}

	scriptPath := fmt.Sprintf("%s/%s", projects.ScriptsFolder, scriptName)
	if err := os.WriteFile(scriptPath, []byte(generatedScript), 0644); err != nil {
		return err
	}

	fmt.Printf("Generated script saved to: %s\n", scriptPath)
	return nil
}
func modifyScriptBasedOnInput(scriptName, userInput string) error {
	scriptPath := filepath.Join(projects.ScriptsFolder, scriptName)
	currentScript, err := os.ReadFile(scriptPath)
	if err != nil {
		return fmt.Errorf("failed to read current script: %w", err)
	}

	prompt := fmt.Sprintf("Attatched is a current version of a script. Add the new requirement to a requirements list at the top of the file. Return the entire file in triple slashes for example: \n ```\n new file contents... \n``` end example. Here is the current file\n```%s```\n Modify it according to the following instructions:\n%s", string(currentScript), userInput)

	responseContent, err := callOpenAI(prompt)
	if err != nil {
		return err
	}

	modifiedScript := projects.ExtractTypeScriptCode(responseContent)
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
func logOpenAIRequest(requestBody []byte, responseBody []byte) {
	logFile, err := os.OpenFile("openai_requests.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("Failed to open log file: %v\n", err)
		return
	}
	defer logFile.Close()

	logEntry := fmt.Sprintf("Request: %s\nResponse Body: %s\n", requestBody, responseBody)
	logFile.WriteString(logEntry + "\n")
}
