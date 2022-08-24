package main

import (
	"log"
	"os"
	"os/signal"

	"github.com/TOGEP/ConohaChatOps/commands"
	"github.com/TOGEP/ConohaChatOps/conoha"
	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

// Bot parameters
var guildID string

var (
	commandList     = commands.Commands
	commandHandlers = commands.CommandHandlers
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println(".env file not find.")
	}

	discord, err := discordgo.New("Bot " + os.Getenv("BOTTOKEN"))
	if err != nil {
		log.Fatalf("Invalid BOTTOKEN: %v", err)
		return
	}

	bot, err := conoha.NewBot(discord)
	if err != nil {
		log.Fatalf("Failed token publish. Please check env file.: %v", err)
	}

	// discordReady
	discord.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		commands.CreateCommands(bot.Session, os.Getenv("GUILDID"))
	})

	discord.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if h, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
			h(bot, i)
		}
	})

	err = discord.Open()
	if err != nil {
		log.Fatalf("Cannot open session: %v", err)
		return
	}

	defer discord.Close()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	log.Println("Press Ctrl+C to exit")
	<-stop

	commands.DeleteCommands(discord, os.Getenv("GUILDID"))

	log.Println("Successfully shut down")
	return
}
