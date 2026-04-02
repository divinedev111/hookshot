package cmd

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/divinedev111/hookshot/internal/capture"
	"github.com/divinedev111/hookshot/internal/output"
	"github.com/divinedev111/hookshot/internal/store"
	"github.com/spf13/cobra"
)

func newListenCmd() *cobra.Command {
	var forward string
	var jsonMode bool

	cmd := &cobra.Command{
		Use:   "listen <addr>",
		Short: "Start a webhook capture server",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			addr := args[0]
			if !strings.Contains(addr, ":") {
				addr = ":" + addr
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

			onEvent := func(e *store.Event) {
				fmt.Println(output.FormatEvent(e, jsonMode))
			}

			h := capture.NewHandler(s, forward, onEvent)
			srv := &http.Server{Addr: addr, Handler: h}

			go func() {
				sig := make(chan os.Signal, 1)
				signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
				<-sig
				fmt.Fprintln(os.Stderr, "\nshutting down...")
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				srv.Shutdown(ctx)
			}()

			fwdMsg := ""
			if forward != "" {
				fwdMsg = fmt.Sprintf(" -> forwarding to %s", forward)
			}
			fmt.Fprintf(os.Stderr, "hookshot listening on %s%s\n", addr, fwdMsg)

			if err := srv.ListenAndServe(); err != http.ErrServerClosed {
				return fmt.Errorf("server error: %w", err)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&forward, "forward", "", "forward requests to this URL")
	cmd.Flags().BoolVar(&jsonMode, "json", false, "output events as JSON lines")
	return cmd
}
