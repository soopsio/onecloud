package shell

import (
	"fmt"

	"yunion.io/x/jsonutils"

	"yunion.io/x/onecloud/pkg/mcclient"
	"yunion.io/x/onecloud/pkg/mcclient/modules"
	"yunion.io/x/onecloud/pkg/mcclient/options"
)

func init() {
	type SecGroupsListOptions struct {
		options.BaseListOptions
	}

	R(&SecGroupsListOptions{}, "secgroup-list", "List all security group", func(s *mcclient.ClientSession, args *SecGroupsListOptions) error {
		var params *jsonutils.JSONDict
		{
			var err error
			params, err = args.BaseListOptions.Params()
			if err != nil {
				return err

			}
		}
		result, err := modules.SecGroups.List(s, params)
		if err != nil {
			return err
		}
		printList(result, modules.SecGroups.GetColumns(s))
		return nil
	})

	type SecGroupsCreateOptions struct {
		NAME  string   `help:"Name of security group to create"`
		RULES []string `help:"security rule to create"`
		Desc  string   `help:"Description"`
	}

	R(&SecGroupsCreateOptions{}, "secgroup-create", "Create a security group", func(s *mcclient.ClientSession, args *SecGroupsCreateOptions) error {
		params := jsonutils.NewDict()
		params.Add(jsonutils.NewString(args.NAME), "name")
		if len(args.Desc) > 0 {
			params.Add(jsonutils.NewString(args.Desc), "description")
		}
		for i, a := range args.RULES {
			params.Add(jsonutils.NewString(a), fmt.Sprintf("rule.%d", i))
		}
		secgroups, err := modules.SecGroups.Create(s, params)
		if err != nil {
			return err
		}
		printObject(secgroups)
		return nil

	})

	type SecGroupsDetailOptions struct {
		ID string `help:"ID or Name of security group"`
	}
	R(&SecGroupsDetailOptions{}, "secgroup-show", "Show details of a security group", func(s *mcclient.ClientSession, args *SecGroupsDetailOptions) error {
		result, err := modules.SecGroups.Get(s, args.ID, nil)
		if err != nil {
			return err
		}
		printObject(result)
		return nil
	})
	R(&SecGroupsDetailOptions{}, "secgroup-delete", "Delete a security group", func(s *mcclient.ClientSession, args *SecGroupsDetailOptions) error {
		secgroups, err := modules.SecGroups.Delete(s, args.ID, nil)
		if err != nil {
			return err
		}
		printObject(secgroups)
		return nil
	})

	R(&SecGroupsDetailOptions{}, "secgroup-public", "Make a security group publicly available", func(s *mcclient.ClientSession, args *SecGroupsDetailOptions) error {
		result, err := modules.SecGroups.PerformAction(s, args.ID, "public", nil)
		if err != nil {
			return err
		}
		printObject(result)
		return nil
	})

	R(&SecGroupsDetailOptions{}, "secgroup-private", "Make a security group private", func(s *mcclient.ClientSession, args *SecGroupsDetailOptions) error {
		result, err := modules.SecGroups.PerformAction(s, args.ID, "private", nil)
		if err != nil {
			return err
		}
		printObject(result)
		return nil
	})

	type SecGroupsUpdateOptions struct {
		ID   string `help:"ID of security group"`
		Name string `help:"Name of security group to update"`
		Desc string `help:"Description of security groups"`
	}

	R(&SecGroupsUpdateOptions{}, "secgroup-update", "Update details of a security group", func(s *mcclient.ClientSession, args *SecGroupsUpdateOptions) error {
		params := jsonutils.NewDict()
		if len(args.Name) > 0 {
			params.Add(jsonutils.NewString(args.Name), "name")
		}
		if len(args.Desc) > 0 {
			params.Add(jsonutils.NewString(args.Desc), "description")
		}
		secgroups, err := modules.SecGroups.Update(s, args.ID, params)
		if err != nil {
			return err
		}
		printObject(secgroups)
		return nil
	})

	type SecGroupsAddRuleOptions struct {
		ID          string `help:"ID or Name of security group"`
		DIRECTION   string `help:"Direction of rule" choices:"in|out"`
		PROTOCOL    string `help:"Protocol of rule" choices:"any|tcp|udp|icmp"`
		ACTION      string `help:"Actin of rule" choices:"allow|deny"`
		PRIORITY    int    `help:"Priority for rule, range 1 ~ 100"`
		Cidr        string `help:"IP or CIRD for rule"`
		Description string `help:"Desciption for rule"`
		Ports       string `help:"Port for rule"`
	}

	R(&SecGroupsAddRuleOptions{}, "secgroup-add-rule", "Add rule for a security group", func(s *mcclient.ClientSession, args *SecGroupsAddRuleOptions) error {
		params, err := options.StructToParams(args)
		secgroups, err := modules.SecGroups.PerformAction(s, args.ID, "add-rule", params)
		if err != nil {
			return err
		}
		printObject(secgroups)
		return nil
	})
}
