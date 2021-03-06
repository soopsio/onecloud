package shell

import (
	"yunion.io/x/jsonutils"

	"yunion.io/x/onecloud/pkg/mcclient"
	"yunion.io/x/onecloud/pkg/mcclient/modules"
	"yunion.io/x/onecloud/pkg/mcclient/options"
)

func init() {

	/**
	 * 列出所有监控指标
	 */
	type ProjectAdminListOptions struct {
		options.BaseListOptions
	}
	R(&ProjectAdminListOptions{}, "projectadmin-list", "List all Project Admins", func(s *mcclient.ClientSession, args *ProjectAdminListOptions) error {
		var params *jsonutils.JSONDict
		{
			var err error
			params, err = args.BaseListOptions.Params()
			if err != nil {
				return err

			}
		}

		result, err := modules.ProjectAdmin.List(s, params)
		if err != nil {
			return err
		}

		printList(result, modules.ProjectAdmin.GetColumns(s))
		return nil
	})

}
