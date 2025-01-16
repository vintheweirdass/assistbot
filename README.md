:warning: WARNING: THE `/help` COMMAND IS STILL IN REWRITE, SO IT DOSENT WORK FOR NOW
# assistbot
Hi, this is my own bot. Its a rewrite from my closed-source Discord.js bot under the same name

The folder system is *kinda react-y*, sorry for that 😔
## Setup
### Before these
Add `ASSISTBOT_DISCORD_TOKEN` to your Environment, and set it to your Discord Bot Token (https://discord.com/developers)

### Install dependecies
run
```shell
go mod download
```
i recommend you to learn each commands that are used in this package, since i will progressively adding more commands to it. like "how do they act?", "how they use the deps", "how do they take the args?", and more


### Run the program

just do 
```shell
go run .
```
but its non-blocking tho, beware

or just exit via ctrl+c

## How to add command
First, you need to make a new file inside `src/command` and name it (e.g. `hi.go`). After that, write it like this

```go
package command

// import the root source folder
import (
	"assistbot/src"
)

// this is your command
var Hi = src.Command{
    // just an alias for discordgo.ApplicationCommand
	Info: src.CmdInfo{
		Name:        "hi",
		Description: "Make the bot says hi!",
	},
    // if you want to throw error, i recommend you to just returning it
    // it will automatically converted to string,
    // and send them back to discord
	Fn: func(opt src.CmdResFnArgs) error {
        // this is the result
		return opt.Result(&src.CmdResData{
            //here we say hi
			Content: "Hi!",
		})
	},
}

```
After that, go to `opt` and find `bot.go`

Edit the file and add one of your command to `Commands`
```go
...
var Commands = []src.Command{
	command.Hello, command.Gary, command.Whois, command.Hi, //your command here
}
...
```
Mostly you can understand it by (again and again) inspecting the code. If you are a new to Go, you can click [this link](https://go.dev/doc/tutorial/getting-started) from the official Go website, to get started
