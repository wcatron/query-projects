package outputs

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/wcatron/query-projects/internal/projects"
)

// FormatOutput formats CSV output based on column headers
func FormatOutput(csvText string, columns []string) string {
	var sb strings.Builder
	sb.WriteString(strings.Join(columns, ",") + "\n")
	sb.WriteString(csvText)
	return sb.String()
}

// WriteCSVTable creates a .csv file summarizing the results
func WriteCSVTable(info ScriptInfo, results []Result) error {
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
		lines := strings.Split(r.StdoutText, "\n")
		for _, line := range lines {
			values := strings.Split(line, ",")
			row := append([]string{r.ProjectPath, r.Status}, values...)
			if err := writer.Write(row); err != nil {
				return err
			}
		}
	}

	fmt.Printf("Results written to %s\n", tableFilePath)
	return nil
}
