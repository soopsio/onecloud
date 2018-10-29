package azure

import (
	"strings"
	"time"

	"yunion.io/x/jsonutils"
	"yunion.io/x/log"
	"yunion.io/x/onecloud/pkg/cloudprovider"
	"yunion.io/x/onecloud/pkg/compute/models"
)

type SClassicDisk struct {
	storage *SClassicStorage

	DiskName        string
	Caching         string
	OperatingSystem string
	IoType          string
	DiskSizeGB      int32
	DiskSize        int32
	diskSizeMB      int32
	CreatedTime     string
	SourceImageName string
	VhdUri          string
	diskType        string
	StorageAccount  SubResource
}

func (self *SRegion) GetStorageAccountsDisksWithSnapshots(storageaccounts ...SStorageAccount) ([]SClassicDisk, []SClassicSnapshot, error) {
	disks, snapshots := []SClassicDisk{}, []SClassicSnapshot{}
	for i := 0; i < len(storageaccounts); i++ {
		_disks, _snapshots, err := self.GetStorageAccountDisksWithSnapshots(storageaccounts[i])
		if err != nil {
			return nil, nil, err
		}
		disks = append(disks, _disks...)
		snapshots = append(snapshots, _snapshots...)
	}
	return disks, snapshots, nil
}

func (self *SRegion) GetStorageAccountDisksWithSnapshots(storageaccount SStorageAccount) ([]SClassicDisk, []SClassicSnapshot, error) {
	disks, snapshots := []SClassicDisk{}, []SClassicSnapshot{}
	containers, err := storageaccount.GetContainers()
	if err != nil {
		return nil, nil, err
	}
	for _, container := range containers {
		if container.Name == "vhds" {
			files, err := container.ListFiles()
			if err != nil {
				log.Errorf("List storage %s container %s files error: %v", storageaccount.Name, container.Name, err)
				return nil, nil, err
			}

			for _, file := range files {
				if strings.HasSuffix(file.Name, ".vhd") {
					diskType := models.DISK_TYPE_DATA
					if _diskType, ok := file.Metadata["microsoftazurecompute_disktype"]; ok && _diskType == "OSDisk" {
						diskType = models.DISK_TYPE_SYS
					}
					diskName := file.Name
					if _diskName, ok := file.Metadata["microsoftazurecompute_diskname"]; ok {
						diskName = _diskName
					}
					if file.Snapshot.IsZero() {
						disks = append(disks, SClassicDisk{
							DiskName:   diskName,
							diskType:   diskType,
							DiskSizeGB: int32(file.Properties.ContentLength / 1024 / 1024 / 1024),
							diskSizeMB: int32(file.Properties.ContentLength / 1024 / 1024),
							VhdUri:     file.GetURL(),
						})
					} else {
						snapshots = append(snapshots, SClassicSnapshot{
							region:   self,
							Name:     file.Snapshot.String(),
							sizeMB:   int32(file.Properties.ContentLength / 1024 / 1024),
							diskID:   file.GetURL(),
							diskName: diskName,
						})
					}
				}
			}
		}
	}
	return disks, snapshots, nil
}

func (self *SRegion) GetClassicDisks() ([]SClassicDisk, error) {
	storageaccounts, err := self.GetClassicStorageAccounts()
	if err != nil {
		return nil, err
	}
	disks, _, err := self.GetStorageAccountsDisksWithSnapshots(storageaccounts...)
	if err != nil {
		return nil, err
	}
	return disks, nil
}

func (self *SClassicDisk) GetMetadata() *jsonutils.JSONDict {
	return nil
}

func (self *SClassicDisk) CreateISnapshot(name, desc string) (cloudprovider.ICloudSnapshot, error) {
	return nil, cloudprovider.ErrNotSupported
}

func (self *SClassicDisk) Delete() error {
	return cloudprovider.ErrNotImplemented
}

func (self *SClassicDisk) GetBillingType() string {
	return models.BILLING_TYPE_POSTPAID
}

func (self *SClassicDisk) GetFsFormat() string {
	return ""
}

func (self *SClassicDisk) GetIsNonPersistent() bool {
	return false
}

func (self *SClassicDisk) GetDriver() string {
	return "scsi"
}

func (self *SClassicDisk) GetCacheMode() string {
	return "none"
}

func (self *SClassicDisk) GetMountpoint() string {
	return ""
}

func (self *SClassicDisk) GetDiskFormat() string {
	return "vhd"
}

func (self *SClassicDisk) GetDiskSizeMB() int {
	if self.DiskSizeGB > 0 {
		return int(self.DiskSizeGB * 1024)
	}
	return int(self.diskSizeMB)
}

func (self *SClassicDisk) GetIsAutoDelete() bool {
	return false
}

func (self *SClassicDisk) GetTemplateId() string {
	return ""
}

func (self *SClassicDisk) GetDiskType() string {
	return self.diskType
}

func (self *SClassicDisk) GetExpiredAt() time.Time {
	return time.Now()
}

func (self *SClassicDisk) GetGlobalId() string {
	return self.VhdUri
}

func (self *SClassicDisk) GetId() string {
	return self.VhdUri
}

func (self *SClassicDisk) GetISnapshot(snapshotId string) (cloudprovider.ICloudSnapshot, error) {
	return nil, cloudprovider.ErrNotSupported
}

func (region *SRegion) GetClassicSnapShots(diskId string) ([]SClassicSnapshot, error) {
	result := []SClassicSnapshot{}
	return result, nil
}

func (self *SClassicDisk) GetISnapshots() ([]cloudprovider.ICloudSnapshot, error) {
	return nil, cloudprovider.ErrNotSupported
}

func (self *SClassicDisk) GetIStorge() cloudprovider.ICloudStorage {
	return self.storage
}

func (self *SClassicDisk) GetName() string {
	return self.DiskName
}

func (self *SClassicDisk) GetStatus() string {
	return models.DISK_READY
}

func (self *SClassicDisk) IsEmulated() bool {
	return false
}

func (self *SClassicDisk) Refresh() error {
	return nil
}

func (self *SClassicDisk) Reset(snapshotId string) error {
	return cloudprovider.ErrNotSupported
}

func (self *SClassicDisk) Resize(size int64) error {
	return cloudprovider.ErrNotSupported
}