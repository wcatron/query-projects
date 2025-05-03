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
		return cmd_info()
	},
}

// cmd_info lists the number of projects and available scripts
func cmd_info() error {
	projectsList, err := projects.LoadProjects()
	if err != nil {
		return fmt.Errorf("failed to load projects: %w", err)
	}

	fmt.Printf("Number of projects: %d\n", len(projectsList.Projects))

	files, err := ioutil.ReadDir(projects.ScriptsFolder)
	if err != nil {
		return fmt.Errorf("failed to read scripts directory: %w", err)
	}

	fmt.Println("Available scripts:")
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".ts") {
			fmt.Printf("- %s\n", file.Name())
		}
	}

	return nil
}
