package cmd

import (
	"fmt"

	"github.com/buvnswrn/photovault/internal/config"
	"github.com/buvnswrn/photovault/internal/drive"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

var drivesAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Register attached drives that are not yet in config",
	RunE: func(cmd *cobra.Command, args []string) error {
		known, err := config.KnownSerials(cfgFile)
		if err != nil {
			return err
		}

		spinner, _ := pterm.DefaultSpinner.Start("Querying drives…")
		attached, err := drive.List()
		spinner.Stop()
		if err != nil {
			return err
		}

		var unknown []drive.Disk
		for _, d := range attached {
			if !known[d.Serial] {
				unknown = append(unknown, d)
			}
		}

		if len(unknown) == 0 {
			pterm.Success.Println("All attached drives are already registered.")
			return nil
		}

		pterm.Info.Printf("%d unregistered drive(s) found.\n\n", len(unknown))

		for _, d := range unknown {
			location := "internal"
			if d.IsExternal() {
				location = "external"
			}
			pterm.DefaultSection.Printf("%s  [%s, %s]", d.Model, d.Bus, location)
			pterm.FgGray.Println("  serial: " + d.Serial)

			add, err := pterm.DefaultInteractiveConfirm.Show("  Register this drive?")
			if err != nil {
				return err
			}
			if !add {
				pterm.FgGray.Println("  Skipped.")
				continue
			}

			name, err := pterm.DefaultInteractiveTextInput.
				WithDefaultText(d.Model).
				Show("  Name")
			if err != nil {
				return err
			}

			role, err := pterm.DefaultInteractiveSelect.
				WithOptions([]string{"ssd", "hdd", "offsite"}).
				Show("  Role")
			if err != nil {
				return err
			}

			if err := config.AppendDrive(cfgFile, config.DriveConfig{
				Name:   name,
				Serial: d.Serial,
				Role:   role,
			}); err != nil {
				return fmt.Errorf("drives add: %w", err)
			}
			pterm.Success.Printf("Added %q (%s) to %s\n", name, role, cfgFile)
		}
		return nil
	},
}

func init() {
	drivesCmd.AddCommand(drivesAddCmd)
}
