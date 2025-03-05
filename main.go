package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/joho/godotenv"
	"github.com/wcatron/query-projects/commands"
)

func main() {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		fmt.Println("No .env file found or error loading .env file")
	}
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
