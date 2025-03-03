package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

var PullCmd = &cobra.Command{
	Use:   "pull",
	Short: "Pull the latest changes for all repositories in projects.json",
	RunE: WrapWithMetrics(func(cmd *cobra.Command, args []string) error {
		return pullAllRepositories()
	}),
}

// pullAllRepositories pulls the latest changes for all repositories listed in projects.json.
func pullAllRepositories() error {
	projects, err := loadProjects()
	if err != nil {
		return err
	}
	for _, p := range projects.Projects {
		fmt.Printf("Pulling updates for %s in %s ...\n", p.Name, p.Path)
		if err := cloneRepository(p.RepoURL, p.Path); err != nil {
			return fmt.Errorf("failed to pull repo %s: %w", p.Name, err)
		}
	}
	return nil
}
