package commands

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/wcatron/query-projects/internal/projects"
)

var PullCmd = &cobra.Command{
	Use:   "pull",
	Short: "Pull the latest changes for all repositories in projects.json",
	RunE: withMetrics(func(cmd *cobra.Command, args []string) error {
		topics, _ := cmd.Flags().GetStringSlice("topics")
		return CMD_pullRepos(topics)
	}),
}

// CMD_pullRepos pulls the latest changes for all repositories listed in projects.json.
func CMD_pullRepos(topics []string) error {
	projectsList, err := projects.LoadProjects()
	if err != nil {
		return err
	}
	filteredProjects := projects.FilterProjectsByTopics(projectsList.Projects, topics)
	for _, p := range filteredProjects {
		fmt.Printf("Pulling updates for %s in %s ...\n", p.Name, p.Path)
		if err := projects.CloneRepository(p.RepoURL, p.Path); err != nil {
			return fmt.Errorf("failed to pull repo %s: %w", p.Name, err)
		}
	}
	return nil
}
