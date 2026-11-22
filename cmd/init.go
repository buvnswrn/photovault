package cmd

import (
	"fmt"
	"os"

	"github.com/buvnswrn/photovault/internal/config"
	"github.com/spf13/cobra"
)

var initForce bool

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialise repositories and configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		if _, err := os.Stat(cfgFile); err == nil && !initForce {
			return fmt.Errorf("%s already exists; use --force to overwrite", cfgFile)
		}
		if err := config.WriteExample(cfgFile); err != nil {
			return err
		}
		fmt.Printf("wrote starter config to %s\n", cfgFile)
		fmt.Println("Edit the file, then run: photovault status")
		return nil
	},
}

func init() {
	initCmd.Flags().BoolVar(&initForce, "force", false, "overwrite an existing config file")
	rootCmd.AddCommand(initCmd)
}
