package plan

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Completer provides command completion for the plan REPL
type Completer struct {
	repos []RepoContext
}

// NewCompleter creates a new completer with the given repos
func NewCompleter(repos []RepoContext) *Completer {
	return &Completer{repos: repos}
}

func LastFunction(line string) (string, string, []string) {
	// Find the last function call in the line
	lastParen := strings.LastIndex(line, "(")
	lastCloseParen := strings.LastIndex(line, ")")

	// If there's no lastParen
	if lastParen == -1 {
		return "", "", nil
	}

	// If there's a closing parenthesis after the last open parenthesis,
	// the function is closed and we should return empty strings
	if lastCloseParen > lastParen {
		return "", "", nil
	}

	// Find the start of the function name "run(" or "abc value("
	// Look backwards from the last open paren until a space or the start of the line
	start := lastParen
	for start >= 0 && line[start] != ' ' {
		start--
	}
	preFunction := line[0 : start+1]

	// Extract the function name and any partial parameters
	var functionName string
	var params []string

	if lastParen >= 0 {
		functionName = line[start+1 : lastParen]

		// Get any partial parameters after the paren
		if len(line) > lastParen+1 {
			// Remove the quotes and trailing comma from the partial parameters
			param := strings.Trim(line[lastParen+1:], "\", ")
			if param != "" {
				params = []string{param}
			}
		}
	}

	return preFunction, functionName, params
}

// Complete provides command completion for the plan REPL
func (c *Completer) Complete(line string) []string {
	if len(c.repos) == 0 {
		return []string{"run", "value", "step"}
	}

	// Use the first repo for file completion
	repo := c.repos[0]

	// Get the last function call and its context
	preFunction, functionName, params := LastFunction(line)
	if functionName == "" {
		return []string{"run", "value", "step"}
	}

	var completions []string

	switch functionName {
	case "value":
		if len(params) == 0 {
			// No parameters yet, suggest files
			files := getPossibleFiles(repo, "")
			for _, file := range files {
				completions = append(completions, fmt.Sprintf(`%s%s("%s", `, preFunction, functionName, file))
			}
		} else if len(params) == 1 {
			// First parameter is a file, suggest fields
			fields := getPossibleFields(repo, params[0], "")
			for _, field := range fields {
				completions = append(completions, fmt.Sprintf(`%s%s("%s", "%s")`, preFunction, functionName, params[0], field))
			}
		} else if len(params) == 2 {
			// Second parameter is a partial field, suggest matching fields
			fields := getPossibleFields(repo, params[0], params[1])
			for _, field := range fields {
				completions = append(completions, fmt.Sprintf(`%s%s("%s", "%s")`, preFunction, functionName, params[0], field))
			}
		}
	case "run":
		if len(params) == 0 {
			// No parameters yet, suggest files
			files := getPossibleFiles(repo, "")
			for _, file := range files {
				completions = append(completions, fmt.Sprintf(`%s"%s"`, preFunction, file))
			}
		} else if len(params) == 1 {
			// First parameter is a partial file, suggest matching files
			files := getPossibleFiles(repo, params[0])
			for _, file := range files {
				completions = append(completions, fmt.Sprintf(`%s"%s"`, preFunction, file))
			}
		}
	}

	return completions
}

func getPossibleFiles(repo RepoContext, partialPath string) []string {
	// partial path could be anything, need to remove the last part
	lastSlash := strings.LastIndex(partialPath, "/")
	if lastSlash == -1 {
		lastSlash = 0
	}
	basePath := partialPath[:lastSlash]

	files, err := os.ReadDir(filepath.Join(repo.Project.Path, basePath))
	if err != nil {
		return []string{}
	}

	var out []string
	for _, f := range files {
		if !f.IsDir() {
			ext := strings.ToLower(filepath.Ext(f.Name()))
			if ext == ".json" || ext == ".xml" {
				out = append(out, fmt.Sprintf("%s%s", basePath, f.Name()))
			}
		}
	}
	return out
}

func getPossibleFieldsXML(data []byte, partialField string) []string {
	var out []string
	var xmlData interface{}
	if err := xml.Unmarshal(data, &xmlData); err == nil {
		if root, ok := xmlData.(map[string]interface{}); ok {
			// Split the partial field into parts
			parts := strings.Split(partialField, ".")
			if len(parts) == 1 {
				// Single level field
				for field := range root {
					if strings.HasPrefix(field, partialField) {
						out = append(out, field)
					}
				}
			} else {
				// Nested field
				current := root
				// Navigate through the nested structure
				for i := 0; i < len(parts)-1; i++ {
					part := parts[i]
					// Check if the current level exists and is a map
					if val, ok := current[part]; ok {
						if nextMap, ok := val.(map[string]interface{}); ok {
							current = nextMap
						} else {
							// Not a map, can't go deeper
							return out
						}
					} else {
						// Part doesn't exist
						return out
					}
				}
				// Get completions for the last part
				lastPart := parts[len(parts)-1]
				for field := range current {
					if strings.HasPrefix(field, lastPart) {
						out = append(out, strings.Join(append(parts[:len(parts)-1], field), "."))
					}
				}
			}
		}
	}
}

func getPossibleFieldsJSON(data []byte, partialField string) []string {
	var out []string
	var jsonData map[string]interface{}
	if err := json.Unmarshal(data, &jsonData); err == nil {
		// Split the partial field into parts
		parts := strings.Split(partialField, ".")
		if len(parts) == 1 {
			// Single level field
			for field := range jsonData {
				if strings.HasPrefix(field, partialField) {
					out = append(out, field)
				}
			}
		} else {
			// Nested field
			current := jsonData
			// Navigate through the nested structure
			for i := 0; i < len(parts)-1; i++ {
				part := parts[i]
				// Check if the current level exists and is a map
				if val, ok := current[part]; ok {
					if nextMap, ok := val.(map[string]interface{}); ok {
						current = nextMap
					} else {
						// Not a map, can't go deeper
						return out
					}
				} else {
					// Part doesn't exist
					return out
				}
			}
			// Get completions for the last part
			lastPart := parts[len(parts)-1]
			for field := range current {
				if strings.HasPrefix(field, lastPart) {
					out = append(out, strings.Join(append(parts[:len(parts)-1], field), "."))
				}
			}
		}
	}
	return out
}

func getPossibleFields(repo RepoContext, filePath string, partialField string) []string {
	// Read the file
	data, err := os.ReadFile(filepath.Join(repo.Project.Path, filePath))
	if err != nil {
		return []string{}
	}

	// Handle based on file extension
	ext := strings.ToLower(filepath.Ext(filePath))

	if ext == ".json" {
		return getPossibleFieldsJSON(data, partialField)
	} else if ext == ".xml" {
		return getPossibleFieldsXML(data, partialField)
	}

	return []string{}
}
