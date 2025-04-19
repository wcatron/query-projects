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
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: os.Getenv("GITHUB_TOKEN")},
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
		metadata, err := fetchGitHubMetadata(ctx, client, project.RepoURL)
		if err != nil {
			fmt.Printf("Error fetching metadata for %s: %v\n", project.Name, err)
			continue
		}
		fmt.Printf("Metadata for %s\n", project.Name)
		// Update the project with the fetched metadata
		projects.Projects[i].Metadata = metadata
	}

	return saveProjects(projects)
}

func fetchGitHubMetadata(ctx context.Context, client *github.Client, repoURL string) (map[string]interface{}, error) {
	ownerRepo := strings.TrimPrefix(strings.TrimSuffix(repoURL, ".git"), "https://github.com/")
	parts := strings.Split(ownerRepo, "/")
	if len(parts) != 2 {
		return nil, errors.New("invalid GitHub repository URL")
	}

	repo, _, err := client.Repositories.Get(ctx, parts[0], parts[1])
	if err != nil {
		return nil, err
	}

	metadata := map[string]interface{}{
		"full_name":      repo.GetFullName(),
		"description":    repo.GetDescription(),
		"stars":          repo.GetStargazersCount(),
		"forks":          repo.GetForksCount(),
		"open_issues":    repo.GetOpenIssuesCount(),
		"created_at":     repo.GetCreatedAt(),
		"updated_at":     repo.GetUpdatedAt(),
		"pushed_at":      repo.GetPushedAt(),
		"default_branch": repo.GetDefaultBranch(),
	}

	return metadata, nil
}
