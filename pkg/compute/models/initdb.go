package models

import (
	"yunion.io/x/log"

	"yunion.io/x/onecloud/pkg/cloudcommon/db"
)

func InitDB() error {
	for _, manager := range []db.IModelManager{
		/*
		 * Important!!!
		 * initialization order matters, do not change the order
		 */
		CloudproviderManager,
		CloudaccountManager,
		CloudregionManager,
		ZoneManager,
		VpcManager,
		WireManager,
		StorageManager,
		SecurityGroupManager,
		NetworkManager,
		LoadbalancerCertificateManager,
		LoadbalancerManager,
		LoadbalancerListenerManager,
		LoadbalancerListenerRuleManager,
		LoadbalancerBackendGroupManager,
		LoadbalancerBackendManager,
		SchedtagManager,
	} {
		err := manager.InitializeData()
		if err != nil {
			log.Errorf("Manager %s initializeData fail %s", manager.Keyword(), err)
			// return err skip error table
		}
	}
	return nil
}
