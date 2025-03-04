package commands

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestPullCmd(t *testing.T) {
	cmd := PullCmd
	if cmd.Use != "pull" {
		t.Errorf("expected use 'pull', got %s", cmd.Use)
	}
	if cmd.Short != "Pull the latest changes for all repositories in projects.json" {
		t.Errorf("expected short description 'Pull the latest changes for all repositories in projects.json', got %s", cmd.Short)
	}
	if cmd.RunE == nil {
		t.Error("expected RunE to be defined")
	}
}
