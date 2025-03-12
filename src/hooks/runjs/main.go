package runjs

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/dop251/goja"
)

func NewFuture(vm *goja.Runtime) (*goja.Object, chan goja.Value) {
	resultChan := make(chan goja.Value, 1)

	// Create a Future-like object
	future := vm.NewObject()
	future.Set("wait", func(call goja.FunctionCall) goja.Value {
		return <-resultChan
	})

	return future, resultChan
}

// createAsync defines async(fn) -> returns a function that auto-runs and returns a Promise
func createAsync(vm *goja.Runtime) goja.Value {
	return vm.ToValue(func(call goja.FunctionCall) goja.Value {
		fn, isFunc := goja.AssertFunction(call.Argument(0))
		if !isFunc {
			return vm.NewTypeError("async() expects a function")
		}

		// Return a function that executes immediately and returns a Future
		return vm.ToValue(func(innerCall goja.FunctionCall) goja.Value {
			resultChan := make(chan goja.Value, 1)
			errorChan := make(chan goja.Value, 1)

			go func() {
				defer func() {
					if r := recover(); r != nil {
						errorChan <- vm.NewGoError(r.(error))
					}
				}()

				// Call function
				result, err := fn(goja.Undefined(), innerCall.Arguments...)
				if err != nil {
					errorChan <- NewPromiseError(vm, err.Error())
					return
				}

				resultChan <- result
			}()

			// Create a Future-like object
			future := vm.NewObject()
			future.Set("wait", func(call goja.FunctionCall) goja.Value {
				select {
				case result := <-resultChan:
					return result
				case err := <-errorChan:
					panic(err) // Ensure error gets propagated
				}
			})

			return future
		})
	})
}

// createAwait implements await(promise) -> waits for resolution or throws error
func createAwait(vm *goja.Runtime) goja.Value {
	return vm.ToValue(func(call goja.FunctionCall) goja.Value {
		promiseObj := call.Argument(0).ToObject(vm)
		waitFunc, ok := goja.AssertFunction(promiseObj.Get("wait"))
		if !ok {
			return vm.NewTypeError("await() expects a valid Future object")
		}

		// Call wait() and return result
		result, err := waitFunc(goja.Undefined())
		if err != nil {
			panic(err) // Ensure error is properly thrown
		}
		return result
	})
}

// createSleep implements sleep(ms) -> returns a Future that resolves after ms milliseconds
func createSleep(vm *goja.Runtime) goja.Value {
	return vm.ToValue(func(call goja.FunctionCall) goja.Value {
		ms := call.Argument(0).ToInteger()
		resultChan := make(chan goja.Value, 1)

		go func() {
			time.Sleep(time.Duration(ms) * time.Millisecond)
			resultChan <- vm.ToValue(nil)
		}()

		// Create Future-like object
		future := vm.NewObject()
		future.Set("wait", func(call goja.FunctionCall) goja.Value {
			return <-resultChan
		})

		return future
	})
}

// DeepEqual checks if two Goja values are the same by structure (not reference)
func DeepEqual(vm *goja.Runtime, a, b goja.Value) bool {
	// Convert both values to JSON
	jsonA, errA := json.Marshal(a.Export())
	jsonB, errB := json.Marshal(b.Export())

	// If any conversion fails, return false
	if errA != nil || errB != nil {
		return false
	}

	// Compare JSON strings
	return string(jsonA) == string(jsonB)
}

// assert(a, b) function
func createAssert(vm *goja.Runtime) goja.Value {
	return vm.ToValue(func(call goja.FunctionCall) goja.Value {
		a := call.Argument(0)
		b := call.Argument(1)

		// Check if values are literally the same
		if !DeepEqual(vm, a, b) {
			return NewAssertError(vm, fmt.Sprintf("Values are not equal\nA: %v\nB: %v", a, b))
		}

		// If equal, return undefined
		return goja.Undefined()
	})
}
func createConsole(_vm *goja.Runtime, output chan string) map[string]any {
	return map[string]any{
		"log": func(call goja.FunctionCall) goja.Value {
			args := make([]any, len(call.Arguments))
			for i, arg := range call.Arguments {
				args[i] = arg.Export()
			}
			output <- fmt.Sprintln(args...)
			return goja.Undefined()
		},
		"info": func(call goja.FunctionCall) goja.Value {
			args := make([]any, len(call.Arguments))
			for i, arg := range call.Arguments {
				args[i] = arg.Export()
			}
			var sb strings.Builder
			sb.WriteString("🔔: ")
			sb.WriteString(fmt.Sprintln(args...))
			output <- sb.String()
			return goja.Undefined()
		},
		"warn": func(call goja.FunctionCall) goja.Value {
			args := make([]any, len(call.Arguments))
			for i, arg := range call.Arguments {
				args[i] = arg.Export()
			}
			var sb strings.Builder
			sb.WriteString("⚠️: ")
			sb.WriteString(fmt.Sprintln(args...))
			output <- sb.String()
			return goja.Undefined()
		},
		"error": func(call goja.FunctionCall) goja.Value {
			args := make([]any, len(call.Arguments))
			for i, arg := range call.Arguments {
				args[i] = arg.Export()
			}
			var sb strings.Builder
			sb.WriteString("🛑: ")
			sb.WriteString(fmt.Sprintln(args...))
			output <- sb.String()
			return goja.Undefined()
		},
	}
}

func UpdateMessage(s *discordgo.Session, channelID, messageID, newContent string, components []discordgo.MessageComponent) error {
	_, err := s.ChannelMessageEditComplex(&discordgo.MessageEdit{
		ID:         messageID,
		Channel:    channelID,
		Content:    &newContent,
		Components: &components,
	})
	return err
}

func createAlert(vm *goja.Runtime, s *discordgo.Session, m *discordgo.MessageCreate) goja.Value {
	return vm.ToValue(func(call goja.FunctionCall) goja.Value {
		message := call.Argument(0).String()

		future, resultChan := NewFuture(vm)

		err := UpdateMessage(s, m.ChannelID, m.Message.ID, fmt.Sprintf("📢 **Alert:** %s", message),
			[]discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.Button{
							Label:    "OK",
							Style:    discordgo.PrimaryButton,
							CustomID: "alert_ok",
						},
					},
				},
			},
		)
		if err != nil {
			fmt.Println("Failed to update alert:", err)
			return goja.Undefined()
		}

		// Button interaction handler
		s.AddHandlerOnce(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			if i.Message.ID != m.Message.ID || i.MessageComponentData().CustomID != "alert_ok" {
				return
			}

			_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseUpdateMessage,
				Data: &discordgo.InteractionResponseData{
					Content:    "📢 **Alert dismissed!**. 🔄 Running your JavaScript code...",
					Components: []discordgo.MessageComponent{},
				},
			})
			resultChan <- goja.Undefined()
		})

		return future
	})
}

func createConfirm(vm *goja.Runtime, s *discordgo.Session, m *discordgo.MessageCreate) goja.Value {
	return vm.ToValue(func(call goja.FunctionCall) goja.Value {
		message := call.Argument(0).String()

		future, resultChan := NewFuture(vm)

		err := UpdateMessage(s, m.ChannelID, m.Message.ID, fmt.Sprintf("❓ **Confirm:** %s", message),
			[]discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.Button{
							Label:    "Yes",
							Style:    discordgo.SuccessButton,
							CustomID: "confirm_yes",
						},
						discordgo.Button{
							Label:    "No",
							Style:    discordgo.DangerButton,
							CustomID: "confirm_no",
						},
					},
				},
			},
		)
		if err != nil {
			fmt.Println("Failed to update confirm:", err)
			return goja.Undefined()
		}

		// Interaction handler
		s.AddHandlerOnce(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			if i.Message.ID != m.Message.ID {
				return
			}

			var result bool
			switch i.MessageComponentData().CustomID {
			case "confirm_yes":
				result = true
			case "confirm_no":
				result = false
			}

			_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseUpdateMessage,
				Data: &discordgo.InteractionResponseData{
					Content:    fmt.Sprintf("✅ You selected: **%t**. 🔄 Running your JavaScript code...", result),
					Components: []discordgo.MessageComponent{},
				},
			})

			resultChan <- vm.ToValue(result)
		})

		return future
	})
}

func createPrompt(vm *goja.Runtime, s *discordgo.Session, m *discordgo.MessageCreate) goja.Value {
	messageID := m.Message.ID
	channelID := m.ChannelID
	userID := m.Author.ID
	return vm.ToValue(func(call goja.FunctionCall) goja.Value {
		message := call.Argument(0).String()

		future, resultChan := NewFuture(vm)

		err := UpdateMessage(s, channelID, messageID, fmt.Sprintf("⌨ **Prompt:** %s\nPlease type your response below.", message),
			[]discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.Button{
							Label:    "Cancel",
							Style:    discordgo.DangerButton,
							CustomID: "prompt_cancel",
						},
					},
				},
			},
		)
		if err != nil {
			fmt.Println("Failed to update prompt:", err)
			return goja.Undefined()
		}

		// Wait for user input
		s.AddHandlerOnce(func(s *discordgo.Session, m *discordgo.MessageCreate) {
			if m.Author.ID != userID || m.ChannelID != channelID {
				return
			}

			// Edit message with the user's response
			UpdateMessage(s, channelID, messageID, "⌨ **Prompt responded!**. 🔄 Running your JavaScript code...", nil)

			resultChan <- vm.ToValue(m.Content)
		})

		// Handle cancel button
		s.AddHandlerOnce(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			if i.Message.ID != messageID || i.MessageComponentData().CustomID != "prompt_cancel" {
				return
			}

			UpdateMessage(s, channelID, messageID, "⌨ **Prompt canceled!**", nil)
			resultChan <- goja.Undefined()
		})

		return future
	})
}

// Register Goja functions
func RegisterFunctions(vm *goja.Runtime, s *discordgo.Session, m *discordgo.MessageCreate, output chan string) {
	vm.Set("async", createAsync(vm))
	vm.Set("await", createAwait(vm))
	vm.Set("sleep", createSleep(vm))
	vm.Set("assert", createAssert(vm))
	vm.Set("console", createConsole(vm, output))
	vm.Set("alert", createAlert(vm, s, m))
	vm.Set("confirm", createConfirm(vm, s, m))
	vm.Set("prompt", createPrompt(vm, s, m))
}
