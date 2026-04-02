package cmd

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/divinedev111/hookshot/internal/capture"
	"github.com/divinedev111/hookshot/internal/store"
	"github.com/spf13/cobra"
)

func newReplayCmd() *cobra.Command {
	var to string
	var jsonMode bool

	cmd := &cobra.Command{
		Use:   "replay <id>",
		Short: "Replay a captured event to a target URL",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid event ID: %w", err)
			}

			path, err := dbPath()
			if err != nil {
				return err
			}
			s, err := store.Open(path)
			if err != nil {
				return err
			}
			defer s.Close()

			e, err := s.Get(id)
			if err != nil {
				return fmt.Errorf("event %d not found: %w", id, err)
			}

			status, resp, err := capture.Forward(to, e.Method, e.Headers, e.Body)
			if err != nil {
				return fmt.Errorf("replay failed: %w", err)
			}

			if jsonMode {
				out := map[string]any{
					"status": status,
					"body":   string(resp),
				}
				b, _ := json.MarshalIndent(out, "", "  ")
				fmt.Println(string(b))
			} else {
				fmt.Printf("Status: %d\n", status)
				if len(resp) > 0 {
					fmt.Printf("Response:\n%s\n", resp)
				}
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&to, "to", "", "target URL to replay to")
	cmd.MarkFlagRequired("to")
	cmd.Flags().BoolVar(&jsonMode, "json", false, "output as JSON")
	return cmd
}
