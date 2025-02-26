package commands

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var AddCmd = &cobra.Command{
	Use:   "add <repo-url>",
	Short: "Add a new repository and clone it locally.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		repoURL := args[0]
		return addRepository(repoURL)
	},
}

// addRepository clones the repo (if not present) and stores it in projects.json.
func addRepository(repoURL string) error {
	projects, err := loadProjects()
	if err != nil {
		return err
	}

	// Derive a project name from the repo URL
	projectName := strings.TrimSuffix(filepath.Base(repoURL), ".git")
	if projectName == "" {
		projectName = "unknown-project"
	}
	projectPath := filepath.Join("projects", projectName)

	if err := cloneRepository(repoURL, projectPath); err != nil {
		return err
	}

	// Add to our in-memory list of projects
	projects.Projects = append(projects.Projects, Project{
		Name:    projectName,
		Path:    projectPath,
		RepoURL: repoURL,
	})
	if err := saveProjects(projects); err != nil {
		return err
	}

	fmt.Printf("Added %s to %s.\n", projectName, projectsFile)
	return nil
}
