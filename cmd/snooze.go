package cmd

import "github.com/spf13/cobra"

var snoozeCmd = &cobra.Command{
	Use:   "snooze",
	Short: "Postpone automatic backups for a duration",
	RunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}

func init() {
	rootCmd.AddCommand(snoozeCmd)
}
