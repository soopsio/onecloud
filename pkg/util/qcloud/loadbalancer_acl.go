package qcloud

import (
	"yunion.io/x/jsonutils"

	api "yunion.io/x/onecloud/pkg/apis/compute"
)

// 腾讯云没有LB ACL
type SLBACL struct{}

func (self *SLBACL) GetId() string {
	return ""
}

func (self *SLBACL) GetName() string {
	return ""
}

func (self *SLBACL) GetGlobalId() string {
	return ""
}

func (self *SLBACL) GetStatus() string {
	return api.LB_BOOL_OFF
}

func (self *SLBACL) Refresh() error {
	return nil
}

func (self *SLBACL) IsEmulated() bool {
	return false
}

func (self *SLBACL) GetMetadata() *jsonutils.JSONDict {
	return nil
}

func (self *SLBACL) GetAclEntries() *jsonutils.JSONArray {
	return nil
}

func (self *SLBACL) GetProjectId() string {
	return ""
}
