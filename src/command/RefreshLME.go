package command

import (
	"assistbot/global/env"
	"assistbot/src"
	"errors"
	"slices"
)

var RefreshLME = src.Command{
	Info: src.CmdInfo{
		Name:        "refreshlme",
		Description: "(Owners only) refresh Lenovo Model Explorer. NOT AFFILIATED IN ANY WAY WITH LENOVO",
	},
	Fn: func(args src.CmdResFnArgs) error {
		if !slices.Contains(env.Owners, args.Interaction.Message.Author.ID) {
			return errors.New("you arent in the Owners object")
		}
		return LMERefreshBscoDB()
	},
}
