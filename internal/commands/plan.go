package commands

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/peterh/liner"
	"github.com/spf13/cobra"
	"github.com/wcatron/query-projects/internal/plan"
	"github.com/wcatron/query-projects/internal/projects"
	lua "github.com/yuin/gopher-lua"
)

var PlanCmd = &cobra.Command{
	Use:   "plan [script]",
	Short: "Plan changes across projects",
	Args:  cobra.MaximumNArgs(1),
	RunE: withMetrics(func(cmd *cobra.Command, args []string) error {
		topics, _ := cmd.Flags().GetStringSlice("topics")
		var script string
		if len(args) > 0 {
			script = args[0]
		}
		return CMD_plan(topics, script)
	}),
}

func CMD_plan(topics []string, script string) error {
	L := lua.NewState()
	defer L.Close()

	projectsList, err := projects.LoadProjects()
	if err != nil {
		return err
	}
	targets := projects.FilterProjectsByTopics(projectsList.Projects, topics)

	fmt.Printf("[%s] Loaded %d projects\n", "main", len(targets))

	// Create repo contexts from project list
	repos := make([]plan.RepoContext, len(targets))
	for i, project := range targets {
		repos[i] = newRepo(&project)
	}

	if script != "" {
		// Read and execute the script file
		content, err := os.ReadFile(script)
		if err != nil {
			return fmt.Errorf("failed to read script file: %v", err)
		}

		var wg sync.WaitGroup
		for _, repo := range repos {
			wg.Add(1)
			go func(r plan.RepoContext) {
				defer wg.Done()
				runCodeInRepo(r, string(content))
			}(repo)
		}
		wg.Wait()
		return nil
	}

	// Start REPL if no script provided
	line := liner.NewLiner()
	defer line.Close()

	line.SetCtrlCAborts(true)
	completer := plan.NewCompleter(repos)
	line.SetCompleter(completer.Complete)

	fmt.Println("Custom Lua REPL with Autocomplete. Type `exit` to quit.")

	for {
		input, err := line.Prompt("plan> ")
		if err != nil || input == "exit" {
			break
		}

		var wg sync.WaitGroup
		for _, repo := range repos {
			wg.Add(1)
			go func(r plan.RepoContext) {
				defer wg.Done()
				runCodeInRepo(r, input)
			}(repo)
		}
		wg.Wait()

		line.AppendHistory(input)
	}
	return nil
}

func newRepo(project *projects.Project) plan.RepoContext {
	vm := lua.NewState()
	// Register DSL
	vm.SetGlobal("run", vm.NewFunction(plan.RunFunc(project)))
	vm.SetGlobal("value", vm.NewFunction(plan.ValueFunc(project)))
	vm.SetGlobal("repoName", lua.LString(project.Name))
	return plan.RepoContext{Project: project, VM: vm}
}

func runCodeInRepo(repo plan.RepoContext, code string) {
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
