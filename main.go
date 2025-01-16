package main

import (
	"fmt"
	"log"
	"os"

	"github.com/bwmarrin/discordgo"
)

func main() {
	dcToken, dcTokenExist := os.LookupEnv("ASSISTBOT_DISCORD_TOKEN")
	if dcTokenExist {
		log.Fatal("Discord token dosent found")
		return
	}
	discord, err := discordgo.New("Bot " + dcToken)
	if err != nil {
		log.Fatal(err)
		return
	}
	CommandLoader(discord)
	HookLoader(discord)
	err = discord.Open()
	if err != nil {
		fmt.Print("Error opening connection,", err)
		return
	}
}
