package commands

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestQueryCmd(t *testing.T) {
	cmd := QueryCmd
	if cmd.Use != "query <question>" {
		t.Errorf("expected use 'query <question>', got %s", cmd.Use)
	}
	if cmd.Short != "Generate a TypeScript script from an OpenAI query." {
		t.Errorf("expected short description 'Generate a TypeScript script from an OpenAI query.', got %s", cmd.Short)
	}
	if cmd.Args != cobra.ArbitraryArgs {
		t.Errorf("expected arbitrary args, got %v", cmd.Args)
	}
}
