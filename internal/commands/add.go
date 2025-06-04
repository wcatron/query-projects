package commands

import (
	"fmt"
	"os"
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
		token, _ := cmd.Flags().GetString("githubToken")
		user, _ := cmd.Flags().GetString("githubUser")
		return CMD_addRepository(repoURL, token, user)
	},
}

func AddCmdInit(cmd *cobra.Command) {
	token := os.Getenv("GITHUB_TOKEN")
	user := os.Getenv("GITHUB_USER")
	cmd.PersistentFlags().StringP("githubToken", "", token, "Token to pull private github repositories defaults to GITHUB_TOKEN env.")
	cmd.PersistentFlags().StringP("githubUser", "", user, "User for token to pull private github repositories defaults to GITHUB_USER env.")
}

// CMD_addRepository clones the repo (if not present) and stores it in projects.json.
func CMD_addRepository(repoURL string, token string, user string) error {
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

	flags := make(map[string]string)
	if err := projects.CloneRepository(repoURL, projectPath, token, user, true, flags); err != nil {
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
