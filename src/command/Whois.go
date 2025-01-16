package command

import (
	"assistbot/src"
	"context"
	"encoding/json"
	"log"

	"github.com/bwmarrin/discordgo"
	"github.com/shlin168/go-whois/whois"
)

var client, clientErr = whois.NewClient()

var WhoisHook src.LoadHook = func(session src.Session) {
	if clientErr != nil {
		log.Fatal(clientErr.Error())
	}
}
var Whois = src.Command{
	Info: src.CmdInfo{
		Name:        "whois",
		Description: "Do a whois lookup",
		Options: src.CmdInfoOpt{
			{
				Name:        "value",
				Description: "The value. Based on the mode you set, or defaults to 'Domain'",
				Type:        src.CmdInfoOptTypeEnum.String,
				Required:    true,
			},
			{
				Name:        "mode",
				Description: "Switch modes between domain, IP, and others",
				Type:        src.CmdInfoOptTypeEnum.Integer,
				Options: []*discordgo.ApplicationCommandOption{
					{
						Name:        "val-type",
						Description: "Value type",
						Type:        src.CmdInfoOptTypeEnum.Integer,
						Choices: []*discordgo.ApplicationCommandOptionChoice{
							{
								Name:  "Domain",
								Value: 1,
							},
							{
								Name:  "IP",
								Value: 2,
							},
							{
								Name:  "Public Suffix",
								Value: 3,
							},
						},
					},
				},
				Required: false,
			},
		},
	},
	Fn: func(opt src.CmdResFnArgs) error {
		ctx := context.Background()
		// client default timeout: 5s,
		// client with custom timeout: whois.NewClient(whois.WithTimeout(10*time.Second))
		value := opt.Args["value"].StringValue()
		var mode int64 = 0
		if modeArg, exist := opt.Args["mode"]; exist {
			mode = modeArg.IntValue()
		} else {
			mode = 1
		}
		res := func(e any) error {
			jsonRes, err := json.MarshalIndent(e, "", "  ")
			if err != nil {
				return err
			}
			embed := &discordgo.MessageEmbed{
				Title:       "Heres the WHOIS result!",
				Description: "```json\n" + string(jsonRes) + "\n```",
				Color:       0xffffff, // Green color
			}
			return opt.Result(&src.CmdResData{
				Embeds: []*discordgo.MessageEmbed{embed},
			})
		}
		switch mode {
		case 1:
			{
				w, err := client.Query(ctx, value, "whois.iana.org")
				if err != nil {
					return err
				}
				w.RawText = ""
				return res(w)
			}
		case 2:
			{
				w, err := client.QueryIP(ctx, value, "whois.iana.org")
				if err != nil {
					return err
				}
				w.RawText = ""
				return res(w)
			}
		case 3:
			{
				w, err := client.QueryPublicSuffix(ctx, value, "whois.iana.org")
				if err != nil {
					return err
				}
				w.RawText = ""
				return res(w)
			}
		default:
			{
				return opt.Result(&src.CmdResData{
					Content: "Can't find the 'whois' based on value",
				})
			}
		}
	},
}
