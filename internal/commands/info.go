package commands

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/spf13/cobra"
	"github.com/wcatron/query-projects/internal/projects"
)

var InfoCmd = &cobra.Command{
	Use:   "info",
	Short: "Displays information about projects and available scripts",
	RunE: func(cmd *cobra.Command, args []string) error {
		debug, _ := cmd.Flags().GetBool("debug")
		return CMD_info(debug)
	},
}

// CMD_info lists the number of projects and available scripts
func CMD_info(debug bool) error {
	pj, err := projects.LoadProjects()
	if err != nil {
		return fmt.Errorf("failed to load projects: %w", err)
	}
	if debug {
		fmt.Printf("pj.RootDirectory = %s\n", pj.RootDirectory)
	}

	fmt.Printf("Number of projects: %d\n", len(pj.Projects))

	files, err := ioutil.ReadDir(projects.ScriptsFolder)
	if err != nil {
		return fmt.Errorf("failed to read scripts directory: %w", err)
	}

	fmt.Println("Available scripts:")
	scriptCount := 0
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".ts") {
			fmt.Printf("- %s\n", file.Name())
			scriptCount += 1
		}
	}
	if scriptCount == 0 {
		fmt.Print("No scripts found\n")
	}

	return nil
}
