package cmd

import "github.com/spf13/cobra"

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialise repositories and configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
