package commands

import (
	"fmt"
	"log"

	"github.com/TOGEP/ConohaChatOps/conoha"
	"github.com/bwmarrin/discordgo"
)

var (
	Commands = []*discordgo.ApplicationCommand{
		{
			Name:        "server-help",
			Description: "Help command",
		},
		{
			Name:        "server-open",
			Description: "Open server",
		},
		{
			Name:        "server-close",
			Description: "Close server",
		},
	}

	CommandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"server-help": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Hi there, I am a bot built on slash commands!\n" +
						"\nOpen the server with `/server-open`." +
						"\nClose the server with `server-close`.",
				},
			})
		},
		"server-close": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			err := conoha.CloseServer()
			if err != nil {
				log.Fatalf("Failed to stop server: %v", err)
			}

			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Stopped the server!",
				},
			})

			//TODO 停止したサーバーをイメージ保存後、サーバー削除
		},
	}
)

func CreateCommands(s *discordgo.Session, g string) {
	//TODO duplicated check
	registeredCommands := make([]*discordgo.ApplicationCommand, len(Commands))
	for i, v := range Commands {
		cmd, err := s.ApplicationCommandCreate(s.State.User.ID, g, v)
		if err != nil {
			log.Panicf("Cannot create '%v' command: %v", v.Name, err)
		}
		registeredCommands[i] = cmd
	}
	log.Println("Commands started")
}

func DeleteCommands(s *discordgo.Session, g string) {
	fmt.Println("Removing commands...")
	for _, v := range Commands {
		err := s.ApplicationCommandDelete(s.State.User.ID, g, v.ID)
		if err != nil {
			log.Fatalf("Cannot delete '%v' command: %v", v.Name, err)
		}
	}
}
