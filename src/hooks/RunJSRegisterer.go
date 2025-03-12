package hooks

import (
	"assistbot/global/env"
	"assistbot/src"
	"assistbot/src/hooks/runjs"
	"fmt"
	"log"
	"slices"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/dop251/goja"
)

// for `assistbot.getOwners()`
var ownerNames = []string{}

func runJSMessageCreate(s src.Session, m *discordgo.MessageCreate) {
	if m.Author.Bot || m.ChannelID != env.ChannelForRunJS {
		return
	}

	var code string
	if strings.HasPrefix(m.Content, "!run ```js") {
		code = strings.TrimPrefix(m.Content, "!run ```js")
	} else if strings.HasPrefix(m.Content, "!run\r\n```js") {
		code = strings.TrimPrefix(m.Content, "!run\r\n```js")
	} else if strings.HasPrefix(m.Content, "!run\n```js") {
		code = strings.TrimPrefix(m.Content, "!run\n```js")
	} else {
		return
	}
	code = strings.TrimSuffix(code, "```")

	// Send initial message to be edited later
	initialMsg, err := s.ChannelMessageSend(m.ChannelID, "🔄 Running your JavaScript code...")
	if err != nil {
		log.Println("Failed to send initial message:", err)
		return
	}

	go func() {
		runtime := goja.New()
		var output string

		console := map[string]any{
			"log": func(call goja.FunctionCall) goja.Value {
				args := make([]any, len(call.Arguments))
				for i, arg := range call.Arguments {
					args[i] = arg.Export()
				}
				output += fmt.Sprintln(args...)
				return goja.Undefined()
			},
		}
		runtime.Set("console", console)

		assistbot := map[string]any{
			"user":      runtime.ToValue(m.Author.Username),
			"userId":    runtime.ToValue(m.Author.ID),
			"channelId": runtime.ToValue(m.ChannelID),
			"messageId": runtime.ToValue(m.ID),
			"owners":    runtime.ToValue(ownerNames),
			"isOwner":   runtime.ToValue(slices.Contains(env.Owners, m.Author.ID)),
		}
		runtime.Set("assistbot", assistbot)
		runjs.RegisterFunctions(runtime)

		_, err := runtime.RunString(code)
		if err != nil {
			output += fmt.Sprintf("🛑 %v\n", err)
		}

		finalMessage := fmt.Sprintf("## Console output:\n```\n%s```", output)

		// Edit the original message with the final result
		_, editErr := s.ChannelMessageEdit(initialMsg.ChannelID, initialMsg.ID, finalMessage)
		if editErr != nil {
			log.Println("Failed to edit message:", editErr)
		}
	}()
}

var RunJSRegisterer src.LoadHook = func(s src.Session) {
	if !env.EnableRunJS {
		return
	}
	log.Println("-- Adding RunJS (Goja) instance --")
	s.AddHandler(runJSMessageCreate)
}
var RunJSLoadOwners src.SessionHook = func(s src.Session, r src.SessionReady) {
	if !env.EnableRunJS {
		return
	}
	log.Println("-- Loading owners for RunJS instance --")
	go (func() {
		for _, name := range env.Owners {
			res, e := s.User(name)
			if e != nil {
				break
			}
			ownerNames = append(ownerNames, res.Username)
		}
	})()
}
