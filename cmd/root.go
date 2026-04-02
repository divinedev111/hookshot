package cmd

import (
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

func dbPath() string {
	if p := os.Getenv("HOOKSHOT_DB"); p != "" {
		return p
	}
	home, _ := os.UserHomeDir()
	dir := filepath.Join(home, ".hookshot")
	os.MkdirAll(dir, 0o755)
	return filepath.Join(dir, "hookshot.db")
}
