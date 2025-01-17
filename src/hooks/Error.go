package hooks

import "assistbot/src"

var Error src.ErrorHook = func(data src.ErrorHookData) {
	if (&src.CmdInfo{} == &data.CmdInfo || data.CmdInfo.Name == "help") {
		src.InteractionRespondRaw(data.Session, data.Interaction, &src.CmdResData{
			Content: "## Error:\n" + data.Message + "\n\n> Check `/help` to find the list of commands",
		})
		return
	}
	src.InteractionRespondRaw(data.Session, data.Interaction, &src.CmdResData{
		Content: "## Error:\n" + data.Message + "\n\n> Check `/help cmd:" + data.CmdInfo.Name + "` for info according to this command",
	})
}
