package cmd

import (
	"github.com/grazor/pkdb/pkg/serve"
	"github.com/pkg/profile"

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
	defer profile.Start().Stop()
	serve.Serve(args[0], args[1:])
}
