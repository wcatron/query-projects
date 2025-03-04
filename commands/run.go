package commands

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

var RunCmd = &cobra.Command{
	Use:   "run [scriptName]",
	Short: "Run a script (or all .ts scripts) across all projects",
	Args:  cobra.MaximumNArgs(1),
	RunE: WrapWithMetrics(func(cmd *cobra.Command, args []string) error {
		// Optional argument: the user can provide a script name or path
		var scriptName string
		if len(args) == 1 {
			scriptName = args[0]
		}
		return runScript(cmd, scriptName)
	}),
	cmd.PersistentFlags().StringSliceP("output", "o", nil, "Specify output formats (md, csv, json)")
}

func determineBestOutputFormat(results []result) []string {
	jsonCount := 0
	singleLineCount := 0

	for _, r := range results {
		if isValidJSON(r.stdoutText) {
			jsonCount++
		}
		if isSingleLine(r.stdoutText) {
			singleLineCount++
		}
	}

	if jsonCount > len(results)/2 {
		return []string{"json"}
	} else if singleLineCount == len(results) {
		return []string{"md", "csv"}
	}

	return []string{"md"} // Default to markdown if no clear format is determined
}

func isValidJSON(s string) bool {
	var js map[string]interface{}
	return json.Unmarshal([]byte(s), &js) == nil
}

func isSingleLine(s string) bool {
	return len(strings.Split(s, "\n")) == 1
}

func RunCmdInit(cmd *cobra.Command) {
	cmd.PersistentFlags().StringSliceP("topics", "t", nil, "Filter projects by topics")
	// Add count flag
	cmd.PersistentFlags().Bool("count", false, "Count the unique responses from the script")
}

// runScript decides which scripts to run:
//  1) If the user gave a path (e.g., `scripts/foo.ts` or `/abs/path.ts`), just run that.
//  2) If the user gave a simple filename (e.g. `foo.ts`), we prepend the scripts folder.
//  3) If no name was given, prompt the user to pick a .ts file from the scripts folder.
func runScript(cmd *cobra.Command, scriptName string) error {
	// Get the topics from the command line flags
	topics, err := cmd.Flags().GetStringSlice("topics")
	count, err := cmd.Flags().GetBool("count")

	projects, err := loadProjects()
	if err != nil {
		return err
	}

	filteredProjects := filterProjectsByTopics(projects.Projects, topics)
	var scriptPaths []string

	// Case 1 & 2: The user specified some script name/path
	if scriptName != "" {
		// If the user provided a path containing a slash or is absolute,
		// we treat it as the full path. Otherwise, prepend scripts folder.
		if filepath.IsAbs(scriptName) || strings.Contains(scriptName, string(os.PathSeparator)) {
			scriptPaths = []string{scriptName}
		} else {
			scriptPaths = []string{filepath.Join(scriptsFolder, scriptName)}
		}
	} else {
		// Case 3: No script name -> prompt user to select from scripts folder
		files, err := os.ReadDir(scriptsFolder)
		if err != nil {
			return fmt.Errorf("failed to read scripts folder: %w", err)
		}

		// Collect all *.ts files
		for _, f := range files {
			if f.Type().IsRegular() && strings.HasSuffix(f.Name(), ".ts") {
				scriptPaths = append(scriptPaths, filepath.Join(scriptsFolder, f.Name()))
			}
		}
		if len(scriptPaths) == 0 {
			fmt.Println("No .ts scripts found in the scripts folder.")
			return nil
		}

		// Ask user to pick a script
		fmt.Println("Select a script to run:")
		for i, sp := range scriptPaths {
			fmt.Printf("%d) %s\n", i+1, filepath.Base(sp))
		}
		fmt.Print("Enter a number: ")

		reader := bufio.NewReader(os.Stdin)
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)
		choice, err := strconv.Atoi(input)
		if err != nil || choice < 1 || choice > len(scriptPaths) {
			return errors.New("invalid selection")
		}
		// User picks exactly one script
		scriptPaths = []string{scriptPaths[choice-1]}
	}

	// Actually run the script(s)
	for _, sp := range scriptPaths {
		if err := runScriptsForAllProjects(sp, filteredProjects, count); err != nil {
			fmt.Printf("Error while running script %s: %v\n", sp, err)
		}
	}
	
	return nil
}

// We'll capture results for printing
type result struct {
	projectPath string
	status      string
	stdoutText  string
	stderrText  string
}

// runScriptsForAllProjects executes the specified .ts script against all projects.
func runScriptsForAllProjects(scriptPath string, projects []Project, count bool, outputFormats []string) error {
	var results []result

	// Get cwd
	cwd, _ := os.Getwd()

	for _, p := range projects {
		r, err := runScriptForProject(filepath.Join(cwd, scriptPath), p.Path)
		if err != nil {
			fmt.Printf("Error in project %s: %v\n", p.Name, err)
		}
		results = append(results, r)
	}

	// Determine the best output format if not specified
	if len(outputFormats) == 0 {
		outputFormats = determineBestOutputFormat(results)
	}

	// Generate outputs based on the specified or determined formats
	for _, format := range outputFormats {
		switch format {
		case "md":
			writeMarkdownTable(scriptPath, results)
		case "csv":
			writeCSVTable(scriptPath, results)
		case "json":
			writeJSONOutput(scriptPath, results)
		default:
			fmt.Printf("Unsupported output format: %s\n", format)
		}
	}

	return nil
}

func countUniqueResponses(results []result) {
	responseCounts := make(map[string]int)
	for _, r := range results {
		responseCounts[r.stdoutText]++
	}

	fmt.Println("--------------------------------------------------")
	fmt.Printf("| %-30s | %-10s |\n", "Unique Response", "Count")
	fmt.Println("--------------------------------------------------")

	for response, count := range responseCounts {
		truncatedResponse := truncateOutput(response, 30) // Adjust length as desired
		fmt.Printf("| %-30s | %-10d |\n", truncatedResponse, count)
	}

	fmt.Println("--------------------------------------------------")
}

// runScriptForProject runs a TypeScript script (with Deno) in the specified project directory.
func runScriptForProject(script, projectPath string) (result, error) {
	fmt.Printf("Running %s for %s...\n", script, projectPath)

	cmd := exec.Command("deno", "run", "--allow-read", "--allow-run", script)
	cmd.Dir = projectPath

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return result{}, err
	}
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return result{}, err
	}

	if err := cmd.Start(); err != nil {
		return result{}, err
	}

	stdoutBytes, _ := io.ReadAll(stdoutPipe)
	stderrBytes, _ := io.ReadAll(stderrPipe)

	// Wait for command completion
	err = cmd.Wait()

	stdoutText := string(stdoutBytes)
	stderrText := string(stderrBytes)

	if len(stdoutText) > 0 {
		fmt.Printf("[%s] stdout:\n%s\n", projectPath, stdoutText)
	}
	if len(stderrText) > 0 {
		fmt.Printf("[%s] stderr:\n%s\n", projectPath, stderrText)
	}

	var status string
	if err == nil {
		fmt.Printf("Successfully ran %s for %s\n", script, projectPath)
		status = "Success"
	} else {
		if exitErr, ok := err.(*exec.ExitError); ok {
			status = fmt.Sprintf("Failed (exit code %d)", exitErr.ExitCode())
			fmt.Printf("Script %s failed for %s: %s\n", script, projectPath, exitErr.Error())
		} else {
			status = "Error"
			fmt.Printf("Error running script %s for %s: %v\n", script, projectPath, err)
		}
	}

	return result{
		projectPath: projectPath,
		status:      status,
		stdoutText:  stdoutText,
		stderrText:  stderrText,
	}, nil
}

// printResultTable prints a simple ASCII-like table of results with the script's stdout (truncated).
func printResultTable(results []result) {
	fmt.Println("--------------------------------------------------------------------------------")
	fmt.Printf("| %-30s | %-10s | %-30s |\n", "Project Path", "Status", "Output (Truncated)")
	fmt.Println("--------------------------------------------------------------------------------")

	for _, r := range results {
		truncated := truncateOutput(r.stdoutText, 50) // adjust length as desired
		fmt.Printf("| %-30s | %-10s | %-30s |\n", r.projectPath, r.status, truncated)
	}
	fmt.Println("--------------------------------------------------------------------------------")
}

// writeMarkdownTable creates a .md table summarizing the results with their output (truncated).
func writeMarkdownTable(scriptPath string, results []result) error {
	filename := filepath.Base(scriptPath)
	resultsFilenameForScript := strings.TrimSuffix(filename, ".ts")

	// Build the table lines
	var sb strings.Builder
	headers := []string{"Project Path", "Status", "Output (Truncated)"}
	sb.WriteString("| " + strings.Join(headers, " | ") + " |\n")
	sb.WriteString("| " + strings.Repeat("--- | ", len(headers)) + "\n")

	for _, r := range results {
		shortOutput := truncateOutput(r.stdoutText, 100) // 100 chars for markdown table
		row := []string{
			r.projectPath,
			r.status,
			strings.ReplaceAll(shortOutput, "\n", "\\n"),
		}
		sb.WriteString("| " + strings.Join(row, " | ") + " |\n")
	}

	// Write to file: e.g. results/foo.md
	tableFilePath := filepath.Join(resultsFolder, resultsFilenameForScript+".md")
	if err := os.WriteFile(tableFilePath, []byte(sb.String()), 0644); err != nil {
		return err
	}
	fmt.Printf("Results written to %s\n", tableFilePath)
	return nil
}

// truncateOutput is a small helper to avoid huge output in tables.
func truncateOutput(output string, maxLen int) string {
	output = strings.TrimSpace(output)
	if len(output) > maxLen {
		return output[:maxLen] + "..."
	}
	return output
}
