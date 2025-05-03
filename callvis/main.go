package main

import (
	"os"

	"github.com/wcatron/query-projects/internal/commands"
	"github.com/wcatron/query-projects/internal/outputs"
	"github.com/wcatron/query-projects/internal/projects"
)

func main() {
	// Add all subcommands
	commands.CMD_runScript("example.ts", []string{}, false, []string{})
	commands.CMD_addRepository("https://github.com/test/test")
	commands.CMD_info()
	commands.CMD_pullRepos([]string{})
	commands.CMD_syncRepos()
	commands.CMD_ask("test question")

	// Initialize command flags
	commands.RunCmdInit(commands.RunCmd)

	// Simulate command execution paths
	_ = commands.RunCmd.Execute()

	// Simulate script execution paths
	scriptInfo := outputs.ScriptInfo{
		Path:    "example.ts",
		Version: "1.0",
		Output:  "text",
	}
	project := projects.Project{
		Name:    "test",
		Path:    "test/path",
		RepoURL: "https://github.com/test/test",
	}

	// Simulate output handling
	results := []outputs.Result{
		{
			ProjectPath: "test/path",
			Status:      "Success",
			StdoutText:  "test output",
			StderrText:  "",
			Index:       0,
		},
	}
	_ = outputs.WriteTable(scriptInfo.Path, results)
	_ = outputs.WriteCSVTable(scriptInfo, results)
	_ = outputs.WriteJSONOutput(scriptInfo.Path, results)

	// Simulate project management
	_, _ = projects.LoadProjects()
	_ = projects.FilterProjectsByTopics([]projects.Project{project}, []string{"test"})

	// Prevent actual execution
	os.Exit(0)
}
