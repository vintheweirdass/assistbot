package hooks

import (
	"assistbot/global/env"
	"assistbot/src"
	"assistbot/src/hooks/run"
	"fmt"
	"log"
	"slices"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/dop251/goja"
)

// for `assistbot.getOwners(

func aiMessageCreate(s src.Session, m *discordgo.MessageCreate) {
	if m.Author.Bot || m.ChannelID != env.ChannelForRunJS {
		return
	}

	if strings.HasPrefix(m.Content, "```js") {
		return
	}
	var code string = strings.TrimSuffix(strings.TrimPrefix(m.Content, "```js"), "```")

	// Send initial message to be edited later
	initialMsg, err := s.ChannelMessageSend(m.ChannelID, "🔄 Running your JavaScript code...")
	if err != nil {
		log.Println("Failed to send initial message:", err)
		return
	}

	go func() {
		runtime := goja.New()
		output := make(chan string, 100)
		assistbot := map[string]any{
			"user":              runtime.ToValue(m.Author.Username),
			"originalCode":      runtime.ToValue(code),
			"userId":            runtime.ToValue(m.Author.ID),
			"channelId":         runtime.ToValue(m.ChannelID),
			"messageId":         runtime.ToValue(m.ID),
			"owners":            runtime.ToValue(ownerNames),
			"everyoneMentioned": runtime.ToValue(m.MentionEveryone),
			"rolesMentioned":    runtime.ToValue(m.MentionRoles),
			"usersMentioned": runtime.ToValue((func() []string {
				res := make([]string, len(m.Mentions))
				for i, u := range m.Mentions {
					res[i] = u.Username
				}
				return res
			})()),
			"isOwner": runtime.ToValue(slices.Contains(env.Owners, m.Author.ID)),
		}
		runtime.Set("assistbot", assistbot)
		run.RegisterFunctions(runtime, s, m, initialMsg, output)
		// Collect all console output
		var messages []string
		done := make(chan bool)

		go func() {
			for msg := range output {
				messages = append(messages, msg)
			}
			done <- true
		}()

		_, err := runtime.RunString(code)
		if err != nil {
			output <- fmt.Sprintf("🛑 %v\n", err)
		}

		close(output) // Ensure output is closed before reading

		<-done // Wait for output reader to finish

		finalMessage := "## Console output:\n```\n" + strings.Join(messages, "\n") + "```"

		// Edit the original message with the final result
		_, editErr := s.ChannelMessageEdit(initialMsg.ChannelID, initialMsg.ID, finalMessage)
		if editErr != nil {
			log.Println("Failed to edit message:", editErr)
		}
	}()
}

var AIRegisterer src.LoadHook = func(s src.Session) {
	if !env.EnableRunJS {
		return
	}
	log.Println("-- Adding RunJS (Goja) instance --")
	s.AddHandler(aiMessageCreate)
}
