package main

import (
	"assistbot/global/env"
	"log"
	"os"
	"os/signal"

	"github.com/bwmarrin/discordgo"
)

func main() {
	var token = env.DiscordToken
	if token == "" {
		log.Fatal("Discord token dosent found")
		return
	}
	discord, err := discordgo.New("Bot " + token)
	if err != nil {
		log.Fatal(err)
		return
	}
	HookLoader(discord)
	err = discord.Open()
	if err != nil {
		log.Fatal("Error opening connection: ", err)
		return
	}
	CommandLoader(discord)
	defer discord.Close()
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	log.Println("Press Ctrl+C to exit")
	<-stop
}
