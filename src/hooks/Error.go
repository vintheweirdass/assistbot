package hooks

import "assistbot/src"

var Error src.ErrorHook = func(data src.ErrorHookData) {
	src.InteractionRespondRaw(data.Session, data.Interaction, &src.CmdResData{
		Content: "## Error:\n" + data.Message + "\n\nCheck `/help " + data.CmdInfo.Name + "` for info according to this command",
	})
}
