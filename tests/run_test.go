package e2e_tests

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
)

func parseResultsPath(s string) (string, error) {
	// `(?m)` enables multi-line mode; `.+` captures everything up to the line break
	re := regexp.MustCompile(`(?m)^Results written to (.+)$`)
	m := re.FindStringSubmatch(s)
	if len(m) == 2 {
		return m[1], nil // m[1] is the captured path
	}
	return "", fmt.Errorf("Unable to find results")
}

func TestRun(t *testing.T) {
	// Define the command and arguments
	cmd := exec.Command("../query-projects", "run", "scripts/return-the-path-to-every-markdown-file-in-the-project.ts", "--output=csv")

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
