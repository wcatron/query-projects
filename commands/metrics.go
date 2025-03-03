package commands

import (
    "fmt"
    "time"

    "github.com/spf13/cobra"
)

// WrapWithMetrics wraps a command function to log its execution duration.
func WrapWithMetrics(fn func(cmd *cobra.Command, args []string) error) func(cmd *cobra.Command, args []string) error {
    return func(cmd *cobra.Command, args []string) error {
        start := time.Now() // Start timing
        err := fn(cmd, args)
        duration := time.Since(start) // Calculate duration
        fmt.Printf("Command '%s' executed in %s\n", cmd.Name(), duration)
        return err
    }
}
