package main

import (
	"assistbot/opt"
	"assistbot/src"
	"errors"
	"fmt"
	"log"
	"slices"
	"sync"
	"time"

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
		var data = i.ApplicationCommandData()
		switch i.Type {
		case discordgo.InteractionApplicationCommand:
			{
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
			handleButtonInteraction(s, i)
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

var activeMessages sync.Map

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
	sendHelpEmbed(s, i, 0)
}

func sendHelpEmbed(s *discordgo.Session, i *discordgo.InteractionCreate, page int) {
	const itemsPerPage = 5
	var commands []src.Command = opt.Commands

	totalPages := (len(commands) + itemsPerPage - 1) / itemsPerPage
	if page < 0 {
		page = 0
	} else if page >= totalPages {
		page = totalPages - 1
	}

	startIndex := page * itemsPerPage
	endIndex := startIndex + itemsPerPage
	if endIndex > len(commands) {
		endIndex = len(commands)
	}

	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("Help (Page %d/%d)", page+1, totalPages),
		Description: "List of available commands:",
		Color:       0x00ff00,
		Fields:      []*discordgo.MessageEmbedField{},
	}

	for _, cmd := range commands[startIndex:endIndex] {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:  cmd.Info.Name,
			Value: cmd.Info.Description,
		})
	}

	messageID := fmt.Sprintf("%s_%d", i.ID, time.Now().UnixNano())

	prevButton := discordgo.Button{
		Label:    "Previous",
		Style:    discordgo.PrimaryButton,
		CustomID: fmt.Sprintf("prev_%s_%d", messageID, page),
		Disabled: page == 0,
	}

	nextButton := discordgo.Button{
		Label:    "Next",
		Style:    discordgo.PrimaryButton,
		CustomID: fmt.Sprintf("next_%s_%d", messageID, page),
		Disabled: page == totalPages-1,
	}

	actionRow := discordgo.ActionsRow{
		Components: []discordgo.MessageComponent{prevButton, nextButton},
	}

	var err error
	if i.Type == discordgo.InteractionApplicationCommand {
		err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Embeds:     []*discordgo.MessageEmbed{embed},
				Components: []discordgo.MessageComponent{actionRow},
			},
		})
	} else {
		_, err = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Embeds:     &[]*discordgo.MessageEmbed{embed},
			Components: &[]discordgo.MessageComponent{actionRow},
		})
	}

	if err != nil {
		log.Println("Error sending help embed:", err)
		return
	}

	// Start a timer to expire the buttons after 30 seconds
	go func() {
		timer := time.NewTimer(30 * time.Second)
		<-timer.C
		expireButtons(s, i)
	}()

	// Store the message ID and its expiration time
	activeMessages.Store(messageID, time.Now().Add(30*time.Second))
}

func expireButtons(s *discordgo.Session, i *discordgo.InteractionCreate) {
	_, err := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Components: &[]discordgo.MessageComponent{},
	})

	if err != nil {
		log.Println("Error expiring buttons:", err)
	}
}

// Add this function to handle button interactions
func handleButtonInteraction(s *discordgo.Session, i *discordgo.InteractionCreate) {
	data := i.MessageComponentData()

	var action, messageID string
	var page int
	fmt.Sscanf(data.CustomID, "%s_%s_%d", &action, &messageID, &page)

	// Check if the message is still active
	expTime, ok := activeMessages.Load(messageID)
	if !ok || time.Now().After(expTime.(time.Time)) {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "This message has expired. Please use the /help command again.",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}

	switch action {
	case "prev":
		page--
	case "next":
		page++
	}

	// Reset the expiration timer
	activeMessages.Store(messageID, time.Now().Add(30*time.Second))

	// Update the message with new content and reset the timer
	sendHelpEmbed(s, i, page)
}
