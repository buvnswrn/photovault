//go:build linux

package drive

import (
	"encoding/json"
	"fmt"
	"os/exec"
)

type lsblkOutput struct {
	Blockdevices []lsblkDevice `json:"blockdevices"`
}

type lsblkDevice struct {
	Serial     string   `json:"serial"`
	Model      string   `json:"model"`
	Mountpoint *string  `json:"mountpoint"`
	Children   []lsblkDevice `json:"children"`
}

func list() ([]Disk, error) {
	out, err := exec.Command(
		"lsblk", "-J", "-d", "-o", "SERIAL,MODEL,MOUNTPOINT",
	).Output()
	if err != nil {
		return nil, fmt.Errorf("drive: lsblk: %w", err)
	}

	var raw lsblkOutput
	if err := json.Unmarshal(out, &raw); err != nil {
		return nil, fmt.Errorf("drive: parse lsblk output: %w", err)
	}

	disks := make([]Disk, 0, len(raw.Blockdevices))
	for _, d := range raw.Blockdevices {
		if d.Serial == "" {
			continue
		}
		disk := Disk{Serial: d.Serial, Model: d.Model}
		if d.Mountpoint != nil && *d.Mountpoint != "" {
			disk.Mounts = []string{*d.Mountpoint}
		}
		disks = append(disks, disk)
	}
	return disks, nil
}
