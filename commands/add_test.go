package commands

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestAddCmd(t *testing.T) {
	cmd := AddCmd
	if cmd.Use != "add <repo-url>" {
		t.Errorf("expected use 'add <repo-url>', got %s", cmd.Use)
	}
	if cmd.Short != "Add a new repository and clone it locally." {
		t.Errorf("expected short description 'Add a new repository and clone it locally.', got %s", cmd.Short)
	}
	if cmd.Args != cobra.ExactArgs(1) {
		t.Errorf("expected exact args 1, got %v", cmd.Args)
	}
}
