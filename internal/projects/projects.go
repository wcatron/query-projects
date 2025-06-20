package projects

import (
	"encoding/json"
	"fmt"
	"maps"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strings"
)

const (
	ProjectsFile  = "projects.json"
	ScriptsFolder = "scripts"
	ResultsFolder = "results"
	QueryPrompt   = "QUERY_PROMPT.md"
)

// Project and ProjectsJSON store information about cloned repos
type Project struct {
	Name     string            `json:"name"`
	Path     string            `json:"path"`
	RepoURL  string            `json:"repoUrl"`
	Topics   []string          `json:"topics"`
	Skip     bool              `json:"skip,omitempty"`
	Metadata interface{}       `json:"metadata,omitempty"`
	Git      map[string]string `json:"git,omitempty"`
}

// This is the defineition
func FilterProjectsByTopics(projects []Project, topics []string) []Project {
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
	RootDirectory string    `json:"-"`
	Projects      []Project `json:"projects"`
}

func findFileInParents(startDir, fileName string) (string, error) {
	currentDir := startDir

	for {
		filePath := filepath.Join(currentDir, fileName)
		_, err := os.Stat(filePath)
		if err == nil {
			return filePath, nil // File found
		} else if !os.IsNotExist(err) {
			return "", err // Some other error occurred
		}

		parentDir := filepath.Dir(currentDir)
		if parentDir == currentDir {
			return "", fmt.Errorf("file %s not found in parent directories", fileName)
		}
		currentDir = parentDir
	}
}

func findProjectsDir() (string, error) {
	cwd, _ := os.Getwd()
	// Use QUERY_PROJECTS_DIRECTORY when not storing all projects under the meta project
	// TODO: Fully bake out that feature.
	dir, hasDir := os.LookupEnv("QUERY_PROJECTS_DIRECTORY")
	if hasDir {
		return dir, nil
	} else {
		foundFile, err := findFileInParents(cwd, ProjectsFile)
		if err != nil {
			return "", err
		}
		return filepath.Dir(foundFile), nil
	}
}

func InProject(pj *ProjectsJSON) *Project {
	cwd, _ := os.Getwd()
	if cwd != pj.RootDirectory {
		index := slices.IndexFunc(pj.Projects, func(project Project) bool {
			return filepath.Join(pj.RootDirectory, project.Path) == cwd
		})
		return &pj.Projects[index]
	} else {
		return nil
	}
}

func LoadProjects() (*ProjectsJSON, error) {
	projectsDir, err := findProjectsDir()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(filepath.Join(projectsDir, ProjectsFile))
	if err != nil {
		// If the file doesn't exist, return an empty structure
		// TODO: Figure out why??
		/* if os.IsNotExist(err) {
			return &ProjectsJSON{}, nil
		}*/
		return nil, err
	}

	var pj ProjectsJSON
	if err := json.Unmarshal(data, &pj); err != nil {
		return nil, err
	}
	pj.RootDirectory = projectsDir
	return &pj, nil
}

// SaveProjects writes the ProjectsJSON struct to projects.json.
func SaveProjects(projects *ProjectsJSON) error {
	data, err := json.MarshalIndent(projects, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(ProjectsFile, data, 0644)
}

func FlagsToArgs(flags map[string]string) []string {
	if len(flags) == 0 {
		return nil
	}

	// Gather keys so we can sort; maps are random‑order by default.
	keys := make([]string, 0, len(flags))
	for k := range flags {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Build the argument slice.
	args := make([]string, 0, len(flags)*2)
	for _, k := range keys {
		v := flags[k]
		args = append(args, "--"+k)
		if v != "" { // handle boolean flags where value may be empty
			args = append(args, v)
		}
	}
	return args
}

var DEFAULT_FLAGS = map[string]string{}

// checkMatchingToken returns true if the personal‑access token embedded in the
// repo's origin URL does NOT match the token argument.
//
// The origin URL can look like either of these:
//
//	https://<PAT>@github.com/owner/repo.git
//	https://x-access-token:<PAT>@github.com/owner/repo.git
//
// If the URL has no embedded credentials (common when a credential helper
// is used) or the command fails, the function treats that as “token mismatch”.
func checkMatchingToken(repoURL string, projectPath string, githubToken string, githubUser string) bool {
	out, err := exec.Command("git", "-C", projectPath, "remote", "get-url", "origin").Output()
	if err != nil {
		fmt.Printf("Error determining checkMatchingToken for %s\n", projectPath)
		return true
	}
	origin := strings.TrimSpace(string(out))

	u, err := url.Parse(origin)
	if err != nil || u.User == nil {
		// Token not being used at all
		fmt.Printf("No token being used for project %s\n", projectPath)
		return false
	}

	var pat string
	if pwd, hasPwd := u.User.Password(); hasPwd {
		pat = pwd
	} else {
		pat = u.User.Username()
	}

	return pat != githubToken
}

// updateRemoteToken sets the repo's origin URL to
//
//	https://<TOKEN>@github.com/owner/repo.git
//
// so future git fetch/pull operations authenticate automatically.
//
// If the URL isn't https://… or the token is empty, the function does nothing.
func updateRemoteToken(repoURL, projectPath, githubToken string, githubUser string) error {
	if githubToken == "" || githubUser == "" {
		return nil // nothing to insert
	}

	// 1) Build the authenticated URL.
	u, err := url.Parse(repoURL)
	if err != nil || u.Scheme != "https" {
		return err // unsupported or malformed
	}
	u.User = url.UserPassword(githubUser, githubToken) // token@github.com/…

	authURL := u.String()

	// 2) Run: git -C <path> remote set-url origin <authURL>
	cmd := exec.Command(
		"git", "-C", projectPath, "remote", "set-url", "origin", authURL,
	)

	// redact token in log
	cleanURL := strings.Replace(authURL, githubToken, "********", 1)
	fmt.Printf("Updating origin URL to %s\n", cleanURL)

	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("remote set-url failed: %w\n%s", err, out)
	}

	return nil
}

// CloneRepository either clones the repository if not present
// or pulls the latest changes if already cloned.
func CloneRepository(repoURL string, projectPath string, githubToken string, githubUser string, updateToken bool, flags map[string]string) error {
	_, err := os.Stat(projectPath)

	finalFlags := maps.Clone(DEFAULT_FLAGS)
	if flags != nil {
		maps.Copy(finalFlags, flags)
	}
	flagArgs := FlagsToArgs(finalFlags)

	if err == nil {
		// Path exists, check if it's a .git repo
		gitDir := filepath.Join(projectPath, ".git")

		if _, errGit := os.Stat(gitDir); errGit == nil {
			if updateToken {
				if checkMatchingToken(repoURL, projectPath, githubToken, githubUser) {
					fmt.Printf("%s Token is already up to date.\n", ProjectPathFmt(projectPath))
				} else {
					updateRemoteToken(repoURL, projectPath, githubToken, githubUser)
				}
			}
			// Git repo exists, do `git pull`
			fmt.Printf("%s Repo cloned. Pulling latest changes from %s\n", ProjectPathFmt(projectPath), repoURL)
			args := append([]string{"-C", projectPath, "pull"}, flagArgs...)
			fmt.Printf("git %s\n", strings.Join(args, " "))
			cmd := exec.Command("git", args...)
			out, err := cmd.CombinedOutput()
			if err != nil {
				return fmt.Errorf("error pulling repository: %s\n%s", err, string(out))
			}
			if len(out) > 0 {
				fmt.Printf("%s %s\n", ProjectPathFmt(projectPath), string(out))
			}
			return nil
		} else {
			return fmt.Errorf("directory exists but is not a Git repository: %s", projectPath)
		}
	} else if os.IsNotExist(err) {
		fmt.Printf("%s Cloning repository from %s\n", ProjectPathFmt(projectPath), repoURL)

		authURL := repoURL
		if githubToken != "" {
			if u, err := url.Parse(repoURL); err == nil && u.Scheme == "https" {
				u.User = url.UserPassword(githubUser, githubToken)
				authURL = u.String()
			}
		}

		args := append([]string{"clone"}, flagArgs...)
		args = append(args, authURL, projectPath)
		cleanArgs := strings.Replace(strings.Join(args, " "), githubToken, "********", 1)
		fmt.Printf("%s git %s\n", ProjectPathFmt(projectPath), cleanArgs)
		cmd := exec.Command("git", args...)
		out, cloneErr := cmd.CombinedOutput()
		if cloneErr != nil {
			fmt.Printf("%s\n\033[31m%s\n\033[0m", ProjectPathFmt(projectPath), string(out))
			return fmt.Errorf("error cloning repository: %s", cloneErr)
		}
		fmt.Printf("%s Repository successfully cloned:\n%s\n", projectPath, string(out))
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

func ProjectPathFmt(projectPath string) string {
	return fmt.Sprintf("\033[33m%s\033[0m", projectPath)
}
