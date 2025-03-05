package commands

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

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
		row := []string{
			r.projectPath,
			r.status,
			strings.ReplaceAll(r.stdoutText, "\n", "\\n"),
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

// writeCSVTable creates a .csv file summarizing the results.
func writeCSVTable(scriptPath string, results []result) error {
	filename := filepath.Base(scriptPath)
	resultsFilenameForScript := strings.TrimSuffix(filename, ".ts")

	// Open the CSV file for writing
	tableFilePath := filepath.Join(resultsFolder, resultsFilenameForScript+".csv")
	file, err := os.Create(tableFilePath)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write headers
	headers := []string{"Project Path", "Status", "Output (Truncated)"}
	if err := writer.Write(headers); err != nil {
		return err
	}

	// Write data
	for _, r := range results {
		row := []string{r.projectPath, r.status, r.stdoutText}
		if err := writer.Write(row); err != nil {
			return err
		}
	}

	fmt.Printf("Results written to %s\n", tableFilePath)
	return nil
}

// writeJSONOutput creates a .json file summarizing the results.
func writeJSONOutput(scriptPath string, results []result) error {
	filename := filepath.Base(scriptPath)
	resultsFilenameForScript := strings.TrimSuffix(filename, ".ts")

	// Convert results to JSON
	data, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return err
	}

	// Write to file: e.g. results/foo.json
	tableFilePath := filepath.Join(resultsFolder, resultsFilenameForScript+".json")
	if err := os.WriteFile(tableFilePath, data, 0644); err != nil {
		return err
	}
	fmt.Printf("Results written to %s\n", tableFilePath)
	return nil
}
