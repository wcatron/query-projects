package commands

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/wcatron/query-projects/internal/projects"
)

var PullCmd = &cobra.Command{
	Use:   "pull",
	Short: "Pull the latest changes for all repositories in projects.json",
	RunE: withMetrics(func(cmd *cobra.Command, args []string) error {
		topics, _ := cmd.Flags().GetStringSlice("topics")
		githubToken, _ := cmd.Flags().GetString("githubToken")
		githubUser, _ := cmd.Flags().GetString("githubUser")
		githubUpdateToken, _ := cmd.Flags().GetBool("githubUpdateToken")

		return CMD_pullRepos(topics, githubToken, githubUser, githubUpdateToken)
	}),
}

func PullCmdInit(cmd *cobra.Command) {
	githubToken := os.Getenv("GITHUB_TOKEN")
	githubUser := os.Getenv("GITHUB_USER")
	cmd.PersistentFlags().String("githubToken", githubToken, "Token to pull private github repositories defaults to GITHUB_TOKEN env.")
	cmd.PersistentFlags().String("githubUser", githubUser, "User for token to pull private github repositories defaults to GITHUB_USER env.")
	cmd.PersistentFlags().Bool("githubUpdateToken", false, "Run script to update token")
}

// CMD_pullRepos pulls or clones every repo whose topic matches.
// It keeps going even if some repos fail and returns a joined error list.
func CMD_pullRepos(topics []string, githubToken string, githubUser string, githubUpdateToken bool) error {
	projectsList, err := projects.LoadProjects()
	if err != nil {
		return err
	}

	filtered := projects.FilterProjectsByTopics(projectsList.Projects, topics)

	var errs []error

	for _, p := range filtered {
		if err := projects.CloneRepository(p.RepoURL, p.Path, githubToken, githubUser, githubUpdateToken, p.Git); err != nil {
			wrap := fmt.Errorf("%s %w", projects.ProjectPathFmt(p.Path), err)
			errs = append(errs, wrap)
		}
	}

	if len(errs) > 0 {
		fmt.Printf("\n%s\n", errors.Join(errs...))
	}
	return nil
}
