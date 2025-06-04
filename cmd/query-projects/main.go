package queryprojects

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
	"github.com/wcatron/query-projects/internal/commands"
	"github.com/wcatron/query-projects/internal/version"
)

var cliVersion string // Define the current version of the CLI tool

var rootCmd = &cobra.Command{
	Use:     "query-projects",
	Short:   "A CLI that manages repositories and runs scripts across them.",
	Version: version.Version(),
}

func Execute() {
	// Attempt to load an .env file
	if err := godotenv.Load(); err != nil {
	}

	// Add subcommands
	rootCmd.AddCommand(commands.AddCmd)
	rootCmd.AddCommand(commands.AskCmd)
	rootCmd.AddCommand(commands.RunCmd)
	rootCmd.AddCommand(commands.PullCmd)
	rootCmd.AddCommand(commands.InfoCmd)
	rootCmd.AddCommand(commands.SyncCmd)
	rootCmd.AddCommand(commands.PlanCmd)
	rootCmd.AddCommand(commands.LoadCmd)

	// Add a flags for commands
	commands.RunCmdInit(commands.RunCmd)
	commands.LoadCmdInit(commands.LoadCmd)
	commands.PullCmdInit(commands.PullCmd)

	// Add flags for the root command
	rootCmd.PersistentFlags().StringSliceP("topics", "t", nil, "Filter projects by topics")
	rootCmd.PersistentFlags().Bool("debug", false, "Include additional information for debugging")

	// Execute the root command
	if err := rootCmd.Execute(); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}
