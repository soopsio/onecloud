package models

import "yunion.io/x/onecloud/pkg/cloudcommon/db"

type SGuestJointsManager struct {
	db.SVirtualJointResourceBaseManager
}

func NewGuestJointsManager(dt interface{}, tableName string, keyword string, keywordPlural string, slave db.IVirtualModelManager) SGuestJointsManager {
	return SGuestJointsManager{
		SVirtualJointResourceBaseManager: db.NewVirtualJointResourceBaseManager(
			dt,
			tableName,
			keyword,
			keywordPlural,
			GuestManager,
			slave,
		),
	}
}

type SGuestJointsBase struct {
	db.SVirtualJointResourceBase

	GuestId string `width:"36" charset:"ascii" nullable:"false" list:"user" create:"required"` // Column(VARCHAR(36, charset='ascii'), nullable=False)
}

func (self *SGuestJointsBase) getGuest() *SGuest {
	guest, _ := GuestManager.FetchById(self.GuestId)
	return guest.(*SGuest)
}
