package commands

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/google/go-github/v71/github"
	"golang.org/x/oauth2"

	"github.com/spf13/cobra"
)

/*
SyncCmd is a Cobra command that synchronizes project metadata from a specified code repository.
Currently, it supports syncing from GitHub. The command requires a single argument specifying
the repository type (e.g., "github"). It uses the GITHUB_TOKEN environment variable for authentication.
*/
var SyncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync project metadata from all configured code repositories.",
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd_syncRepos()
	},
}

/*
syncFromGitHub fetches metadata for all projects listed in the projects.json file from GitHub.
It updates the project metadata with topics and archive status. The function requires the GITHUB_TOKEN
environment variable to be set for authentication.
*/
func syncFromGitHubProject(project Project) (Project, error) {
	ctx := context.Background()
	githubToken := os.Getenv("GITHUB_TOKEN")
	if githubToken == "" {
		return project, errors.New("GITHUB_TOKEN environment variable is not set")
	}

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: githubToken},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	fmt.Printf("Fetching metadata for project '%s' from GitHub...\n", project.Name)
	repo, err := fetchGitHubMetadata(ctx, client, project.RepoURL)
	if err != nil {
		return project, fmt.Errorf("error fetching metadata for project '%s': %w", project.Name, err)
	}
	fmt.Printf("Successfully fetched metadata for project '%s'.\n", project.Name)
	// Pull abstracted fields into the top level
	project.Topics = repo.Topics
	project.Skip = project.Skip || repo.GetArchived()
	project.Metadata = repo

	return project, nil
}

/*
fetchGitHubMetadata retrieves metadata for a given GitHub repository URL. It extracts the owner and
repository name from the URL and uses the GitHub API to fetch repository details.
*/
func fetchGitHubMetadata(ctx context.Context, client *github.Client, repoURL string) (*github.Repository, error) {
	ownerRepo := strings.TrimPrefix(strings.TrimSuffix(repoURL, ".git"), "https://github.com/")
	parts := strings.Split(ownerRepo, "/")
	if len(parts) != 2 {
		return nil, errors.New("invalid GitHub repository URL")
	}

	repo, _, err := client.Repositories.Get(ctx, parts[0], parts[1])
	if err != nil {
		return nil, err
	}

	return repo, nil
}
func cmd_syncRepos() error {
	projects, err := loadProjects()
	if err != nil {
		return err
	}

	for index, project := range projects.Projects {
		if project.Skip {
			continue
		}

		if strings.Contains(project.RepoURL, "github.com") {
			fmt.Printf("Syncing GitHub project '%s'...\n", project.Name)
			updatedProject, err := syncFromGitHubProject(project)
			if err != nil {
				fmt.Printf("Error syncing GitHub project '%s': %v\n", project.Name, err)
			}
			projects.Projects[index] = updatedProject
		} else if strings.Contains(project.RepoURL, "bitbucket.org") {
			fmt.Printf("Bitbucket sync for project '%s' is not implemented yet.\n", project.Name)
		} else if strings.Contains(project.RepoURL, "azure.com") {
			fmt.Printf("Azure sync for project '%s' is not implemented yet.\n", project.Name)
		} else {
			fmt.Printf("Unsupported repository type for project '%s'.\n", project.Name)
		}
	}
	return saveProjects(projects)
}
