package commands

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/fatih/color"
	"github.com/rodaine/table"
	"github.com/spf13/cobra"
	"github.com/wcatron/query-projects/internal/outputs"
	"github.com/wcatron/query-projects/internal/projects"
	"github.com/wcatron/query-projects/internal/scripts"
)

var RunCmd = &cobra.Command{
	Use:   "run [scriptName]",
	Short: "Run scripts across all projects in your configuration.",
	RunE: withMetrics(func(cmd *cobra.Command, args []string) error {
		// Get the topics from the command line flags
		topics, _ := cmd.Flags().GetStringSlice("topics")
		count, _ := cmd.Flags().GetBool("count")
		outputFormats, _ := cmd.Flags().GetStringSlice("output")
		scriptName, _ := cmd.Flags().GetString("script")
		return CMD_runScript(scriptName, topics, count, outputFormats, args)
	}),
}

// getScriptInfo executes a script with the --info flag and returns the parsed JSON output.
func getScriptInfo(scriptPath string, pj projects.ProjectsJSON) (outputs.ScriptInfo, error) {
	cmd := exec.Command("deno", "run", "--allow-all", scriptPath, "--info")
	cmd.Dir = pj.RootDirectory
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("%s \n%s", scripts.ScriptPathFmt(scriptPath), output)
		return outputs.ScriptInfo{}, fmt.Errorf("failed to run script with --info: %w", err)
	}

	var info outputs.ScriptInfo
	if err := json.Unmarshal(output, &info); err != nil {
		return outputs.ScriptInfo{}, fmt.Errorf("failed to parse script info: %w", err)
	}

	info.Path = scriptPath

	return info, nil
}

func RunCmdInit(cmd *cobra.Command) {
	cmd.PersistentFlags().Bool("count", false, "Count the unique responses from the script")
	cmd.PersistentFlags().StringSliceP("output", "o", nil, "Comma seperated output formats (md, csv, json)")
	cmd.PersistentFlags().StringSliceP("topics", "t", nil, "Filter projects by topics")
	cmd.PersistentFlags().StringP("script", "s", "", "Path to script to run")
}

func CMD_runScript(scriptName string, topics []string, count bool, outputFormats []string, args []string) error {
	projectsList, err := projects.LoadProjects()
	if err != nil {
		return err
	}
	targets := projects.FilterProjectsByTopics(projectsList.Projects, topics)

	scriptInfo, err := func() (outputs.ScriptInfo, error) {
		if scriptName != "" {
			return getScriptInfo(scriptName, *projectsList) // path or name provided
		}
		return selectScriptInfo(*projectsList) // prompt user to pick
	}()
	if err != nil {
		return err
	}

	targetOveride := projects.InProject(projectsList)
	if targetOveride != nil {
		targets = []projects.Project{*targetOveride}
	}

	if err := runScriptsForProjectsList(projectsList, scriptInfo, targets, count, outputFormats, args); err != nil {
		return fmt.Errorf("error running %s: %w", scriptInfo.Path, err)
	}

	return nil
}

// findScriptFiles returns a list of TypeScript files in the scripts folder
func findScriptFiles() ([]string, error) {
	files, err := os.ReadDir(projects.ScriptsFolder)
	if err != nil {
		return nil, err
	}

	var scriptPaths []string
	for _, f := range files {
		if f.Type().IsRegular() && strings.HasSuffix(f.Name(), ".ts") {
			scriptPaths = append(scriptPaths, filepath.Join(projects.ScriptsFolder, f.Name()))
		}
	}

	if len(scriptPaths) == 0 {
		return nil, fmt.Errorf("No .ts scripts found in the scripts folder: %q", projects.ScriptsFolder)
	}

	return scriptPaths, nil
}

// gatherScriptInfos collects information about each script
func gatherScriptInfos(scriptPaths []string, pj projects.ProjectsJSON) []outputs.ScriptInfo {
	var scriptInfos []outputs.ScriptInfo
	for _, sp := range scriptPaths {
		info, err := getScriptInfo(sp, pj)
		if err != nil {
			fmt.Printf("Error getting info for script %s: %v\n", sp, err)
			continue
		}
		scriptInfos = append(scriptInfos, info)
	}
	return scriptInfos
}

// displayScriptTable shows a formatted table of available scripts
func displayScriptTable(scriptInfos []outputs.ScriptInfo) {
	headerFmt := color.New(color.FgGreen, color.Underline).SprintfFunc()
	columnFmt := color.New(color.FgYellow).SprintfFunc()
	tbl := table.New("#", "Name", "Version", "Output")
	tbl.WithHeaderFormatter(headerFmt).WithFirstColumnFormatter(columnFmt)
	for i, si := range scriptInfos {
		tbl.AddRow(
			fmt.Sprintf("%d", i+1),
			filepath.Base(si.Path),
			si.Version,
			si.Output,
		)
	}
	tbl.Print()
}

// getUserSelection prompts the user to select a script and returns the index
func getUserSelection(scriptInfos []outputs.ScriptInfo) (int, error) {
	fmt.Print("Enter a number: ")
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)
	choice, err := strconv.Atoi(input)
	if err != nil || choice < 1 || choice > len(scriptInfos) {
		return 0, errors.New("invalid selection")
	}
	return choice - 1, nil
}

func selectScriptInfo(pj projects.ProjectsJSON) (outputs.ScriptInfo, error) {
	// Find available scripts
	scriptPaths, err := findScriptFiles()
	if err != nil {
		return outputs.ScriptInfo{}, err
	}

	// Gather information about each script
	scriptInfos := gatherScriptInfos(scriptPaths, pj)
	if len(scriptInfos) == 0 {
		return outputs.ScriptInfo{}, errors.New("no valid scripts found")
	}

	// Display the table of scripts
	displayScriptTable(scriptInfos)

	// Get user selection
	choice, err := getUserSelection(scriptInfos)
	if err != nil {
		return outputs.ScriptInfo{}, err
	}

	return scriptInfos[choice], nil
}

// runScriptsForProjectsList executes the specified .ts script against all projects.
func runScriptsForProjectsList(pj *projects.ProjectsJSON, scriptInfo outputs.ScriptInfo, projectsList []projects.Project, count bool, outputFormats []string, args []string) error {
	var wg sync.WaitGroup
	resultsChan := make(chan outputs.Result, len(projectsList))

	for i, p := range projectsList {
		wg.Add(1)
		go func(project projects.Project, index int) {
			defer wg.Done()
			r, err := scripts.RunScriptForProject(pj, scriptInfo, project.Path, args, true)
			r.Index = index
			if err != nil {
				fmt.Printf("Error in project %s: %v\n", project.Name, err)
			}
			resultsChan <- r
		}(p, i)
	}

	wg.Wait()
	close(resultsChan)

	var results []outputs.Result = collectResults(resultsChan, len(projectsList))

	if len(outputFormats) == 0 {
		if scriptInfo.Output == "text" {
			outputFormats = []string{"csv", "md"}
		} else {
			outputFormats = []string{scriptInfo.Output}
		}
	}

	// If count flag is enabled, count unique responses and print the table
	if count {
		printUniqueResponsesToConsole(results)
	} else {
		// Always print results in markdown to the console
		outputs.PrintToConsole(results)
	}

	// Generate outputs based on the specified or determined formats
	for _, format := range outputFormats {
		switch format {
		case "md":
			outputs.WriteTable(scriptInfo.Path, results)
		case "csv":
			outputs.WriteCSVTable(scriptInfo, results)
		case "json":
			outputs.WriteJSONOutput(scriptInfo.Path, results)
		default:
			fmt.Printf("Unsupported output format: %s\n", format)
		}
	}

	return nil
}

func printUniqueResponsesToConsole(results []outputs.Result) {
	responseCounts := make(map[string]int)
	for _, r := range results {
		responseCounts[r.StdoutText]++
	}

	tbl := table.New("Unique Response", "Count")
	for response, count := range responseCounts {
		tbl.AddRow(strings.TrimSpace(response), fmt.Sprintf("%d", count))
	}
	tbl.Print()
}

func collectResults(resultsChan <-chan outputs.Result, total int) []outputs.Result {
	// Allocate the full slice up front; every slot will be written exactly once.
	results := make([]outputs.Result, total)

	for r := range resultsChan {
		results[r.Index] = r // O(1) placement, no mutex needed
	}
	return results
}
