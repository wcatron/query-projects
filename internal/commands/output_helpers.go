package commands

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/wcatron/query-projects/internal/projects"
)

func createMarkdownString(results []result) strings.Builder {
	var sb strings.Builder
	headers := []string{"Project Path", "Status", "Output"}
	sb.WriteString("| " + strings.Join(headers, " | ") + " |\n")
	sb.WriteString("| " + strings.Repeat("--- | ", len(headers)) + "\n")

	for _, r := range results {
		lines := strings.Split(r.stdoutText, "\n")
		for _, line := range lines {
			row := []string{
				r.projectPath,
				r.status,
				line,
			}
			sb.WriteString("| " + strings.Join(row, " | ") + " |\n")
		}
	}

	return sb
}

// writeMarkdownTable creates a .md table summarizing the results with their output).
func writeMarkdownTable(scriptPath string, results []result) error {
	filename := filepath.Base(scriptPath)
	resultsFilenameForScript := strings.TrimSuffix(filename, ".ts")

	// Build the table lines
	var sb strings.Builder = createMarkdownString(results)

	// Write to file: e.g. results/foo.md
	tableFilePath := filepath.Join(projects.ResultsFolder, resultsFilenameForScript+".md")
	if err := os.WriteFile(tableFilePath, []byte(sb.String()), 0644); err != nil {
		return err
	}
	fmt.Printf("Results written to %s\n", tableFilePath)
	return nil
}

// writeCSVTable creates a .csv file summarizing the results.
func writeCSVTable(info ScriptInfo, results []result) error {
	filename := filepath.Base(info.Path)
	resultsFilenameForScript := strings.TrimSuffix(filename, ".ts")

	// Open the CSV file for writing
	tableFilePath := filepath.Join(projects.ResultsFolder, resultsFilenameForScript+".csv")
	file, err := os.Create(tableFilePath)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write headers
	headers := []string{"Project Path", "Status"}
	if info.Columns != nil {
		headers = append(headers, info.Columns...)
	} else {
		headers = append(headers, "Output")
	}
	if err := writer.Write(headers); err != nil {
		return err
	}

	for _, r := range results {
		lines := strings.Split(r.stdoutText, "\n")
		for _, line := range lines {
			values := strings.Split(line, ",")
			row := append([]string{r.projectPath, r.status}, values...)
			if err := writer.Write(row); err != nil {
				return err
			}
		}
	}

	fmt.Printf("Results written to %s\n", tableFilePath)
	return nil
}

func writeJSONOutput(scriptPath string, results []result) error {
	filename := filepath.Base(scriptPath)
	resultsFilenameForScript := strings.TrimSuffix(filename, ".ts")

	// ─── Transform []result → []map[string]any ───────────────────────────────
	var payload []map[string]any

	for _, r := range results {
		entry := map[string]any{
			"Project Path": r.projectPath,
			"Status":       r.status,
		}

		// Parse stdoutText as JSON; fall back to raw string on error.
		var out any
		if err := json.Unmarshal([]byte(r.stdoutText), &out); err == nil {
			entry["Output"] = out
		} else {
			entry["StdOut"] = r.stdoutText
		}

		if strings.TrimSpace(r.stderrText) != "" {
			entry["StdErr"] = r.stderrText
		}

		payload = append(payload, entry)
	}

	// ─── Encode & write to disk ──────────────────────────────────────────────
	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal results: %w", err)
	}

	tableFilePath := filepath.Join(projects.ResultsFolder, resultsFilenameForScript+".json")
	if err := os.WriteFile(tableFilePath, data, 0o644); err != nil {
		return fmt.Errorf("write results file: %w", err)
	}

	fmt.Printf("Results written to %s\n", tableFilePath)
	return nil
}
