package cmd

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/grazor/pkdb/pkg/runner"
	"github.com/spf13/cobra"
)

var serveCmd = &cobra.Command{
	Use:   "serve <source> [<interbace>[ <interface>[ ...]]]",
	Args:  cobra.MinimumNArgs(2),
	Short: "Monitor pkdb documents and serve via one or more protocols",
	Long:  "Waches pkdb filesystem, (re-)indexes data and serves via defined interfaces",
	Run:   runServer,
}

func runServer(cmd *cobra.Command, args []string) {
	ctx, cancel := context.WithCancel(context.Background())

	log.Println("Starting")
	wg, err := runner.Serve(ctx, args[0], args[1:]...)
	if err != nil {
		cancel()
		return
	}

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	wg.Add(1)
	go func() {
		defer wg.Done()
		<-signals
		cancel()
		log.Println("Terminating")
	}()
	wg.Wait()
}
