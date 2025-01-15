package src

import "github.com/bwmarrin/discordgo"

type CmdInfo = discordgo.ApplicationCommand
type CmdInfoOpt = []*discordgo.ApplicationCommandOption
type Session = *discordgo.Session
type CmdIntr = *discordgo.InteractionCreate
type CmdInfoOptType = discordgo.ApplicationCommandOptionType
type tCmdInfoOptType struct {
	SubCommand      CmdInfoOptType
	SubCommandGroup CmdInfoOptType
	String          CmdInfoOptType
	Integer         CmdInfoOptType
	Boolean         CmdInfoOptType
	User            CmdInfoOptType
	Channel         CmdInfoOptType
	Role            CmdInfoOptType
	Mentionable     CmdInfoOptType
	Number          CmdInfoOptType
	Attachment      CmdInfoOptType
}

var CmdInfoOptTypeEnum = tCmdInfoOptType{
	SubCommand:      1,
	SubCommandGroup: 2,
	String:          3,
	Integer:         4,
	Boolean:         5,
	User:            6,
	Channel:         7,
	Role:            8,
	Mentionable:     9,
	Number:          10,
	Attachment:      11,
}

type acido = *discordgo.ApplicationCommandInteractionDataOption

// return value (arguments, is OG arguments length small)
func GetInteractionArgs(i CmdIntr) (map[string]acido, int) {
	options := i.ApplicationCommandData().Options
	optLen := len(options)
	optionMap := make(map[string]acido, len(options))

	if optLen < 1 {
		return optionMap, 0
	}

	for _, opt := range options {
		optionMap[opt.Name] = opt
	}
	return optionMap, optLen
}

type CmdResData = discordgo.InteractionResponseData
type CmdResFn = func(data *discordgo.InteractionResponseData) error

func InteractionRespondRaw(s *discordgo.Session, i *discordgo.InteractionCreate, data *discordgo.InteractionResponseData) error {
	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: data,
	})
}

type Command struct {
	Info *CmdInfo
	Fn   func(session Session, intr CmdIntr, res CmdResFn)
}
type Hook = func(session Session)
