package cmd

import (
	"fmt"
	"strconv"

	"github.com/divinedev111/hookshot/internal/output"
	"github.com/divinedev111/hookshot/internal/store"
	"github.com/spf13/cobra"
)

func newExportCmd() *cobra.Command {
	var format string

	cmd := &cobra.Command{
		Use:   "export <id>",
		Short: "Export a captured event as a curl command",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid event ID: %w", err)
			}

			s, err := store.Open(dbPath())
			if err != nil {
				return err
			}
			defer s.Close()

			e, err := s.Get(id)
			if err != nil {
				return fmt.Errorf("event %d not found: %w", id, err)
			}

			switch format {
			case "curl":
				fmt.Println(output.ExportAsCurl(e))
			default:
				return fmt.Errorf("unsupported format: %s (supported: curl)", format)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&format, "as", "curl", "export format (curl)")
	return cmd
}
