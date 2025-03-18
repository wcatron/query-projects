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

	"github.com/charmbracelet/glamour"
	"github.com/spf13/cobra"
)

type ScriptInfo struct {
	Path    string
	Version string
	Output  string
	Columns string
}

// projectHasSkipField checks if the project has the skip field explicitly set.
func projectHasSkipField(p Project) bool {
	return p.Skip || !reflect.ValueOf(p).FieldByName("Skip").IsZero()

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
		return runScript(cmd, scriptName)
	}),
}

// formatCSVOutput formats CSV output based on column headers.
func formatCSVOutput(csvText, columns string) string {
	var sb strings.Builder
	headers := strings.Split(columns, ",")
	sb.WriteString(strings.Join(headers, ",") + "\n")
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

	var info map[string]string
	if err := json.Unmarshal(output, &info); err != nil {
		return ScriptInfo{}, fmt.Errorf("failed to parse script info: %w", err)
	}

	return ScriptInfo{
		Path:    scriptPath,
		Version: info["version"],
		Output:  info["output"],
		Columns: info["columns"],
	}, nil
}

// printMarkdownToConsole renders the results in markdown format to the console using Glamour.
func printMarkdownToConsole(results []result) {
	var sb strings.Builder
	headers := []string{"Project Path", "Status", "Output"}
	sb.WriteString("| " + strings.Join(headers, " | ") + " |\n")
	sb.WriteString("| " + strings.Repeat("--- | ", len(headers)) + "\n")

	for _, r := range results {
		row := []string{
			r.projectPath,
			r.status,
			r.stdoutText,
		}
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

// runScript decides which scripts to run:
//  1. If the user gave a path (e.g., `scripts/foo.ts` or `/abs/path.ts`), just run that.
//  2. If the user gave a simple filename (e.g. `foo.ts`), we prepend the scripts folder.
//  3. If no name was given, prompt the user to pick a .ts file from the scripts folder.
func runScript(cmd *cobra.Command, scriptName string) error {
	// Get the topics from the command line flags
	topics, err := cmd.Flags().GetStringSlice("topics")
	count, err := cmd.Flags().GetBool("count")
	outputFormats, err := cmd.Flags().GetStringSlice("output")

	projects, err := loadProjects()
	if err != nil {
		return err
	}

	filteredProjects := filterProjectsByTopics(projects.Projects, topics)
	var scriptPaths []string
	var scriptInfos []ScriptInfo

	// Case 1 & 2: The user specified some script name/path
	if scriptName != "" {
		// If the user provided a path containing a slash or is absolute,
		// we treat it as the full path. Otherwise, prepend scripts folder.
		info, _ := getScriptInfo(scriptName)
		scriptInfos = []ScriptInfo{info}
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
			return err
		}
		fmt.Print(out)

		// Ask user to pick a script
		fmt.Print("Enter a number: ")
		reader := bufio.NewReader(os.Stdin)
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)
		choice, err := strconv.Atoi(input)
		if err != nil || choice < 1 || choice > len(scriptInfos) {
			return errors.New("invalid selection")
		}
		// User picks exactly one script
		scriptInfos = []ScriptInfo{scriptInfos[choice-1]}

	}

	// Actually run the script(s)
	for _, si := range scriptInfos {
		// Skip projects marked with "skip": true or if the skip field is absent
		if p.Skip || !projectHasSkipField(p) {
			fmt.Printf("Skipping project: %s\n", p.Name)
			continue
		}
		if err := runScriptsForAllProjects(si, filteredProjects, count, outputFormats); err != nil {
			fmt.Printf("Error while running script %s: %v\n", si.Path, err)
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
func runScriptsForAllProjects(scriptInfo ScriptInfo, projects []Project, count bool, outputFormats []string) error {
	var results []result

	for _, p := range projects {

		r, err := runScriptForProject(scriptInfo, p.Path)
		if err != nil {
			fmt.Printf("Error in project %s: %v\n", p.Name, err)
		}
		results = append(results, r)
	}

	// Determine the best output format if not specified
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
			writeCSVTable(scriptInfo.Path, results)
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
