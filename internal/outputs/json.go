package outputs

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/wcatron/query-projects/internal/projects"
)

// WriteJSONOutput creates a .json file summarizing the results
func WriteJSONOutput(scriptPath string, results []Result) error {
	filename := filepath.Base(scriptPath)
	resultsFilenameForScript := strings.TrimSuffix(filename, ".ts")

	// Transform []Result → []map[string]any
	var payload []map[string]any

	for _, r := range results {
		entry := map[string]any{
			"Project Path": r.ProjectPath,
			"Status":       r.Status,
		}

		// Parse StdoutText as JSON; fall back to raw string on error.
		var out any
		if err := json.Unmarshal([]byte(r.StdoutText), &out); err == nil {
			entry["Output"] = out
		} else {
			entry["StdOut"] = r.StdoutText
		}

		if strings.TrimSpace(r.StderrText) != "" {
			entry["StdErr"] = r.StderrText
		}

		payload = append(payload, entry)
	}

	// Encode & write to disk
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
