package commands

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

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
	projects, err := loadProjects()
	if err != nil {
		return err
	}

	for i, project := range projects.Projects {
		if project.Skip {
			continue
		}
		fmt.Printf("Fetching metadata for %s from GitHub...\n", project.Name)
		metadata, err := fetchGitHubMetadata(project.RepoURL)
		if err != nil {
			fmt.Printf("Error fetching metadata for %s: %v\n", project.Name, err)
			continue
		}
		fmt.Printf("Metadata for %s: %v\n", project.Name, metadata)
		// Update the project with the fetched metadata
		projects.Projects[i].Metadata = metadata
	}

	return saveProjects(projects)
}

func fetchGitHubMetadata(repoURL string) (map[string]interface{}, error) {
	apiURL := strings.Replace(repoURL, "https://github.com/", "https://api.github.com/repos/", 1)
	resp, err := http.Get(apiURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch metadata: %s %s", resp.Status, apiURL)
	}

	var metadata map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&metadata); err != nil {
		return nil, err
	}

	return metadata, nil
}
