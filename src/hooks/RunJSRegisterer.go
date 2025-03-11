package hooks

import (
	"assistbot/global/env"
	"assistbot/src"
	"fmt"
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/dop251/goja"
)

func runJSMessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.Bot || m.ChannelID != env.ChannelForRunJS {
		return
	}

	if strings.HasPrefix(m.Content, "!run ```js") {
		code := strings.TrimPrefix(m.Content, "!run ```js")
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
			}
			runtime.Set("console", console)
			asbt := map[string]any{
				"user": func(call goja.FunctionCall) goja.Value {
					return runtime.ToValue(m.Author.Username)
				},
				"messageId": func(call goja.FunctionCall) goja.Value {
					return runtime.ToValue(m.ID)
				},
			}
			runtime.Set("asbt", asbt)

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
}

var RunJSRegisterer src.LoadHook = func(s src.Session) {
	if !env.EnableRunJS {
		return
	}
	log.Println("-- Adding RunJS (Goja) instance --")
	s.AddHandler(runJSMessageCreate)
}
