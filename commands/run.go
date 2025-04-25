package commands

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/charmbracelet/glamour"
	"github.com/spf13/cobra"
)

type ScriptInfo struct {
	Path    string   `json:"path"`
	Version string   `json:"version"`
	Output  string   `json:"output"`
	Columns []string `json:"columns"`
}

var RunCmd = &cobra.Command{
	Use:   "run [scriptName]",
	Short: "Run scripts across all projects in your configuration.",
	Args:  cobra.MaximumNArgs(1),
	RunE: WrapWithMetrics(func(cmd *cobra.Command, args []string) error {
		// Optional argument: the user can provide a script name or path
		var scriptName string
		if len(args) == 1 {
			scriptName = args[0]
		}
		// Get the topics from the command line flags
		topics, _ := cmd.Flags().GetStringSlice("topics")
		count, _ := cmd.Flags().GetBool("count")
		outputFormats, _ := cmd.Flags().GetStringSlice("output")
		return cmd_runScript(scriptName, topics, count, outputFormats)
	}),
}

// formatCSVOutput formats CSV output based on column headers.
func formatCSVOutput(csvText string, columns []string) string {
	var sb strings.Builder
	sb.WriteString(strings.Join(columns, ",") + "\n")
	sb.WriteString(csvText)
	return sb.String()
}

// getScriptInfo executes a script with the --info flag and returns the parsed JSON output.
func getScriptInfo(scriptPath string) (ScriptInfo, error) {
	cmd := exec.Command("deno", "run", "--allow-all", scriptPath, "--info")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return ScriptInfo{}, fmt.Errorf("failed to run script with --info: %w", err)
	}

	var info ScriptInfo
	if err := json.Unmarshal(output, &info); err != nil {
		return ScriptInfo{}, fmt.Errorf("failed to parse script info: %w", err)
	}

	info.Path = scriptPath

	return info, nil
}

// printMarkdownToConsole renders the results in markdown format to the console using Glamour.
func printMarkdownToConsole(results []result) {
	var sb strings.Builder = createMarkdownString(results)

	// Render the markdown table using Glamour
	out, err := glamour.Render(sb.String(), "dark")
	if err != nil {
		fmt.Println("Error rendering markdown:", err)
		return
	}
	fmt.Print(out)
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
	cmd.PersistentFlags().Bool("count", false, "Count the unique responses from the script")
	cmd.PersistentFlags().StringSliceP("output", "o", nil, "Specify output formats (md, csv, json)")
}

func cmd_runScript(scriptName string, topics []string, count bool, outputFormats []string) error {
	projects, err := loadProjects()
	if err != nil {
		return err
	}
	targets := filterProjectsByTopics(projects.Projects, topics)

	scriptInfo, err := func() (ScriptInfo, error) {
		if scriptName != "" {
			return getScriptInfo(scriptName) // path or name provided
		}
		return selectScriptInfo() // prompt user to pick
	}()
	if err != nil {
		return err
	}

	if err := runScriptsForAllProjects(scriptInfo, targets, count, outputFormats); err != nil {
		return fmt.Errorf("running %s: %w", scriptInfo.Path, err)
	}

	return nil
}

func selectScriptInfo() (ScriptInfo, error) {
	var scriptPaths []string
	var scriptInfos []ScriptInfo

	files, err := os.ReadDir(scriptsFolder)
	if err != nil {
		return ScriptInfo{}, err
	}

	// Collect all *.ts files
	for _, f := range files {
		if f.Type().IsRegular() && strings.HasSuffix(f.Name(), ".ts") {
			scriptPaths = append(scriptPaths, filepath.Join(scriptsFolder, f.Name()))
		}
	}
	if len(scriptPaths) == 0 {
		fmt.Println()
		return ScriptInfo{}, fmt.Errorf("No .ts scripts found in the scripts folder: %q", scriptsFolder)
	}

	// Gather script information
	for _, sp := range scriptPaths {
		info, err := getScriptInfo(sp)
		if err != nil {
			fmt.Printf("Error getting info for script %s: %v\n", sp, err)
			continue
		}
		scriptInfos = append(scriptInfos, info)
	}

	// Display script information in a table
	var sb strings.Builder
	headers := []string{"#", "Name", "Version", "Output"}
	sb.WriteString("| " + strings.Join(headers, " | ") + " |\n")
	sb.WriteString("| " + strings.Repeat("--- | ", len(headers)) + "\n")

	for i, si := range scriptInfos {
		row := []string{
			fmt.Sprintf("%d", i+1),
			filepath.Base(si.Path),
			si.Version,
			si.Output,
		}
		sb.WriteString("| " + strings.Join(row, " | ") + " |\n")
	}

	// Render the markdown table using Glamour
	out, err := glamour.Render(sb.String(), "dark")
	if err != nil {
		fmt.Println("Error rendering markdown:", err)
		return ScriptInfo{}, err
	}
	fmt.Print(out)

	// Ask user to pick a script
	fmt.Print("Enter a number: ")
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)
	choice, err := strconv.Atoi(input)
	if err != nil || choice < 1 || choice > len(scriptInfos) {
		return ScriptInfo{}, errors.New("invalid selection")
	}
	return scriptInfos[choice-1], nil

}

// We'll capture results for printing
type result struct {
	projectPath string
	status      string
	stdoutText  string
	stderrText  string
	index       int
}

func collectResults(resultsChan <-chan result, total int) []result {
	// Allocate the full slice up front; every slot will be written exactly once.
	results := make([]result, total)

	for r := range resultsChan {
		results[r.index] = r // O(1) placement, no mutex needed
	}
	return results
}

// runScriptsForAllProjects executes the specified .ts script against all projects.
func runScriptsForAllProjects(scriptInfo ScriptInfo, projects []Project, count bool, outputFormats []string) error {
	var wg sync.WaitGroup
	resultsChan := make(chan result, len(projects))

	for i, p := range projects {
		wg.Add(1)
		go func(project Project, index int) {
			defer wg.Done()
			r, err := runScriptForProject(scriptInfo, project.Path)
			r.index = i
			if err != nil {
				fmt.Printf("Error in project %s: %v\n", project.Name, err)
			}
			resultsChan <- r
		}(p, i)
	}

	wg.Wait()
	close(resultsChan)

	var results []result = collectResults(resultsChan, len(projects))

	if len(outputFormats) == 0 {
		outputFormats = determineBestOutputFormat(results)
	}

	// If count flag is enabled, count unique responses and print the table
	if count {
		printUniqueResponsesToConsole(results)
	} else {
		// Always print results in markdown to the console
		printMarkdownToConsole(results)
	}

	// Generate outputs based on the specified or determined formats
	for _, format := range outputFormats {
		switch format {
		case "md":
			writeMarkdownTable(scriptInfo.Path, results)
		case "csv":
			writeCSVTable(scriptInfo, results)
		case "json":
			writeJSONOutput(scriptInfo.Path, results)
		default:
			fmt.Printf("Unsupported output format: %s\n", format)
		}
	}

	return nil
}

func printUniqueResponsesToConsole(results []result) {
	responseCounts := make(map[string]int)
	for _, r := range results {
		responseCounts[r.stdoutText]++
	}

	var sb strings.Builder
	headers := []string{"Unique Response", "Count"}
	sb.WriteString("| " + strings.Join(headers, " | ") + " |\n")
	sb.WriteString("| " + strings.Repeat("--- | ", len(headers)) + "\n")

	for response, count := range responseCounts {
		row := []string{strings.TrimSpace(response), fmt.Sprintf("%d", count)}
		sb.WriteString("| " + strings.Join(row, " | ") + " |\n")
	}

	// Render the markdown table using Glamour
	out, err := glamour.Render(sb.String(), "dark")
	if err != nil {
		fmt.Println("Error rendering markdown:", err)
		return
	}
	fmt.Print(out)
}

// runScriptForProject runs a TypeScript script (with Deno) in the specified project directory.
func runScriptForProject(scriptInfo ScriptInfo, projectPath string) (result, error) {
	fmt.Printf("Running %s for %s...\n", scriptInfo.Path, projectPath)

	// Get cwd
	cwd, _ := os.Getwd()
	scriptPath := filepath.Join(cwd, scriptInfo.Path)

	cmd := exec.Command("deno", "run", "--allow-all", scriptPath)
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

	// Format CSV output if applicable
	if scriptInfo.Output == "csv" && len(stdoutText) > 0 {
		fmt.Printf("[%s] CSV Output:\n", projectPath)
		fmt.Println(formatCSVOutput(stdoutText, scriptInfo.Columns))
	} else if len(stdoutText) > 0 {
		fmt.Printf("[%s] stdout:\n%s\n", projectPath, stdoutText)
	}
	if len(stderrText) > 0 {
		fmt.Printf("[%s] stderr:\n%s\n", projectPath, stderrText)
	}

	var status string
	if err == nil {
		fmt.Printf("Successfully ran %s for %s\n", scriptInfo.Path, projectPath)
		status = "Success"
	} else {
		if exitErr, ok := err.(*exec.ExitError); ok {
			status = fmt.Sprintf("Failed (exit code %d)", exitErr.ExitCode())
			fmt.Printf("Script %s failed for %s: %s\n", scriptInfo.Path, projectPath, exitErr.Error())
		} else {
			status = "Error"
			fmt.Printf("Error running script %s for %s: %v\n", scriptInfo.Path, projectPath, err)
		}
	}

	return result{
		projectPath: projectPath,
		status:      status,
		stdoutText:  strings.TrimSpace(stdoutText),
		stderrText:  strings.TrimSpace(stderrText),
	}, nil
}
