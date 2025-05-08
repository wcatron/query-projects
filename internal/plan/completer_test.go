package plan

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/wcatron/query-projects/internal/projects"
)

func setupTestRepo(t *testing.T) (string, func()) {
	// Create a temporary directory
	tmpDir, err := os.MkdirTemp("", "completer-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Create test files
	files := map[string]string{
		"package.json": `{"name": "test", "version": "1.0.0", "dependencies": {"test": "1.0.0"}}`,
		"config.xml":   `<config><setting>value</setting></config>`,
		"README.md":    "# Test",
	}

	for name, content := range files {
		if err := os.WriteFile(filepath.Join(tmpDir, name), []byte(content), 0644); err != nil {
			os.RemoveAll(tmpDir)
			t.Fatalf("Failed to create test file %s: %v", name, err)
		}
	}

	// Return cleanup function
	cleanup := func() {
		os.RemoveAll(tmpDir)
	}

	return tmpDir, cleanup
}

func TestLastFunction_Basic(t *testing.T) {
	preFunction, lastFunction, partialParams := LastFunction("if value(\"pack")
	if preFunction != "if " {
		t.Errorf("Expected pre function to be if not %s", preFunction)
	}
	if lastFunction != "value" {
		t.Errorf("Expected last function to be value not %s", lastFunction)
	}
	if len(partialParams) != 1 || partialParams[0] != "pack" {
		t.Errorf("Expected partial params contain pack not %v", partialParams)
	}
}

func TestLastFunction_NoParams(t *testing.T) {
	preFunction, lastFunction, partialParams := LastFunction("if value(")
	if preFunction != "if " {
		t.Errorf("Expected pre function to be if not %s", preFunction)
	}
	if lastFunction != "value" {
		t.Errorf("Expected last function to be value not %s", lastFunction)
	}
	if len(partialParams) != 0 {
		t.Errorf("Expected no partial params not %v", partialParams)
	}
}

func TestLastFunction_ClosedFunction(t *testing.T) {
	preFunction, lastFunction, _ := LastFunction("value(\"file\", \"field\") then ")
	if preFunction != "" {
		t.Errorf("Expected pre function to be empty not %s", preFunction)
	}
	if lastFunction != "" {
		t.Errorf("Expected last function to be empty not %s", lastFunction)
	}
}

func TestLastFunction_OpenFunction(t *testing.T) {
	preFunction, lastFunction, _ := LastFunction("value(\"file\", ")
	if preFunction != "" {
		t.Errorf("Expected pre function to be value( not %s", preFunction)
	}
	if lastFunction != "value" {
		t.Errorf("Expected last function to be file not %s", lastFunction)
	}
}

func TestCompleter_BasicCommands(t *testing.T) {
	completer := NewCompleter([]RepoContext{})
	completions := completer.Complete("")

	expected := []string{"run", "value", "step"}
	if len(completions) != len(expected) {
		t.Errorf("Expected %d completions, got %d", len(expected), len(completions))
	}

	for _, cmd := range expected {
		found := false
		for _, comp := range completions {
			if comp == cmd {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected completion %q not found", cmd)
		}
	}
}

func TestCompleter_EmptyRepos(t *testing.T) {
	completer := NewCompleter([]RepoContext{})
	completions := completer.Complete("value(")

	// Should return basic commands when no repos are available
	expected := []string{"run", "value", "step"}
	if len(completions) != len(expected) {
		t.Errorf("Expected %d completions, got %d", len(expected), len(completions))
	}
}

func TestCompleter_InvalidRepo(t *testing.T) {
	project := &projects.Project{
		Name: "invalid-repo",
		Path: "/nonexistent/path",
	}

	repo := RepoContext{
		Project: project,
	}

	completer := NewCompleter([]RepoContext{repo})
	completions := completer.Complete("value(")

	// Should return empty slice for invalid repo path
	if len(completions) != 0 {
		t.Errorf("Expected 0 completions for invalid repo, got %d", len(completions))
	}
}
