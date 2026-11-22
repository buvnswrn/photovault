package cmd

import "github.com/spf13/cobra"

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Run a manual backup cycle",
	RunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}

func init() {
	rootCmd.AddCommand(syncCmd)
}
