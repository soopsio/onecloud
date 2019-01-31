package models

import (
	"yunion.io/x/onecloud/pkg/cloudcommon/db"
	"time"
)

type SActionlogManager struct {
	db.SOpsLogManager
}

type SActionlog struct {
	db.SOpsLog

	StartTime time.Time `nullable:"false" list:"user"`                           // = Column(DateTime, nullable=False)
	Success   bool      `default:"true" list:"user"`                             // = Column(Boolean, default=True)

}

var ActionLog *SActionlogManager

func init() {
	ActionLog = &SActionlogManager{db.SOpsLogManager{db.NewModelBaseManager(SActionlog{}, "action_tbl", "action", "actions")}}
}
