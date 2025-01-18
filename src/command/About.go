package command

import (
	"assistbot/global"
	"assistbot/src"

	"github.com/bwmarrin/discordgo"
)

var About = src.Command{
	Info: src.CmdInfo{
		Name:        "about",
		Description: "About this bot",
	},
	Fn: func(opt src.CmdResFnArgs) error {
		embed := &discordgo.MessageEmbed{
			Image: &discordgo.MessageEmbedImage{
				URL: global.LogoDataUrl,
			},
			Title:       "About this bot",
			Description: "Just an assistant that helps to add the missing features on most discord bots nowadays. Inspired by Stef's bot",
			Color:       0x00ff00, // Green color
		}

		actionRow := discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Label: "Source Code (Go)",
					Style: discordgo.LinkButton,
					URL:   "https://github.com/vintheweirdass/assistbot",
				},
				discordgo.Button{
					Label: "vintheweirdass' bio",
					Style: discordgo.LinkButton,
					URL:   "https://vtwa.is-a.dev",
				},
				discordgo.Button{
					Label: "Stef's bot",
					Style: discordgo.LinkButton,
					URL:   "https://github.com/Stef-00012/userApps",
				},
			},
		}

		return opt.Result(&src.CmdResData{
			Embeds:     []*discordgo.MessageEmbed{embed},
			Components: []discordgo.MessageComponent{actionRow},
		})
	},
}
