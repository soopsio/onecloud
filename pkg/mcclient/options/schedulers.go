package options

import (
	"yunion.io/x/onecloud/pkg/apis/scheduler"
	"yunion.io/x/onecloud/pkg/mcclient"
)

type SchedulerTestBaseOptions struct {
	ServerConfigs

	Mem  int  `help:"Memory size (MB), default 512" metavar:"MEMORY" default:"512"`
	Ncpu int  `help:"#CPU cores of VM server, default 1" default:"1" metavar:"<SERVER_CPU_COUNT>"`
	Log  bool `help:"Record to schedule history"`
}

func (o SchedulerTestBaseOptions) data(s *mcclient.ClientSession) (*scheduler.ServerConfig, error) {
	config, err := o.ServerConfigs.Data()
	if err != nil {
		return nil, err
	}

	data := new(scheduler.ServerConfig)
	data.ServerConfigs = config

	if o.Mem > 0 {
		data.Memory = o.Mem
	}
	if o.Ncpu > 0 {
		data.Ncpu = o.Ncpu
	}
	return data, nil
}

func (o SchedulerTestBaseOptions) options() *scheduler.ScheduleBaseConfig {
	opt := new(scheduler.ScheduleBaseConfig)
	opt.RecordLog = o.Log
	return opt
}

type SchedulerTestOptions struct {
	SchedulerTestBaseOptions
	SuggestionLimit int64 `help:"Number of schedule candidate informations" default:"50"`
	SuggestionAll   bool  `help:"Show all schedule candidate informations"`
	Details         bool  `help:"Show suggestion details"`
}

func (o *SchedulerTestOptions) Params(s *mcclient.ClientSession) (*scheduler.ScheduleInput, error) {
	data, err := o.data(s)
	if err != nil {
		return nil, err
	}
	opts := o.options()
	input := new(scheduler.ScheduleInput)
	input.ServerConfig = *data
	input.ScheduleBaseConfig = *opts
	input.SuggestionLimit = o.SuggestionLimit
	input.SuggestionAll = o.SuggestionAll
	input.Details = o.Details

	return input, nil
}

type SchedulerForecastOptions struct {
	SchedulerTestBaseOptions
}

func (o SchedulerForecastOptions) Params(s *mcclient.ClientSession) (*scheduler.ScheduleInput, error) {
	data, err := o.data(s)
	if err != nil {
		return nil, err
	}
	opts := o.options()
	input := new(scheduler.ScheduleInput)
	input.ServerConfig = *data
	input.ScheduleBaseConfig = *opts
	return input, nil
}
