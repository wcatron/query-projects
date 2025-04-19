package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

const (
	projectsFile  = "projects.json"
	scriptsFolder = "scripts"
	resultsFolder = "results"
	queryPrompt   = "QUERY_PROMPT.md"
)

// Project and ProjectsJSON store information about cloned repos
type Project struct {
	Name     string                 `json:"name"`
	Path     string                 `json:"path"`
	RepoURL  string                 `json:"repoUrl"`
	Topics   []string               `json:"topics"`
	Skip     bool                   `json:"skip,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

func filterProjectsByTopics(projects []Project, topics []string) []Project {
	var filteredProjects []Project
	for _, project := range projects {
		if project.Skip {
			continue
		}
		include := len(topics) == 0
		mustInclude := true

		for _, topic := range topics {
			if strings.HasPrefix(topic, "+") {
				// Must include projects with this topic
				if !contains(project.Topics, topic[1:]) {
					mustInclude = false
					break
				} else {
					include = true
				}
			} else if strings.HasPrefix(topic, "-") {
				// Exclude projects with this topic
				if contains(project.Topics, topic[1:]) {
					mustInclude = false
					break
				} else {
					include = true
				}
			} else {
				// Include if at least one topic matches
				if contains(project.Topics, topic) {
					include = true
				}
			}
		}

		fmt.Println("Total: ", project.Topics, topics, mustInclude, include)

		if mustInclude && include {
			filteredProjects = append(filteredProjects, project)
		}
	}
	return filteredProjects
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

type ProjectsJSON struct {
	Projects []Project `json:"projects"`
}

func loadProjects() (*ProjectsJSON, error) {
	data, err := os.ReadFile(projectsFile)
	if err != nil {
		// If the file doesn't exist, return an empty structure
		if os.IsNotExist(err) {
			return &ProjectsJSON{}, nil
		}
		return nil, err
	}

	var pj ProjectsJSON
	if err := json.Unmarshal(data, &pj); err != nil {
		return nil, err
	}
	return &pj, nil
}

// saveProjects writes the ProjectsJSON struct to projects.json.
func saveProjects(projects *ProjectsJSON) error {
	data, err := json.MarshalIndent(projects, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(projectsFile, data, 0644)
}

// cloneRepository either clones the repository if not present
// or pulls the latest changes if already cloned.
func cloneRepository(repoURL, projectPath string) error {
	_, err := os.Stat(projectPath)
	if err == nil {
		// Path exists, check if it's a .git repo
		gitDir := filepath.Join(projectPath, ".git")
		if _, errGit := os.Stat(gitDir); errGit == nil {
			// Git repo exists, do `git pull`
			fmt.Printf("Repository already cloned. Pulling latest changes from %s...\n", repoURL)
			cmd := exec.Command("git", "-C", projectPath, "pull")
			out, err := cmd.CombinedOutput()
			if err != nil {
				return fmt.Errorf("error pulling repository: %s\n%s", err, string(out))
			}
			fmt.Printf("Pulled latest changes:\n%s\n", string(out))
			return nil
		} else {
			return fmt.Errorf("directory exists but is not a Git repository: %s", projectPath)
		}
	} else if os.IsNotExist(err) {
		// Path doesn't exist, clone
		fmt.Printf("Cloning repository from %s to %s...\n", repoURL, projectPath)
		cmd := exec.Command("git", "clone", "--depth", "1", repoURL, projectPath)
		out, cloneErr := cmd.CombinedOutput()
		if cloneErr != nil {
			return fmt.Errorf("error cloning repository: %s\n%s", cloneErr, string(out))
		}
		fmt.Printf("Repository successfully cloned:\n%s\n", string(out))
		return nil
	} else {
		return err
	}
}

// extractTypeScriptCode finds the first ```ts or ```typescript code block in a string
// and returns its contents.
func ExtractTypeScriptCode(response string) string {
	codeBlockRegex := regexp.MustCompile("(?s)```(?:typescript|ts)?\\s*\\n(.+?)\\s*```")
	matches := codeBlockRegex.FindStringSubmatch(response)
	if len(matches) >= 2 {
		return strings.TrimSpace(matches[1])
	}
	return ""
}
