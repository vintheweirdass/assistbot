package src

import "github.com/bwmarrin/discordgo"

type CmdInfo = discordgo.ApplicationCommand
type CmdInfoOpt = []*discordgo.ApplicationCommandOption
type Session = *discordgo.Session
type SessionReady = *discordgo.Ready
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

type ACIDO = *discordgo.ApplicationCommandInteractionDataOption

// return value (arguments, is OG arguments length small)
func GetInteractionArgs(i CmdIntr) (map[string]ACIDO, int) {
	options := i.ApplicationCommandData().Options
	optLen := len(options)
	optionMap := make(map[string]ACIDO, len(options))

	if optLen < 1 {
		return optionMap, 0
	}

	for _, opt := range options {
		optionMap[opt.Name] = opt
	}
	return optionMap, optLen
}

type CmdResData = discordgo.InteractionResponseData
type CmdResFn = func(data *CmdResData) error
type CmdResFnArgs struct {
	Session     Session
	Interaction CmdIntr
	Result      CmdResFn
	Args        map[string]ACIDO
	ArgsLen     int
}

func InteractionRespondRaw(s Session, i *discordgo.InteractionCreate, data *discordgo.InteractionResponseData) error {
	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: data,
	})
}

type Command struct {
	Info CmdInfo
	Fn   func(args CmdResFnArgs) error
	// Category string
}
type Commands = map[string]Command

type ErrorHookData struct {
	CmdInfo     CmdInfo
	Message     string
	Interaction *discordgo.InteractionCreate
	Session     Session
}
type SessionHook = func(session Session, r *discordgo.Ready)
type LoadHook = func(session Session)
type ErrorHook = func(data ErrorHookData)
type Hooks struct {
	OnSession []SessionHook
	OnError   []ErrorHook
	OnLoad    []LoadHook
}
