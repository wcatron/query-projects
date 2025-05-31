package scripts

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/wcatron/query-projects/internal/outputs"
	"github.com/wcatron/query-projects/internal/projects"
)

// GetScriptInfo executes a script with the --info flag and returns the parsed JSON output.
func GetScriptInfo(scriptPath string) (outputs.ScriptInfo, error) {
	cmd := exec.Command("deno", "run", "--allow-all", scriptPath, "--info")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return outputs.ScriptInfo{}, fmt.Errorf("failed to run script with --info: %w", err)
	}

	var info outputs.ScriptInfo
	if err := json.Unmarshal(output, &info); err != nil {
		return outputs.ScriptInfo{}, fmt.Errorf("failed to parse script info: %w", err)
	}

	info.Path = scriptPath

	return info, nil
}

// RunScriptForProject runs a TypeScript script (with Deno) in the specified project directory.
func RunScriptForProject(pj *projects.ProjectsJSON, scriptInfo outputs.ScriptInfo, projectPath string, args []string, print bool) (outputs.Result, error) {
	if print {
		fmt.Printf("Running %s with args [%s] for %s...\n", scriptInfo.Path, strings.Join(args, " "), projectPath)
	}

	var rootDirectory string
	if pj == nil {
		rootDirectory, _ = os.Getwd()
	} else {
		rootDirectory = pj.RootDirectory
	}

	scriptPath := filepath.Join(rootDirectory, scriptInfo.Path)

	cmd := exec.Command("deno", "run", "--allow-all", scriptPath, strings.Join(args, " "))
	cmd.Dir = filepath.Join(rootDirectory, projectPath)

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return outputs.Result{}, err
	}
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return outputs.Result{}, err
	}

	if err := cmd.Start(); err != nil {
		return outputs.Result{}, err
	}

	stdoutBytes, _ := io.ReadAll(stdoutPipe)
	stderrBytes, _ := io.ReadAll(stderrPipe)

	// Wait for command completion
	err = cmd.Wait()

	StdoutText := string(stdoutBytes)
	StderrText := string(stderrBytes)

	// Format CSV output if applicable
	if print {
		if scriptInfo.Output == "csv" && len(StdoutText) > 0 {
			fmt.Printf("[%s] CSV Output:\n", projectPath)
			fmt.Println(outputs.FormatOutput(StdoutText, scriptInfo.Columns))
		} else if len(StdoutText) > 0 {
			fmt.Printf("[%s] stdout:\n%s\n", projectPath, StdoutText)
		}
		if len(StderrText) > 0 {
			fmt.Printf("[%s] stderr:\n%s\n", projectPath, StderrText)
		}
	}

	var status string
	if err == nil {
		if print {
			fmt.Printf("Successfully ran %s for %s\n", scriptInfo.Path, projectPath)
		}
		status = "Success"
	} else {
		if exitErr, ok := err.(*exec.ExitError); ok {
			status = fmt.Sprintf("Failed (exit code %d)", exitErr.ExitCode())
			if print {
				fmt.Printf("Script %s failed for %s: %s\n", scriptInfo.Path, projectPath, exitErr.Error())
			}
		} else {
			status = "Error"
			if print {
				fmt.Printf("Error running script %s for %s: %v\n", scriptInfo.Path, projectPath, err)
			}
		}
	}

	return outputs.Result{
		ProjectPath: projectPath,
		Status:      status,
		StdoutText:  strings.TrimSpace(StdoutText),
		StderrText:  strings.TrimSpace(StderrText),
	}, nil
}
