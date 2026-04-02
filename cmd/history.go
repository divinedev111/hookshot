package cmd

import (
	"fmt"

	"github.com/divinedev111/hookshot/internal/output"
	"github.com/divinedev111/hookshot/internal/store"
	"github.com/spf13/cobra"
)

func newHistoryCmd() *cobra.Command {
	var provider string
	var limit int
	var jsonMode bool

	cmd := &cobra.Command{
		Use:   "history",
		Short: "List captured webhook events",
		RunE: func(cmd *cobra.Command, args []string) error {
			path, err := dbPath()
			if err != nil {
				return err
			}
			s, err := store.Open(path)
			if err != nil {
				return err
			}
			defer s.Close()

			events, err := s.List(store.ListOpts{
				Provider: provider,
				Limit:    limit,
			})
			if err != nil {
				return err
			}

			fmt.Println(output.FormatEventList(events, jsonMode))
			return nil
		},
	}

	cmd.Flags().StringVar(&provider, "provider", "", "filter by provider")
	cmd.Flags().IntVar(&limit, "limit", 50, "max events to show")
	cmd.Flags().BoolVar(&jsonMode, "json", false, "output as JSON")
	return cmd
}
