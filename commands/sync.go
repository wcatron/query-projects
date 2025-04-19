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
	Use:   "sync <repository>",
	Short: "Sync project metadata from a specified code repository.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		repoType := strings.ToLower(args[0])
		switch repoType {
		case "github":
			return syncFromGitHub()
		case "azure":
			fmt.Println("Azure sync is not implemented yet.")
		case "bitbucket":
			fmt.Println("Bitbucket sync is not implemented yet.")
		default:
			return errors.New("unsupported repository type")
		}
		return nil
	},
}

/*
syncFromGitHub fetches metadata for all projects listed in the projects.json file from GitHub.
It updates the project metadata with topics and archive status. The function requires the GITHUB_TOKEN
environment variable to be set for authentication.
*/
func syncFromGitHub() error {
	ctx := context.Background()
	githubToken := os.Getenv("GITHUB_TOKEN")
	if githubToken == "" {
		return errors.New("GITHUB_TOKEN environment variable is not set")
	}

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: githubToken},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	projects, err := loadProjects()
	if err != nil {
		return err
	}

	for i, project := range projects.Projects {
		if project.Skip {
			continue
		}
		fmt.Printf("Fetching metadata for project '%s' from GitHub...\n", project.Name)
		repo, err := fetchGitHubMetadata(ctx, client, project.RepoURL)
		if err != nil {
			fmt.Printf("Error fetching metadata for project '%s': %v\n", project.Name, err)
			continue
		}
		fmt.Printf("Successfully fetched metadata for project '%s'.\n", project.Name)
		// Pull abstracted fields into the top level
		projects.Projects[i].Topics = repo.Topics
		projects.Projects[i].Skip = projects.Projects[i].Skip || repo.GetArchived()
		projects.Projects[i].Metadata = repo
	}

	return saveProjects(projects)
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
