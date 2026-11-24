package cmd

import (
	"strings"

	"github.com/buvnswrn/photovault/internal/config"
	"github.com/buvnswrn/photovault/internal/drive"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

var drivesCmd = &cobra.Command{
	Use:   "drives",
	Short: "List all attached drives, marking ones known from config",
	RunE: func(cmd *cobra.Command, args []string) error {
		knownBySerial := map[string]config.DriveConfig{}
		if cfg, err := config.Load(cfgFile); err == nil {
			for _, d := range cfg.Drives {
				knownBySerial[d.Serial] = d
			}
		}

		spinner, _ := pterm.DefaultSpinner.Start("Querying drives…")
		attached, err := drive.List()
		if err != nil {
			spinner.Fail(err.Error())
			return err
		}
		spinner.Stop()

		if len(attached) == 0 {
			pterm.Warning.Println("No drives found.")
			return nil
		}

		rows := pterm.TableData{{"Serial", "Model", "Bus", "Location", "Role", "Mounts"}}
		for _, d := range attached {
			location := "internal"
			locationStyle := pterm.FgGray
			if d.IsExternal() {
				location = "external"
				locationStyle = pterm.FgCyan
			}

			role := "-"
			if known, ok := knownBySerial[d.Serial]; ok {
				role = known.Role + " (" + known.Name + ")"
			}

			mounts := strings.Join(d.Mounts, ", ")
			if mounts == "" {
				mounts = "-"
			}

			rows = append(rows, []string{
				d.Serial,
				d.Model,
				d.Bus,
				locationStyle.Sprint(location),
				role,
				mounts,
			})
		}

		return pterm.DefaultTable.WithHasHeader().WithData(rows).Render()
	},
}

func init() {
	rootCmd.AddCommand(drivesCmd)
}
