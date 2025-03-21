package opt

import (
	"assistbot/src"
	"assistbot/src/command"
	"assistbot/src/hooks"
)

// Note: `/help` is programmed directly on botloader.go
// as `func helpCommandLoader`
var Commands = []src.Command{
	command.Hello, command.Gary, command.Whois, command.RandomImage, command.FunFact, command.LME, command.RefreshLME, command.About,
}
var Hooks = src.Hooks{
	OnSession: []src.SessionHook{hooks.LoginAnnouncer, hooks.RunJSLoadOwners},
	OnError:   []src.ErrorHook{hooks.Error},
	OnLoad:    []src.LoadHook{hooks.RunJSRegisterer, hooks.TempWebpage, command.WhoisHook, command.LMEHook},
}
