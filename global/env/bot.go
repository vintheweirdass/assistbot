package env

import "os"

var DiscordToken = os.Getenv("ASSISTBOT_DISCORD_TOKEN")
var Owners = []string{"999863217557880842"}

// useful if you want to host to koyeb or whatsoever
var EnableTempWebserver = false

var Port = os.Getenv("PORT")

// var Port = "8000"

var EnableRunJS = true

// runjs currenly only support one channel
var ChannelForRunJS = "1348657323844833431"
