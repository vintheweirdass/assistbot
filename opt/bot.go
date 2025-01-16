package opt

import (
	"assistbot/src"
	"assistbot/src/command"
	"assistbot/src/hooks"
)

var Commands = []src.Command{
	command.Hello,
}
var Hooks = src.Hooks{
	OnSession: []src.SessionHook{hooks.LoginAnnouncer},
	OnError:   []src.ErrorHook{hooks.Error},
	OnLoad:    nil,
}
