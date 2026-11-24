package drive

// Disk is a physical disk currently attached to the system.
type Disk struct {
	Serial   string
	Model    string
	Bus      string   // USB, SATA, NVMe, SCSI, IDE, …
	Mounts   []string // e.g. ["C:", "D:"] on Windows or ["/mnt/photos"] on Linux
}

// IsExternal returns true when the disk is connected over USB.
func (d Disk) IsExternal() bool { return d.Bus == "USB" }

// List returns all physical disks currently visible to the OS.
// The implementation is OS-specific (see drive_windows.go, drive_linux.go).
func List() ([]Disk, error) {
	return list()
}
