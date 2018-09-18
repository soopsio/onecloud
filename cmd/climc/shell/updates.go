package shell

import (
	"yunion.io/x/jsonutils"
	"yunion.io/x/onecloud/pkg/mcclient"
	"yunion.io/x/onecloud/pkg/mcclient/modules"
	"yunion.io/x/onecloud/pkg/mcclient/options"
)

func init() {
	type UpdateListOptions struct {
		options.BaseListOptions
		Region string `help:"cloud region ID or Name"`
	}
	R(&UpdateListOptions{}, "update-list", "List updates", func(s *mcclient.ClientSession, args *UpdateListOptions) error {
		var params *jsonutils.JSONDict
		{
			var err error
			params, err = args.BaseListOptions.Params()
			if err != nil {
				return err

			}
		}
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
}
