package models

import (
	"context"

	"yunion.io/x/jsonutils"
	"yunion.io/x/log"

	"yunion.io/x/onecloud/pkg/cloudcommon/db"
	"yunion.io/x/onecloud/pkg/mcclient"
)

type SGroupnetworkManager struct {
	SGroupJointsManager
}

var GroupnetworkManager *SGroupnetworkManager

func init() {
	db.InitManager(func() {
		GroupnetworkManager = &SGroupnetworkManager{SGroupJointsManager: NewGroupJointsManager(SGroupnetwork{}, "groupnetworks_tbl", "groupnetwork", "groupnetworks", NetworkManager)}
	})
}

type SGroupnetwork struct {
	SGroupJointsBase

	NetworkId string `width:"36" charset:"ascii" nullable:"false" list:"user" create:"required"` // Column(VARCHAR(36, charset='ascii'), nullable=False)

	IpAddr string `width:"16" charset:"ascii" nullable:"true" list:"user" create:"optional"` // Column(VARCHAR(16, charset='ascii'), nullable=True)
	// # ip6_addr = Column(VARCHAR(64, charset='ascii'), nullable=True)

	Index int8 `nullable:"false" default:"0" list:"user" list:"user" update:"user" create:"optional"` // Column(TINYINT, nullable=False, default=0)

	EipId string `width:"36" charset:"ascii" nullable:"true"` // Column(VARCHAR(36, charset='ascii'), nullable=True)
}

func (joint *SGroupnetwork) Master() db.IStandaloneModel {
	return db.JointMaster(joint)
}

func (joint *SGroupnetwork) Slave() db.IStandaloneModel {
	return db.JointSlave(joint)
}

func (self *SGroupnetwork) GetCustomizeColumns(ctx context.Context, userCred mcclient.TokenCredential, query jsonutils.JSONObject) *jsonutils.JSONDict {
	extra := self.SGroupJointsBase.GetCustomizeColumns(ctx, userCred, query)
	return db.JointModelExtra(self, extra)
}

func (self *SGroupnetwork) GetExtraDetails(ctx context.Context, userCred mcclient.TokenCredential, query jsonutils.JSONObject) (*jsonutils.JSONDict, error) {
	extra, err := self.SGroupJointsBase.GetExtraDetails(ctx, userCred, query)
	if err != nil {
		return nil, err
	}
	return db.JointModelExtra(self, extra), nil
}

func (self *SGroupnetwork) GetNetwork() *SNetwork {
	obj, err := NetworkManager.FetchById(self.NetworkId)
	if err != nil {
		log.Errorf("%s", err)
		return nil
	}
	return obj.(*SNetwork)
}

func (self *SGroupnetwork) Delete(ctx context.Context, userCred mcclient.TokenCredential) error {
	return db.DeleteModel(ctx, userCred, self)
}

func (self *SGroupnetwork) Detach(ctx context.Context, userCred mcclient.TokenCredential) error {
	return db.DetachJoint(ctx, userCred, self)
}
