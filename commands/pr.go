package commands

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var PRCmd = &cobra.Command{
	Use:   "pr <message>",
	Short: "Open a pull request for each repository (placeholder).",
	Args:  cobra.ArbitraryArgs,
	RunE: WrapWithMetrics(func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return errors.New("please provide a pull request message after 'pr'")
		}
		message := strings.Join(args, " ")
		return cmd_createPR(message)
	}),
}

// cmd_createPR is a placeholder function for opening PRs for each project.
func cmd_createPR(message string) error {
	projects, err := loadProjects()
	if err != nil {
		return err
	}
	for _, project := range projects.Projects {
		openPullRequest(project, message)
	}
	return saveProjects(&ProjectsJSON{Projects: updatedProjects})
}

// openPullRequest is a placeholder for opening a PR on GitHub or similar.
func openPullRequest(project Project, message string) {
	// TODO: Implement GitHub PR logic using GitHub API or `git`
	fmt.Printf("[PR Placeholder] Opening a pull request for '%s' with message: \"%s\"\n",
		project.Name, message)
}
