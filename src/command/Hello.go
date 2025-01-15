package command

import (
	"assistbot/src"
)

var Hello = src.Command{
	Info: &src.CmdInfo{
		Name:    "hello",
		Options: src.CmdInfoOpt{},
	},
	Fn: func(session src.Session, intr src.CmdIntr, res src.CmdResFn) {
		var opt, len = src.GetInteractionArgs(intr)
		if len < 0 {
			res(&src.CmdResData{
				Content: "hallo",
			})
		}
	},
}
