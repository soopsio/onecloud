package main

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"yunion.io/x/log"

	app_common "yunion.io/x/onecloud/pkg/cloudcommon/app"
	common_options "yunion.io/x/onecloud/pkg/cloudcommon/options"
	"yunion.io/x/onecloud/pkg/lbagent"
)

func main() {
	opts := &lbagent.Options{}
	commonOpts := &opts.CommonOptions
	{
		common_options.ParseOptions(opts, os.Args, "lbagent.conf", "lbagent")
		app_common.InitAuth(commonOpts, func() {
			log.Infof("auth finished ok")
		})
	}
	if err := opts.ValidateThenInit(); err != nil {
		log.Fatalf("opts validate: %s", err)
	}

	var haproxyHelper *lbagent.HaproxyHelper
	var apiHelper *lbagent.ApiHelper
	var haStateWatcher *lbagent.HaStateWatcher
	var err error
	{
		haStateWatcher, err = lbagent.NewHaStateWatcher(opts)
		if err != nil {
			log.Fatalf("init ha state watcher failed: %s", err)
		}
	}
	{
		haproxyHelper, err = lbagent.NewHaproxyHelper(opts)
		if err != nil {
			log.Fatalf("init haproxy helper failed: %s", err)
		}
	}
	{
		apiHelper, err = lbagent.NewApiHelper(opts)
		if err != nil {
			log.Fatalf("init api helper failed: %s", err)
		}
		apiHelper.SetHaStateProvider(haStateWatcher)
	}

	{
		wg := &sync.WaitGroup{}
		cmdChan := make(chan *lbagent.LbagentCmd) // internal
		ctx, cancelFunc := context.WithCancel(context.Background())
		ctx = context.WithValue(ctx, "wg", wg)
		ctx = context.WithValue(ctx, "cmdChan", cmdChan)
		wg.Add(3)
		go haStateWatcher.Run(ctx)
		go haproxyHelper.Run(ctx)
		go apiHelper.Run(ctx)

		go func() {
			sigChan := make(chan os.Signal)
			signal.Notify(sigChan, syscall.SIGINT)
			signal.Notify(sigChan, syscall.SIGTERM)
			sig := <-sigChan
			log.Infof("signal received: %s", sig)
			cancelFunc()
		}()
		wg.Wait()
	}
}
