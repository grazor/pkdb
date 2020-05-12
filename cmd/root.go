package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use: "pkdb",
	Short: "Personal Knowledge Database management util",
	Long: "Personal Knowledge Database management util",
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(serveCmd)
}
