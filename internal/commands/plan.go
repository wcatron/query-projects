package commands

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/peterh/liner"
	"github.com/spf13/cobra"
	"github.com/wcatron/query-projects/internal/projects"
	lua "github.com/yuin/gopher-lua"
)

// RepoContext represents an execution context for one repo
type RepoContext struct {
	Project *projects.Project
	VM      *lua.LState
}

var PlanCmd = &cobra.Command{
	Use:   "plan",
	Short: "Plan changes across projects",
	RunE: withMetrics(func(cmd *cobra.Command, args []string) error {
		topics, _ := cmd.Flags().GetStringSlice("topics")
		return CMD_plan(topics)
	}),
}

func CMD_plan(topics []string) error {
	L := lua.NewState()
	defer L.Close()

	line := liner.NewLiner()
	defer line.Close()

	line.SetCtrlCAborts(true)
	line.SetCompleter(func(line string) []string {
		return completer(line)
	})

	projectsList, err := projects.LoadProjects()
	if err != nil {
		return err
	}
	targets := projects.FilterProjectsByTopics(projectsList.Projects, topics)

	fmt.Printf("[%s] Loaded %d projects\n", "main", len(targets))

	// Create repo contexts from project list
	repos := make([]RepoContext, len(targets))
	for i, project := range targets {
		repos[i] = newRepo(&project)
	}

	fmt.Println("Custom Lua REPL with Autocomplete. Type `exit` to quit.")

	for {
		input, err := line.Prompt("plan> ")
		if err != nil || input == "exit" {
			break
		}

		var wg sync.WaitGroup
		for _, repo := range repos {
			wg.Add(1)
			go func(r RepoContext) {
				defer wg.Done()
				runCodeInRepo(r, input)
			}(repo)
		}
		wg.Wait()

		line.AppendHistory(input)
	}
	return nil
}

func newRepo(project *projects.Project) RepoContext {
	vm := lua.NewState()
	// Register DSL
	vm.SetGlobal("run", vm.NewFunction(runFunc(project)))
	vm.SetGlobal("value", vm.NewFunction(valueFunc(project)))
	vm.SetGlobal("repoName", lua.LString(project.Name))
	return RepoContext{Project: project, VM: vm}
}

func runCodeInRepo(repo RepoContext, code string) {
	var output strings.Builder

	// Replace `print` in Lua
	repo.VM.SetGlobal("print", repo.VM.NewFunction(makePrintFunc(repo.Project.Name, &output)))

	err := repo.VM.DoString(code)

	// Always print output
	if output.Len() > 0 {
		fmt.Print(output.String())
	}

	if err != nil {
		fmt.Printf("[%s] Error: %v\n", repo.Project.Name, err)
	}
}

func makePrintFunc(repoName string, logger *strings.Builder) lua.LGFunction {
	return func(L *lua.LState) int {
		top := L.GetTop()
		args := make([]string, top)
		for i := 1; i <= top; i++ {
			args[i-1] = L.ToString(i)
		}
		msg := fmt.Sprintf("[%s] %s", repoName, strings.Join(args, "\t"))
		logger.WriteString(msg + "\n")
		return 0
	}
}

func runFunc(project *projects.Project) func(L *lua.LState) int {
	return func(L *lua.LState) int {
		script := L.CheckString(1)
		fmt.Printf("[%s] Running script: %s (in %s)\n", project.Name, script, project.Path)
		scriptInfo, err := getScriptInfo(script)
		if err != nil {
			L.RaiseError("failed to get script info: %v", err)
			return 0
		}
		output, err := runScriptForProject(scriptInfo, project.Path, false)
		if err != nil {
			L.RaiseError("failed to run script: %v", err)
			return 0
		}
		L.Push(lua.LString(output.StdoutText))
		return 1
	}
}

func valueFunc(project *projects.Project) func(L *lua.LState) int {
	return func(L *lua.LState) int {
		file := L.CheckString(1)
		field := L.CheckString(2)

		data, err := os.ReadFile(filepath.Join(project.Path, file))
		if err != nil {
			L.RaiseError("failed to read file %s: %v", file, err)
			return 0
		}

		var value string
		ext := strings.ToLower(filepath.Ext(file))
		switch ext {
		case ".json":
			var jsonData map[string]interface{}
			if err := json.Unmarshal(data, &jsonData); err != nil {
				L.RaiseError("failed to parse JSON file %s: %v", file, err)
				return 0
			}
			if val, ok := jsonData[field]; ok {
				value = fmt.Sprint(val)
			}
		case ".xml":
			var xmlData map[string]interface{}
			decoder := xml.NewDecoder(bytes.NewReader(data))
			if err := decoder.Decode(&xmlData); err != nil {
				L.RaiseError("failed to parse XML file %s: %v", file, err)
				return 0
			}
			if val, ok := xmlData[field]; ok {
				value = fmt.Sprint(val)
			}
		default:
			L.RaiseError("unsupported file type: %s", ext)
			return 0
		}

		if value == "" {
			// L.RaiseError("field %s not found in file %s", field, file)
			L.Push(lua.LString(""))
			return 1
		}

		L.Push(lua.LString(value))
		return 1
	}
}

func completer(line string) []string {

	// TODO: Actual completion

	// Simulate per-repo context
	currentRepo := "repo-a"
	files := map[string][]string{
		"repo-a": {"package.json", "tsconfig.json"},
		"repo-b": {"custom.yaml"},
	}
	fields := map[string][]string{
		"package.json": {"version", "name"},
	}

	if strings.HasPrefix(line, "value(") {
		var out []string
		for _, f := range files[currentRepo] {
			for _, field := range fields[f] {
				out = append(out, fmt.Sprintf(`%s"%s", "%s"`, line, f, field))
			}
		}
		return out
	}
	return []string{"run", "value", "step"}
}
