package commands

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestRunCmd(t *testing.T) {
	cmd := RunCmd
	if cmd.Use != "run [scriptName]" {
		t.Errorf("expected use 'run [scriptName]', got %s", cmd.Use)
	}
	if cmd.Short != "Run a script (or all .ts scripts) across all projects" {
		t.Errorf("expected short description 'Run a script (or all .ts scripts) across all projects', got %s", cmd.Short)
	}
	if cmd.Args != cobra.MaximumNArgs(1) {
		t.Errorf("expected maximum args 1, got %v", cmd.Args)
	}
}
