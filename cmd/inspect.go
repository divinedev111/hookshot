package cmd

import (
	"fmt"
	"strconv"

	"github.com/divinedev111/hookshot/internal/output"
	"github.com/divinedev111/hookshot/internal/store"
	"github.com/spf13/cobra"
)

func newInspectCmd() *cobra.Command {
	var jsonMode bool

	cmd := &cobra.Command{
		Use:   "inspect <id>",
		Short: "Show full details of a captured event",
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

			fmt.Println(output.FormatEventDetail(e, jsonMode))
			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonMode, "json", false, "output as JSON")
	return cmd
}
