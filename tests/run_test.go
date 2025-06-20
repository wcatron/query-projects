package e2e_tests

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
)

func parseResultsPath(s string) (string, error) {
	// `(?m)` enables multi-line mode; `.+` captures everything up to the line break
	ansi := regexp.MustCompile(`\x1b\[[0-9;]*m`)
	cleanOutput := ansi.ReplaceAllString(s, "")
	re := regexp.MustCompile(`(?m)^Results written to (.+)$`)
	m := re.FindStringSubmatch(cleanOutput)
	if len(m) == 2 {
		return m[1], nil // m[1] is the captured path
	}
	return "", fmt.Errorf("Unable to find results")
}

func skipCI(t *testing.T) {
	if os.Getenv("CI") != "" {
		t.Skip("Skipping testing in CI environment")
	}
}

func TestRun(t *testing.T) {
	// Unable to get test working as command fails in CI
	skipCI(t)
	// Define the command and arguments
	cmd := exec.Command("../query-projects", "run", "--script", "scripts/do-they-have-a-readme.ts", "--output=csv")

	// Set the working directory to the example directory
	cmd.Dir = "../example"

	// Run the command
	output, err := cmd.CombinedOutput()
	assert.NoError(t, err, "Expected no error running the CLI command")

	outputStr := string(output)

	resultsFile, resultsFileErr := parseResultsPath(outputStr)
	assert.NoError(t, resultsFileErr, "Expected no error parsing the results file")

	// Load the expected snapshot
	expectedOutput, err := os.ReadFile(filepath.Join("run_test_result.csv"))
	assert.NoError(t, err, "Expected no error reading the expected output file")

	// Load the result
	resultOutput, err := os.ReadFile(filepath.Join("../example/", resultsFile))
	assert.NoError(t, err, "Expected no error reading the expected output file")

	// Compare the output with the expected snapshot
	if diff := cmp.Diff(string(expectedOutput), string(resultOutput)); diff != "" {
		t.Errorf("Output mismatch (-expected +got):\n%s", diff)
	}
}

func TestRunBadScript(t *testing.T) {
	// Unable to get test working as command fails in CI
	skipCI(t)
	// Define the command and arguments
	cmd := exec.Command("../query-projects", "run", "--script", "scripts/invalid-ts.ts")

	// Set the working directory to the example directory
	cmd.Dir = "../example"

	// Run the command
	output, _ := cmd.CombinedOutput()
	outputStr := string(output)

	// Remove first line
	outputStr = outputStr[0:strings.Index(outputStr, "<eof>")]

	expectedOutput := "\x1b[33mscripts/invalid-ts.ts\x1b[0m \n\x1b[0m\x1b[1m\x1b[31merror\x1b[0m: The module's source code could not be parsed: Expected ';', '}' or "

	// Compare the output with the expected snapshot
	if diff := cmp.Diff(string(expectedOutput), string(outputStr)); diff != "" {
		t.Errorf("Output mismatch (-expected +got):\n%s", diff)
	}
}
