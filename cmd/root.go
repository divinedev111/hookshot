package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

// NewRoot returns the root cobra command with all subcommands registered.
func NewRoot() *cobra.Command {
	root := &cobra.Command{
		Use:   "hookshot",
		Short: "Webhook testing from your terminal",
		Long:  "Capture, inspect, replay, and export webhook requests. Provider-aware with SQLite-backed persistence.",
		SilenceUsage: true,
	}

	root.AddCommand(
		newListenCmd(),
		newHistoryCmd(),
		newInspectCmd(),
		newReplayCmd(),
		newExportCmd(),
	)

	return root
}

func dbPath() (string, error) {
	if p := os.Getenv("HOOKSHOT_DB"); p != "" {
		return p, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolving home directory: %w", err)
	}
	dir := filepath.Join(home, ".hookshot")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("creating %s: %w", dir, err)
	}
	return filepath.Join(dir, "hookshot.db"), nil
}
