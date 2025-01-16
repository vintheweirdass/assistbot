package opt

import (
	"assistbot/src"
	"assistbot/src/command"
	"assistbot/src/hooks"
)

// Note: `/help` is programmed directly on botloader.go
// as `func helpCommandLoader(session src.Session)`
var Commands = []src.Command{
	command.Hello, command.Gary, command.Whois,
}
var Hooks = src.Hooks{
	OnSession: []src.SessionHook{hooks.LoginAnnouncer},
	OnError:   []src.ErrorHook{hooks.Error},
	OnLoad:    []src.LoadHook{command.WhoisHook},
}
