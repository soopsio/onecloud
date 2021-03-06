package k8s

import (
	"yunion.io/x/onecloud/pkg/mcclient"
	"yunion.io/x/onecloud/pkg/mcclient/modules/k8s"
	o "yunion.io/x/onecloud/pkg/mcclient/options/k8s"
)

func initJob() {
	cmd := initK8sNamespaceResource("job", k8s.Jobs)
	cmdN := cmd.CommandNameFactory

	createCmd := NewCommand(
		&o.JobCreateOptions{},
		cmdN("create"),
		"Create job resource",
		func(s *mcclient.ClientSession, args *o.JobCreateOptions) error {
			params, err := args.Params()
			if err != nil {
				return err
			}
			ret, err := k8s.Jobs.Create(s, params)
			if err != nil {
				return err
			}
			printObject(ret)
			return nil
		})

	cmd.AddR(createCmd)
}
