package shell

import (
	"yunion.io/x/onecloud/pkg/mcclient"
	"yunion.io/x/onecloud/pkg/mcclient/modules"
)

func init() {

	/**
	 * 列出指定监控类型下的全部指标类型
	 */
	type MonitorTypesBaseOptions struct {
		ID string `help:"ID of the monitor type"`
	}
	R(&MonitorTypesBaseOptions{}, "monitortype-metrictype-list", "List metric types of the monitor type", func(s *mcclient.ClientSession, args *MonitorTypesBaseOptions) error {
		result, err := modules.MetricsTypes.ListInContext(s, nil, &modules.MonitorTypes, args.ID)

		if err != nil {
			return err
		}

		printList(result, modules.MetricsTypes.GetColumns(s))
		return nil
	})

}
