package cmd

import "github.com/spf13/cobra"

var drivesCmd = &cobra.Command{
	Use:   "drives",
	Short: "List known drives and their current mount paths",
	RunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}

func init() {
	rootCmd.AddCommand(drivesCmd)
}
