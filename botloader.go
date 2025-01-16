package main

import (
	"assistbot/opt"
	"assistbot/src"
	"log"

	"github.com/bwmarrin/discordgo"
)

func triggerLoadHook(data src.Session) {
	for _, o := range opt.Hooks.OnLoad {
		o(data)
	}
}
func triggerErrorHook(data src.ErrorHookData) {
	for _, o := range opt.Hooks.OnError {
		o(data)
	}
}

func HookLoader(session src.Session) {
	if len(opt.Hooks.OnLoad) > 0 {
		triggerLoadHook(session)
	}
	if len(opt.Hooks.OnSession) > 0 {
		for _, o := range opt.Hooks.OnSession {
			session.AddHandler(func(s src.Session, r *discordgo.Ready) {
				o(s, r)
			})
		}
	}
}

var cmdsObj = map[string]src.Command{}
var cmdsObjArgs = map[string]map[string]*discordgo.ApplicationCommandOption{}

func CommandLoader(session src.Session) {
	for _, k := range opt.Commands {
		cmdsObj[k.Info.Name] = k
		if len(k.Info.Options) < 1 {
			continue
		}
		var o = map[string]*discordgo.ApplicationCommandOption{}
		for _, arg := range k.Info.Options {
			if arg == nil {
				break
			}
			o[arg.Name] = arg
		}
		cmdsObjArgs[k.Info.Name] = o
	}
	helpCommandLoader(session)
	session.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		var data = i.ApplicationCommandData()
		if cmd, defined := cmdsObj[data.Name]; defined {
			options := data.Options
			optLen := len(options)
			var result = func(err error) {
				if err != nil {
					triggerErrorHook(src.ErrorHookData{
						CmdInfo:     cmd.Info,
						Message:     err.Error(),
						Interaction: i,
						Session:     s,
					})
				}
			}
			if optLen > 0 {
				optionMap := make(map[string]src.ACIDO, len(options))
				for _, opt := range options {
					optionMap[opt.Name] = opt
				}
				for name, p := range cmdsObjArgs[data.Name] {
					if p != nil && p.Required && optionMap[name] == nil {
						triggerErrorHook(src.ErrorHookData{
							CmdInfo:     cmd.Info,
							Message:     "`" + name + "` argument are required",
							Interaction: i,
							Session:     s,
						})
						return
					}
				}
				result(cmd.Fn(src.CmdResFnArgs{
					Session:     s,
					Interaction: i,
					Result: func(e *src.CmdResData) error {
						return src.InteractionRespondRaw(s, i, e)
					},
					Args:    optionMap,
					ArgsLen: optLen,
				}))
				return
			}
			result(cmd.Fn(src.CmdResFnArgs{
				Session:     s,
				Interaction: i,
				Result: func(e *src.CmdResData) error {
					return src.InteractionRespondRaw(s, i, e)
				},
				ArgsLen: 0,
			}))
		}
	})
	go func() {
		for _, guild := range session.State.Guilds {
			var cmds = make([]*src.CmdInfo, len(opt.Commands))
			for i, cmd := range opt.Commands {
				cmds[i] = &cmd.Info
			}
			_, err := session.ApplicationCommandBulkOverwrite(session.State.User.ID, guild.ID, cmds)
			if err != nil {
				log.Fatalf("Cannot create commands on guild %v: %v", guild.ID, err)
			}
		}
	}()
}

// TODO:
func helpCommandLoader(session src.Session) {

}
