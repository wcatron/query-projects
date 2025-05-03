package commands

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/wcatron/query-projects/internal/projects"
)

var AddCmd = &cobra.Command{
	Use:   "add <repo-url>",
	Short: "Add a new repository and clone it locally.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		repoURL := args[0]
		return cmd_addRepository(repoURL)
	},
}

// cmd_addRepository clones the repo (if not present) and stores it in projects.json.
func cmd_addRepository(repoURL string) error {
	projectsList, err := projects.LoadProjects()
	if err != nil {
		return err
	}

	// Derive a project name from the repo URL
	projectName := strings.TrimSuffix(filepath.Base(repoURL), ".git")
	if projectName == "" {
		projectName = "unknown-project"
	}
	projectPath := filepath.Join("projects", projectName)

	if err := projects.CloneRepository(repoURL, projectPath); err != nil {
		return err
	}

	// Add to our in-memory list of projects
	projectsList.Projects = append(projectsList.Projects, projects.Project{
		Name:    projectName,
		Path:    projectPath,
		RepoURL: repoURL,
	})
	if err := projects.SaveProjects(projectsList); err != nil {
		return err
	}

	fmt.Printf("Added %s to %s.\n", projectName, projects.ProjectsFile)
	return nil
}
