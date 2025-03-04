package commands

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestInfoCmd(t *testing.T) {
	cmd := InfoCmd
	if cmd.Use != "info" {
		t.Errorf("expected use 'info', got %s", cmd.Use)
	}
	if cmd.Short != "Displays information about projects and available scripts" {
		t.Errorf("expected short description 'Displays information about projects and available scripts', got %s", cmd.Short)
	}
	if cmd.RunE == nil {
		t.Error("expected RunE to be defined")
	}
}
