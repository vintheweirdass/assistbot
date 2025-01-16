package command

import (
	"assistbot/src"
)

var Hello = src.Command{
	Info: src.CmdInfo{
		Name:        "hello",
		Description: "you wasted 3 secs to see these",
		Options: src.CmdInfoOpt{
			{
				Name:        "name",
				Description: "whoever the name is",
				Type:        src.CmdInfoOptTypeEnum.String,
				Required:    false,
			},
		},
	},
	Fn: func(opt src.CmdResFnArgs) error {
		if opt.Args["name"] != nil {
			return opt.Result(&src.CmdResData{
				Content: "Hello, " + opt.Args["name"].StringValue() + " !",
			})
		}
		return opt.Result(&src.CmdResData{
			Content: "Hello!",
		})
	},
}
