package main

import (
	"fmt"
	"log"
	"os"

	"github.com/bwmarrin/discordgo"
)

func main() {
	ok := cmd.Ok
	dcToken, dcTokenValid := os.LookupEnv("ASSISTBOT_DISCORD_TOKEN")
	if dcTokenValid {
		log.Fatal("Discord token dosent found")
	}
	discord, err := discordgo.New("Bot " + dcToken)
	if err != nil {
		log.Fatal(err)
	}
	err = discord.Open()
	if err != nil {
		fmt.Println("Error opening connection,", err)
		return
	}
	discord.ApplicationCommand()
}
