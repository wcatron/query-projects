package e2e_tests

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
)

func TestInfo(t *testing.T) {
	// Define the command and arguments
	cmd := exec.Command("../query-projects", "info")

	// Set the working directory to the example directory
	cmd.Dir = "../example"

	// Run the command
	output, err := cmd.CombinedOutput()
	assert.NoError(t, err, "Expected no error running the CLI command")

	// Load the expected snapshot
	expectedOutput, err := os.ReadFile(filepath.Join("info.txt"))
	assert.NoError(t, err, "Expected no error reading the expected output file")

	// Compare the output with the expected snapshot
	if diff := cmp.Diff(string(expectedOutput), string(output)); diff != "" {
		t.Errorf("Output mismatch (-expected +got):\n%s", diff)
	}
}
