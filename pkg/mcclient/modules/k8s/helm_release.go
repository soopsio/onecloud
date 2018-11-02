package k8s

import (
	"yunion.io/x/jsonutils"

	"yunion.io/x/onecloud/pkg/mcclient"
	"yunion.io/x/onecloud/pkg/mcclient/modules"
)

var (
	Releases         *ReleaseManager
	MeterReleaseApps *ReleaseAppManager
	dummyReleaseApps *ReleaseAppManager
)

type ReleaseManager struct {
	*NamespaceResourceManager
}

type ReleaseAppManager struct {
	*NamespaceResourceManager
}

func NewReleaseAppManager(keyword, keywordPlural string) *ReleaseAppManager {
	return &ReleaseAppManager{
		NamespaceResourceManager: NewNamespaceResourceManager(keyword, keywordPlural, NewColumns(), NewColumns()),
	}
}

func (m *ReleaseAppManager) Create(session *mcclient.ClientSession, params jsonutils.JSONObject) (jsonutils.JSONObject, error) {
	return m.CreateInContext(session, params, dummyReleaseApps, "")
}

func init() {
	Releases = &ReleaseManager{
		NewNamespaceResourceManager("release", "releases", NewColumns(), NewColumns())}
	dummyReleaseApps = NewReleaseAppManager("releaseapp", "releaseapps")
	MeterReleaseApps = NewReleaseAppManager("app_meter", "app_meters")
	modules.Register(Releases)
	modules.Register(MeterReleaseApps)
}
