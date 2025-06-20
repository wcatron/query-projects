package outputs

import (
	"fmt"
	"os"
	"path/filepath"
)

// Result represents the output of running a script on a project
type Result struct {
	ProjectPath string
	Status      string
	StdoutText  string
	StderrText  string
	Index       int
}

// ScriptInfo represents information about a script
type ScriptInfo struct {
	Path    string   `json:"path"`
	Version string   `json:"version"`
	Output  string   `json:"output"`
	Columns []string `json:"columns"`
}

func CleanPath(absPath string) string {
	cwd, err := os.Getwd()
	if err != nil {
		// Fallback to original if CWD can't be determined
		return absPath
	}

	rel, err := filepath.Rel(cwd, absPath)
	if err != nil {
		// Fallback to original if paths are on different volumes, etc.
		return absPath
	}

	return fmt.Sprintf("\033[2m%s\033[0m", filepath.Clean(rel))
}
