package shell

import (
	"yunion.io/x/jsonutils"
	"yunion.io/x/onecloud/pkg/mcclient"
	"yunion.io/x/onecloud/pkg/mcclient/modules"
	"yunion.io/x/onecloud/pkg/mcclient/options"
)

func init() {
	type SUpdateListOptions struct {
		options.BaseListOptions
		Region string `help:"cloud region ID or Name"`
	}

	R(&SUpdateListOptions{}, "update-list", "List updates", func(s *mcclient.ClientSession, args *SUpdateListOptions) error {
		var params *jsonutils.JSONDict
		var err error
		var result *modules.ListResult
		if len(args.Region) > 0 {
			result, err = modules.Updates.ListInContext(s, params, &modules.Cloudregions, args.Region)
		} else {
			result, err = modules.Updates.List(s, params)
		}
		if err != nil {
			return err
		}
		printList(result, modules.Updates.GetColumns(s))
		return nil
	})

	type SUpdatePerformOptions struct {
		Cmp bool `help:"update all the compute nodes automatically"`
	}

	R(&SUpdatePerformOptions{}, "update-perform", "Update the Controler", func(s *mcclient.ClientSession, args *SUpdatePerformOptions) error {
		params := jsonutils.NewDict()
		if args.Cmp {
			params.Add(jsonutils.JSONTrue, "cmp")
		}

		result, err := modules.Updates.List(s, nil)

		if err != nil {
			return err
		}
		modules.Updates.DoUpdate(s, params)
		printList(result, modules.Updates.GetColumns(s))
		return nil
	})
}
