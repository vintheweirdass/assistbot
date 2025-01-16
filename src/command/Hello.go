package command

import (
	"assistbot/src"
)

var Hello = src.Command{
	Info: src.CmdInfo{
		Name: "hello",
		Options: src.CmdInfoOpt{
			{
				Name:        "name",
				Description: "The name",
				Required:    false,
			},
		},
	},
	Fn: func(opt src.CmdResFnArgs) error {
		if opt.Args["name"] != nil {
			return opt.Result(&src.CmdResData{
				Content: "Hello! " + opt.Args["name"].StringValue(),
			})
		}
		return opt.Result(&src.CmdResData{
			Content: "Hello!",
		})
	},
}
