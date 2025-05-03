package main

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
	"github.com/wcatron/query-projects/internal/commands"
)

var cliVersion string // Define the current version of the CLI tool
func main() {
	// Attempt to load an .env file
	if err := godotenv.Load(); err != nil {
	}

	rootCmd := &cobra.Command{
		Use:     "query-projects",
		Short:   "A CLI that manages repositories and runs scripts across them.",
		Version: cliVersion,
	}

	// Add subcommands
	rootCmd.AddCommand(commands.AddCmd)
	rootCmd.AddCommand(commands.AskCmd)
	rootCmd.AddCommand(commands.RunCmd)
	rootCmd.AddCommand(commands.PullCmd)
	rootCmd.AddCommand(commands.InfoCmd)
	rootCmd.AddCommand(commands.SyncCmd)

	// Add a flag for the run
	commands.RunCmdInit(commands.RunCmd)

	// Add flags for the root command
	rootCmd.PersistentFlags().StringSliceP("topics", "t", nil, "Filter projects by topics")

	// Execute the root command
	if err := rootCmd.Execute(); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}
