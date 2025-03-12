package hooks

import (
	"assistbot/global/env"
	"assistbot/src"
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

	resultChan := make(chan string)
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
			"info": func(call goja.FunctionCall) goja.Value {
				args := make([]any, len(call.Arguments))
				for i, arg := range call.Arguments {
					args[i] = arg.Export()
				}
				var sb strings.Builder
				// we use emoji for infos. just ignore ur IDE highlights
				sb.WriteString("ℹ️: ")
				sb.WriteString(fmt.Sprintln(args...))
				output += sb.String()
				return goja.Undefined()
			},
			"warn": func(call goja.FunctionCall) goja.Value {
				args := make([]any, len(call.Arguments))
				for i, arg := range call.Arguments {
					args[i] = arg.Export()
				}
				var sb strings.Builder
				// we use emoji for infos. just ignore ur IDE highlights
				sb.WriteString("⚠️: ")
				sb.WriteString(fmt.Sprintln(args...))
				output += sb.String()
				return goja.Undefined()
			},
			"error": func(call goja.FunctionCall) goja.Value {
				args := make([]any, len(call.Arguments))
				for i, arg := range call.Arguments {
					args[i] = arg.Export()
				}
				var sb strings.Builder
				// we use emoji for infos. just ignore ur IDE highlights
				sb.WriteString("🛑: ")
				sb.WriteString(fmt.Sprintln(args...))
				output += sb.String()
				return goja.Undefined()
			},
			"debug": func(call goja.FunctionCall) goja.Value {
				output += "⚠️: console.debug is not supported\n"
				return goja.Undefined()
			},
		}
		runtime.Set("console", console)
		assistbot := map[string]any{
			"user":      runtime.ToValue(m.Author.Username),
			"userId":    runtime.ToValue(m.Author.ID),
			"channelId": runtime.ToValue(m.ChannelID),
			"messageId": runtime.ToValue(m.ID),
			"timestamp": runtime.ToValue(m.Timestamp.String()),
			"owners":    runtime.ToValue(ownerNames),
			"isOwner":   runtime.ToValue(slices.Contains(env.Owners, m.Author.Username)),
		}
		runtime.Set("assistbot", assistbot)

		runtime.Set("require", func(call goja.FunctionCall) goja.Value {
			output += "⚠️: require is not supported\n"
			return goja.Undefined()
		})

		_, err := runtime.RunString(code)
		if err != nil {
			output += fmt.Sprintf("Error: %v\n", err)
		}

		resultChan <- fmt.Sprintf("## Console output:\n```\n%s```", output)
		close(resultChan)
	}()

	for result := range resultChan {
		s.ChannelMessageSendReply(m.ChannelID, result, m.Reference())
	}

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
