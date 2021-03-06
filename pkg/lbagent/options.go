package lbagent

import (
	"fmt"
	"os"
	"path/filepath"

	common_options "yunion.io/x/onecloud/pkg/cloudcommon/options"
	agentutils "yunion.io/x/onecloud/pkg/lbagent/utils"
)

type LbagentOptions struct {
	ApiLbagentId                  string `require:"true"`
	ApiLbagentHbInterval          int    `default:"10"`
	ApiLbagentHbTimeoutRelaxation int    `default:"120" help:"If agent is to stale out in specified seconds in the future, consider it staled to avoid race condition when doing incremental api data fetch"`

	ApiSyncInterval  int
	ApiListBatchSize int `default:"1024"`

	DataPreserveN int `default:"8" help:"number of recent data to preserve on disk"`

	BaseDataDir      string // `required:"true"`
	apiDataStoreDir  string
	haproxyConfigDir string
	haproxyRunDir    string
	haproxyShareDir  string
	haStateChan      chan string

	KeepalivedBin string `default:"keepalived"`
	HaproxyBin    string `default:"haproxy"`
	GobetweenBin  string `default:"gobetween"`
	TelegrafBin   string `default:"telegraf"`
}

type Options struct {
	common_options.CommonOptions

	LbagentOptions
}

func (opts *Options) ValidateThenInit() error {
	if opts.ApiListBatchSize <= 0 {
		return fmt.Errorf("negative api batch list size: %d",
			opts.ApiListBatchSize)
	}
	if err := opts.initDirs(); err != nil {
		return err
	}

	return nil
}

func (opts *Options) initDirs() error {
	opts.apiDataStoreDir = filepath.Join(opts.BaseDataDir, "data")
	opts.haproxyConfigDir = filepath.Join(opts.BaseDataDir, "configs")
	opts.haproxyRunDir = filepath.Join(opts.BaseDataDir, "run")
	opts.haproxyShareDir = filepath.Join(opts.BaseDataDir, "share")
	dirs := []string{
		opts.apiDataStoreDir,
		opts.haproxyConfigDir,
		opts.haproxyRunDir,
		opts.haproxyShareDir,
	}
	for _, dir := range dirs {
		err := os.MkdirAll(dir, agentutils.FileModeDir)
		if err != nil {
			return fmt.Errorf("mkdir -p %q: %s",
				dir, err)
		}
	}

	return nil
}
