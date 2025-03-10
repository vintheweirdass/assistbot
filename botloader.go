package main

import (
	"assistbot/opt"
	"assistbot/src"
	"errors"
	"fmt"
	"log"
	"slices"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
)

const maxCmdShown = 5

var cmdPages [][]src.CmdInfo = make([][]src.CmdInfo, int(len(opt.Commands)/maxCmdShown)+1)

func triggerInternalLoadHook(_ src.Session) {
	const mcs = maxCmdShown - 1
	var idx = 0
	var count = 0
	for _, o := range opt.Commands {
		if count > mcs {
			idx++
			count = 0
		}
		if cmdPages[idx] == nil {
			cmdPages[idx] = make([]src.CmdInfo, maxCmdShown)
		}
		cmdPages[idx] = append(cmdPages[idx], o.Info)
		count++
	}
}
func triggerLoadHook(data src.Session) {
	for _, o := range opt.Hooks.OnLoad {
		o(data)
	}
	triggerInternalLoadHook(data)
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
var helpInfo = src.CmdInfo{
	Name:        "help",
	Description: "get list of commands in here",
	Options: src.CmdInfoOpt{
		{
			Name:        "cmd",
			Description: "The command name",
			Type:        src.CmdInfoOptTypeEnum.String,
			Required:    false,
		},
		{
			Name:        "arg",
			Description: "The argument (from command)",
			Type:        src.CmdInfoOptTypeEnum.String,
			Required:    false,
		},
	},
}

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
		switch i.Type {
		case discordgo.InteractionApplicationCommand:
			{
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
					return
				} else if data.Name == "help" {
					helpCommandLoader(s, i, data)
					return
				}
				triggerErrorHook(src.ErrorHookData{
					CmdInfo:     src.CmdInfo{},
					Message:     "command `" + data.Name + "` dosent found",
					Interaction: i,
					Session:     s,
				})
			}
		case discordgo.InteractionMessageComponent:
			handleButtonInteraction(s, i.Interaction)
		}
	})
	go func() {
		for _, guild := range session.State.Guilds {
			var cmds = make([]*src.CmdInfo, len(opt.Commands))
			for i, cmd := range opt.Commands {
				cmds[i] = &cmd.Info
			}
			_, err := session.ApplicationCommandBulkOverwrite(session.State.User.ID, guild.ID, append(cmds, &helpInfo))
			if err != nil {
				log.Fatalf("Cannot create commands on guild %v: %v", guild.ID, err)
			}
		}
	}()
}

func helpCommandLoader(s *discordgo.Session, i *discordgo.InteractionCreate, data discordgo.ApplicationCommandInteractionData) {
	var errorResult = func(t src.CmdInfo, err error) {
		if err != nil {
			triggerErrorHook(src.ErrorHookData{
				CmdInfo:     t,
				Message:     err.Error(),
				Interaction: i,
				Session:     s,
			})
		}
	}

	var result = func(e *src.CmdResData) error {
		return src.InteractionRespondRaw(s, i, e)
	}

	type acid = *discordgo.ApplicationCommandInteractionDataOption
	var opt = data.Options
	var cmdIdx = slices.IndexFunc(opt, func(e acid) bool {
		return e.Name == "cmd"
	})
	var argIdx = slices.IndexFunc(opt, func(e acid) bool {
		return e.Name == "arg"
	})

	if cmdIdx < 0 && argIdx > -1 {
		errorResult(helpInfo, errors.New("if you want to use the `arg`, you need to set the command too (`cmd`)"))
		return
	}

	if cmdIdx > -1 {
		var cmdFind = opt[cmdIdx]
		var cmdName = cmdFind.StringValue()
		var cmd, exists = cmdsObj[cmdName]
		if !exists {
			errorResult(helpInfo, errors.New("command `"+cmdName+"` wasn't found"))
			return
		}
		if argIdx > -1 {
			var argFind = opt[argIdx]
			var argName = argFind.StringValue()
			var arg, exists = cmdsObjArgs[cmdName][argName]
			if !exists {
				errorResult(helpInfo, errors.New("argument `"+argName+"` inside command `"+cmdName+"` wasn't found"))
				return
			}
			var requiredText = ""
			if arg.Required {
				requiredText = "**(REQUIRED)**"
			} else {
				requiredText = ""
			}
			var embed = &discordgo.MessageEmbed{
				Title:       fmt.Sprintf("Help: %s", cmd.Info.Name),
				Description: requiredText + " " + arg.Description,
				Color:       0x00ff00,
				Fields:      []*discordgo.MessageEmbedField{},
			}
			result(&src.CmdResData{
				Embeds: []*discordgo.MessageEmbed{embed},
			})
		}
		var embed = &discordgo.MessageEmbed{
			Title:       fmt.Sprintf("Help: %s", cmd.Info.Name),
			Description: cmd.Info.Description,
			Color:       0x00ff00,
			Fields:      []*discordgo.MessageEmbedField{},
		}

		if len(cmd.Info.Options) > 0 {
			var optionsField = &discordgo.MessageEmbedField{
				Name:  "Options",
				Value: "",
			}
			for _, opt := range cmd.Info.Options {
				var requiredText = ""
				if opt.Required {
					requiredText = "**(REQUIRED)** "
				} else {
					requiredText = ""
				}
				optionsField.Value += fmt.Sprintf("**%s**: %s%s\n", opt.Name, requiredText, opt.Description)
			}
			embed.Fields = append(embed.Fields, optionsField)
		}

		result(&src.CmdResData{
			Embeds: []*discordgo.MessageEmbed{embed},
		})
		return
	}

	// Generate paginated help for all commands
	print("s")
	sendHelpEmbed(s, i.Interaction, 0)
}

func sendHelpEmbed(s *discordgo.Session, i *discordgo.Interaction, page int) {
	messageID := i.ID

	totalPages := len(cmdPages)
	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("Help (Page %d/%d)", page+1, totalPages),
		Description: "List of available commands:",
		Color:       0x00ff00,
		Fields:      []*discordgo.MessageEmbedField{},
	}

	for _, info := range cmdPages[page] {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:  info.Name,
			Value: info.Description,
		})
	}

	prevButton := discordgo.Button{
		Label:    "Previous",
		Style:    discordgo.SecondaryButton,
		CustomID: fmt.Sprintf("asbt--helpbtn%s_%d", messageID, page-1),
		Disabled: page == 0,
	}

	nextButton := discordgo.Button{
		Label:    "Next",
		Style:    discordgo.PrimaryButton,
		CustomID: fmt.Sprintf("asbt--helpbtn%s_%d", messageID, page+1),
		Disabled: page == totalPages-1,
	}

	actionRow := discordgo.ActionsRow{
		Components: []discordgo.MessageComponent{prevButton, nextButton},
	}

	// var act = activeMessages[messageID]
	var err error
	if i.Type == discordgo.InteractionApplicationCommand {
		err = s.InteractionRespond(i, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Embeds:     []*discordgo.MessageEmbed{embed},
				Components: []discordgo.MessageComponent{actionRow},
			},
		})
	} else {
		err = s.InteractionRespond(i, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseUpdateMessage,
			Data: &discordgo.InteractionResponseData{
				Embeds:     []*discordgo.MessageEmbed{embed},
				Components: []discordgo.MessageComponent{actionRow},
			},
		})
	}

	if err != nil {
		log.Printf("Error sending help embed: %v", err)
	}
}

func handleButtonInteraction(s *discordgo.Session, i *discordgo.Interaction) {
	data := i.MessageComponentData()
	if !strings.HasPrefix(data.CustomID, "asbt--helpbtn") {
		return
	}
	page, _ := strconv.Atoi(strings.Split(data.CustomID, "_")[1])
	sendHelpEmbed(s, i, page)
}
