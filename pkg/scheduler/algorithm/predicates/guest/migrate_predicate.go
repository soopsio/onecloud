package guest

import (
	"yunion.io/x/onecloud/pkg/scheduler/algorithm/predicates"
	"yunion.io/x/onecloud/pkg/scheduler/core"
)

// MigratePredicate filters whether the current candidate can be migrated.
type MigratePredicate struct {
	predicates.BasePredicate
}

func (p *MigratePredicate) Name() string {
	return "host_migrate"
}

func (p *MigratePredicate) Clone() core.FitPredicate {
	return &MigratePredicate{}
}

func (p *MigratePredicate) PreExecute(u *core.Unit, cs []core.Candidater) (bool, error) {
	return len(u.SchedData().HostId) > 0, nil
}

func (p *MigratePredicate) Execute(u *core.Unit, c core.Candidater) (bool, []core.PredicateFailureReason, error) {
	h := predicates.NewPredicateHelper(p, u, c)

	if u.SchedData().HostId == c.IndexKey() {
		h.Exclude(predicates.ErrHostIsSpecifiedForMigration)
	}

	return h.GetResult()
}
