package azure

import (
	"fmt"
	"strings"

	"yunion.io/x/jsonutils"
	"yunion.io/x/log"
	"yunion.io/x/onecloud/pkg/cloudprovider"
	"yunion.io/x/onecloud/pkg/compute/models"
)

type SnapshotSku struct {
	Name string
	Tier string
}

type SSnapshot struct {
	region *SRegion

	ID         string
	Name       string
	Location   string
	ManagedBy  string
	Sku        SnapshotSku
	Properties DiskProperties
}

func (self *SSnapshot) GetId() string {
	return self.ID
}

func (self *SSnapshot) GetGlobalId() string {
	return strings.ToLower(self.ID)
}

func (self *SSnapshot) GetMetadata() *jsonutils.JSONDict {
	return nil
}

func (self *SSnapshot) GetName() string {
	return self.Name
}

func (self *SSnapshot) GetStatus() string {
	switch self.Properties.ProvisioningState {
	case "Succeeded":
		return models.SNAPSHOT_READY
	default:
		log.Errorf("Unknow azure snapshot %s status: %s", self.ID, self.Properties.ProvisioningState)
		return models.SNAPSHOT_UNKNOWN
	}
}

func (self *SSnapshot) IsEmulated() bool {
	return false
}

func (self *SRegion) CreateSnapshot(diskId, snapName, desc string) (*SSnapshot, error) {
	snapshot := SSnapshot{}
	return &snapshot, self.client.Create(jsonutils.Marshal(snapshot), &snapshot)
}

func (self *SSnapshot) Delete() error {
	return self.region.DeleteSnapshot(self.ID)
}

func (self *SSnapshot) GetSize() int32 {
	return self.Properties.DiskSizeGB
}

func (self *SRegion) DeleteSnapshot(snapshotId string) error {
	return self.client.Delete(snapshotId)
}

type AccessURIOutput struct {
	AccessSas string
}

type AccessProperties struct {
	Output AccessURIOutput
}

type AccessURI struct {
	Name       string
	Properties AccessProperties
}

func (self *SRegion) GrantAccessSnapshot(snapshotId string) (string, error) {
	body, err := self.client.PerformAction(snapshotId, "beginGetAccess", fmt.Sprintf(`{"access": "Read", "durationInSeconds": %d}`, 3600*24))
	if err != nil {
		return "", err
	}
	accessURI := AccessURI{}
	return accessURI.Properties.Output.AccessSas, body.Unmarshal(&accessURI)
}

func (self *SSnapshot) Refresh() error {
	snapshot, err := self.region.GetSnapshotDetail(self.ID)
	if err != nil {
		return err
	}
	return jsonutils.Update(self, snapshot)
}

func (self *SRegion) GetISnapshotById(snapshotId string) (cloudprovider.ICloudSnapshot, error) {
	if strings.HasPrefix(snapshotId, "https://") {
		//TODO
		return nil, cloudprovider.ErrNotImplemented
	}
	return self.GetSnapshotDetail(snapshotId)
}

func (self *SRegion) GetISnapshots() ([]cloudprovider.ICloudSnapshot, error) {
	snapshots, err := self.GetSnapShots("")
	if err != nil {
		return nil, err
	}
	classicSnapshots := []SClassicSnapshot{}
	storages, err := self.GetStorageAccounts()
	if err != nil {
		return nil, err
	}
	_, _classicSnapshots, err := self.GetStorageAccountsDisksWithSnapshots(storages...)
	if err != nil {
		return nil, err
	}
	classicSnapshots = append(classicSnapshots, _classicSnapshots...)
	classicStorages, err := self.GetClassicStorageAccounts()
	if err != nil {
		return nil, err
	}
	_, _classicSnapshots, err = self.GetStorageAccountsDisksWithSnapshots(classicStorages...)
	if err != nil {
		return nil, err
	}
	classicSnapshots = append(classicSnapshots, _classicSnapshots...)
	isnapshots := make([]cloudprovider.ICloudSnapshot, len(snapshots)+len(classicSnapshots))
	for i := 0; i < len(snapshots); i++ {
		isnapshots[i] = &snapshots[i]
	}
	for i := 0; i < len(classicSnapshots); i++ {
		isnapshots[len(snapshots)+i] = &classicSnapshots[i]
	}
	return isnapshots, nil
}

func (self *SSnapshot) GetDiskId() string {
	return self.Properties.CreationData.SourceResourceID
}

func (self *SSnapshot) GetManagerId() string {
	return self.region.client.providerId
}

func (self *SSnapshot) GetRegionId() string {
	return self.region.GetId()
}

func (self *SSnapshot) GetDiskType() string {
	return ""
}