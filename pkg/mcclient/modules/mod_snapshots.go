package modules

var (
	Snapshots ResourceManager
)

func init() {
	Snapshots = NewComputeManager("snapshot", "snapshots",
		[]string{"ID", "Name", "Size", "Status",
			"Disk_id", "Guest_id", "Created_at"},
		[]string{"Storage_id", "Storage_type", "Create_by", "Location", "Out_of_chain", "disk_type", "provider"})

	registerComputeV2(&Snapshots)
}
