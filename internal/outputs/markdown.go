package outputs

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/glamour"
	"github.com/wcatron/query-projects/internal/projects"
)

// createMarkdownString creates a markdown table string from results
func createMarkdownString(results []Result) strings.Builder {
	var sb strings.Builder
	headers := []string{"Project Path", "Status", "Output"}
	sb.WriteString("| " + strings.Join(headers, " | ") + " |\n")
	sb.WriteString("| " + strings.Repeat("--- | ", len(headers)) + "\n")

	for _, r := range results {
		lines := strings.Split(r.StdoutText, "\n")
		for _, line := range lines {
			row := []string{
				r.ProjectPath,
				r.Status,
				line,
			}
			sb.WriteString("| " + strings.Join(row, " | ") + " |\n")
		}
	}

	return sb
}

// PrintToConsole renders the results in markdown format to the console using Glamour
func PrintToConsole(results []Result) {
	var sb strings.Builder = createMarkdownString(results)

	// Render the markdown table using Glamour
	out, err := glamour.Render(sb.String(), "dark")
	if err != nil {
		fmt.Println("Error rendering markdown:", err)
		return
	}
	fmt.Print(out)
}

// WriteTable creates a .md table summarizing the results with their output
func WriteTable(scriptPath string, results []Result) error {
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
