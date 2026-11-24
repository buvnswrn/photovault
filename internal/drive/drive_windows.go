//go:build windows

package drive

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

// powershell script: uses Get-PhysicalDisk (Storage module) for accurate BusType
// (correctly reports USB even for UASP drives that WMI shows as SCSI), then
// walks the storage pipeline to resolve mount letters.
const wmiScript = `
$result = Get-PhysicalDisk | ForEach-Object {
  $mounts = $_ | Get-Disk | Get-Partition | Where-Object { $_.DriveLetter -ne 0 } |
    ForEach-Object { "$($_.DriveLetter):" }
  [PSCustomObject]@{
    Serial = $_.SerialNumber.Trim()
    Model  = $_.FriendlyName.Trim()
    Bus    = [string]$_.BusType
    Mounts = @($mounts)
  }
}
ConvertTo-Json -InputObject @($result) -Depth 2
`

type wmiDisk struct {
	Serial string   `json:"Serial"`
	Model  string   `json:"Model"`
	Bus    string   `json:"Bus"`
	Mounts []string `json:"Mounts"`
}

func list() ([]Disk, error) {
	out, err := exec.Command(
		"powershell", "-NoProfile", "-NonInteractive", "-Command", wmiScript,
	).Output()
	if err != nil {
		return nil, fmt.Errorf("drive: wmi query: %w", err)
	}

	// PowerShell may return null when no disks found.
	trimmed := strings.TrimSpace(string(out))
	if trimmed == "" || trimmed == "null" {
		return nil, nil
	}

	var raw []wmiDisk
	if err := json.Unmarshal([]byte(trimmed), &raw); err != nil {
		return nil, fmt.Errorf("drive: parse wmi output: %w", err)
	}

	disks := make([]Disk, 0, len(raw))
	for _, r := range raw {
		if r.Serial == "" {
			continue // virtual/system disks often have no serial
		}
		disks = append(disks, Disk{
			Serial: r.Serial,
			Model:  r.Model,
			Bus:    r.Bus,
			Mounts: r.Mounts,
		})
	}
	return disks, nil
}
