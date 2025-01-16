package main

import (
	"log"
	"os"
	"os/signal"

	"github.com/bwmarrin/discordgo"
)

func main() {
	dcToken := os.Getenv("ASSISTBOT_DISCORD_TOKEN")
	if dcToken == "" {
		log.Fatal("Discord token dosent found")
		return
	}
	discord, err := discordgo.New("Bot " + dcToken)
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
