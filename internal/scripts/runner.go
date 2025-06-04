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

// prefixLines adds "<prefix> " in front of every line of s.
func prefixLines(s, prefix string) string {
	if s == "" {
		return ""
	}
	p := prefix + "\t"
	// Trim a trailing newline so we don’t end up with an empty last line.
	return p + strings.ReplaceAll(strings.TrimSuffix(s, "\n"), "\n", "\n"+p)
}

// RunScriptForProject runs a TypeScript script (with Deno) in the specified project directory.
func RunScriptForProject(pj *projects.ProjectsJSON, scriptInfo outputs.ScriptInfo, projectPath string, args []string, print bool) (outputs.Result, error) {
	if print {
		fmt.Printf("%s Running %s...\n", projects.ProjectPathFmt(projectPath), ScriptPathFmt(scriptInfo.Path))
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

	if print {
		if len(StdoutText) > 0 {
			fmt.Printf("%s\n", prefixLines(StdoutText, projects.ProjectPathFmt(projectPath)))
		}
		if len(StderrText) > 0 {
			fmt.Printf("%s\n", prefixLines(StderrText, projects.ProjectPathFmt(projectPath)))
		}
	}

	var status string
	if err == nil {
		status = "Success"
	} else {
		if exitErr, ok := err.(*exec.ExitError); ok {
			// TODO: Determine if this status is ever reached
			status = fmt.Sprintf("Failed (exit code %d)", exitErr.ExitCode())
			if print {
				fmt.Printf("%s Script %s failed %s\n", projects.ProjectPathFmt(projectPath), scriptInfo.Path, exitErr.Error())
			}
		} else {
			status = "Error"
			if print {
				fmt.Printf("%s Error running script %s error %v\n", projects.ProjectPathFmt(projectPath), scriptInfo.Path, err)
			}
		}
		if print {
			fmt.Printf("%s %s %s\n", projects.ProjectPathFmt(projectPath), status, ScriptPathFmt(scriptInfo.Path))
		}
	}

	return outputs.Result{
		ProjectPath: projectPath,
		Status:      status,
		StdoutText:  strings.TrimSpace(StdoutText),
		StderrText:  strings.TrimSpace(StderrText),
	}, nil
}
