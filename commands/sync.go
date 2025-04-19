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
		fmt.Printf("Fetching metadata for %s from GitHub...\n", project.Name)
		repo, err := fetchGitHubMetadata(ctx, client, project.RepoURL)
		if err != nil {
			fmt.Printf("Error fetching metadata for %s: %v\n", project.Name, err)
			continue
		}
		fmt.Printf("Metadata for %s: %+v\n", project.Name, repo)
		// Update the project with the fetched metadata
		projects.Projects[i].Metadata = repo
	}

	return saveProjects(projects)
}

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
