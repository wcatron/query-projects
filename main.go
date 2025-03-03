package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/wcatron/query-projects/commands"
)

func main() {
	// Define the root command
	rootCmd := &cobra.Command{
		Use:   "query-projects",
		Short: "A CLI that manages repositories and runs scripts across them.",
	}

	// Add subcommands
	rootCmd.AddCommand(commands.AddCmd)
	rootCmd.AddCommand(commands.QueryCmd)
	rootCmd.AddCommand(commands.RunCmd)
	rootCmd.AddCommand(commands.PRCmd)
	rootCmd.AddCommand(commands.PullCmd)
	rootCmd.AddCommand(commands.InfoCmd)

	// Add flags for the root command
	rootCmd.PersistentFlags().StringSliceP("topics", "t", nil, "Filter projects by topics")

	// Execute the root command
	if err := rootCmd.Execute(); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}
