package main

import (
	"assistbot/opt"
	"assistbot/src"
	"log"

	"github.com/bwmarrin/discordgo"
)

func triggerLoadHook(data src.Session) {
	for _, o := range opt.Hooks.OnLoad {
		go o(data)
	}
}
func triggerErrorHook(data src.ErrorHookData) {
	for _, o := range opt.Hooks.OnError {
		go o(data)
	}
}

func HookLoader(session src.Session) {
	if len(opt.Hooks.OnLoad) > 0 {
		triggerLoadHook(session)
	}
	if len(opt.Hooks.OnSession) > 0 {
		for _, o := range opt.Hooks.OnSession {
			session.AddHandler(func(s src.Session, r *discordgo.Ready) {
				go o(s, r)
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
	session.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		var data = i.ApplicationCommandData()
		if cmd, defined := cmdsObj[data.Name]; defined {
			options := data.Options
			optLen := len(options)
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
					}
				}
			}
			cmd.Fn(src.CmdResFnArgs{
				Session:     s,
				Interaction: i,
				Result: func(e *src.CmdResData) error {
					return src.InteractionRespondRaw(s, i, e)
				},
				ArgsLen: optLen,
			})
		}
	})
	go func() {
		for _, guild := range session.State.Guilds {
			cmdsInfo := make([]*src.CmdInfo, len(opt.Commands))
			for i := range cmdsInfo {
				cmdsInfo[i] = &opt.Commands[i].Info
			}
			_, err := session.ApplicationCommandBulkOverwrite(session.State.User.ID, guild.ID, cmdsInfo)
			if err != nil {
				log.Panicf("Cannot create commands on guild %v: %v", guild.ID, err)
			}
		}
	}()
}
