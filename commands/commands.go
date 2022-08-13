package commands

import (
	"fmt"
	"log"
	"time"

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
			//TODO 既にサーバーが閉じられているか確認する
			err := conoha.CloseServer()
			if err != nil {
				log.Fatalf("Failed to stop server: %v", err)
			}

			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Started closing server.\n" +
						"Do not enter any other commands for 5 minute...",
				},
			})

			//TODO 決め打ちで待機してるけど、もっといい方法ありそう...
			time.Sleep(time.Minute * 2)

			//TODO 実行していた時間の利用料金も表示させたい
			err = conoha.CleateImage()
			if err != nil {
				log.Fatalf("Failed to create image: %v", err)
			}

			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Started saving server.",
				},
			})

			//TODO 決め打ちで待機してるけど、もっといい方法ありそう...
			//TODO 排他制御的な仕組みで他のコマンドを受け付けないようにしたい
			time.Sleep(time.Minute * 2)

			//TODO サーバー削除
			err = conoha.DeleteServer()
			if err != nil {
				log.Fatalf("Failed to delete server: %v", err)
			}

			time.Sleep(time.Minute * 2)

			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Closed sercer gracefully.",
				},
			})
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
