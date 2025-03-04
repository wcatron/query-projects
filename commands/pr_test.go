package commands

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestPRCmd(t *testing.T) {
	cmd := PRCmd
	if cmd.Use != "pr <message>" {
		t.Errorf("expected use 'pr <message>', got %s", cmd.Use)
	}
	if cmd.Short != "Open a pull request for each repository (placeholder)." {
		t.Errorf("expected short description 'Open a pull request for each repository (placeholder).', got %s", cmd.Short)
	}
	if cmd.Args != cobra.ArbitraryArgs {
		t.Errorf("expected arbitrary args, got %v", cmd.Args)
	}
}
