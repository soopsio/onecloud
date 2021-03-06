package shell

import (
	"yunion.io/x/onecloud/pkg/util/qcloud"
	"yunion.io/x/onecloud/pkg/util/shellutils"
)

func init() {
	type ProjectListOptions struct {
	}
	shellutils.R(&ProjectListOptions{}, "project-list", "List project", func(cli *qcloud.SRegion, args *ProjectListOptions) error {
		project, err := cli.GetClient().GetIProjects()
		if err != nil {
			return err
		}
		printList(project, 0, 0, 0, nil)
		return nil
	})
}
