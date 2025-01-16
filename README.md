:warning: WARNING: THE `/help` COMMAND STILL IN REWRITE, SO IT DOSENT WORK FOR NOW
# assistbot
Hi, this is my own bot. Its a rewrite from my closed-source Discord.js bot under the same name

The folder system is *kinda react-y*, sorry for that 😔
## Setup

### Install dependecies
run
```shell
go mod download
```
i recommend you to learn each commands that are used in this package, since i will progressively adding more command to it. like "how do they act?", "how they use the deps", "how do they take the args?", and more


### Run the program

just do 
```shell
go run .
```
but its non-blocking tho, beware
or just exit via ctrl+c

## How to add command
First, you need to make a new file inside `command`, like this

```go
package command

// import the root source folder
import (
	"assistbot/src"
)

// this is your command
var Hello = src.Command{
    // just an alias for discordgo.ApplicationCommand
	Info: src.CmdInfo{
		Name:        "hello",
		Description: "you wasted 3 secs to see these",
		Options: src.CmdInfoOpt{
			{
				Name:        "name",
				Description: "whoever the name is",
				Type:        src.CmdInfoOptTypeEnum.String,
				Required:    false,
			},
		},
	},
    // if you want to throw error, i recommend you to just returning it
    // it will automatically converted to string,
    // and send them back to discord
	Fn: func(opt src.CmdResFnArgs) error {
        // this is the result
		return opt.Result(&src.CmdResData{
            //here we say hi
			Content: "Hello!",
		})
	},
}

```