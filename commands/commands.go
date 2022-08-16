package commands

import (
	"fmt"
	"log"
	"strconv"

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
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "memory-size",
					Description: "Choice memory size",
					Type:        discordgo.ApplicationCommandOptionInteger,
					Required:    true,
					Choices: []*discordgo.ApplicationCommandOptionChoice{
						{
							Name:  "1gb-flavor",
							Value: 1,
						},
						{
							Name:  "2gb-flavor",
							Value: 2,
						},
						{
							Name:  "4gb-flavor",
							Value: 4,
						},
						{
							Name:  "8gb-flavor",
							Value: 8,
						},
						{
							Name:  "16gb-flavor",
							Value: 16,
						},
						{
							Name:  "32gb-flavor",
							Value: 32,
						},
						{
							Name:  "64gb-flavor",
							Value: 64,
						},
					},
				},
			},
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

			//TODO 実行していた時間の利用料金も表示させたい
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Started closing server.\n" +
						"Do not enter any other commands for 5 minute...",
				},
			})

			err := conoha.CloseServer()
			if err != nil {
				log.Fatalf("Failed to stop server: %v", err)
			}
			log.Println("closed server.")

			err = conoha.CleateImage()
			if err != nil {
				log.Fatalf("Failed to create image: %v", err)
			}
			log.Println("saved server image.")

			err = conoha.DeleteServer()
			if err != nil {
				log.Fatalf("Failed to delete server: %v", err)
			}
			log.Println("deleted server")
		},
		"server-open": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			//TODO 既にサーバーが開いているか確認する

			options := i.ApplicationCommandData().Options
			memSize := options[0].IntValue()

			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Started opening server.(memory size " + strconv.FormatInt(memSize, 10) + "GB plan)\n" +
						"Do not enter any other commands for 5 minute...",
				},
			})

			// 保存したイメージIDの取り出し
			imageRef, err := conoha.GetImageRef()
			if err != nil {
				log.Fatalf("Image could not be found: %v", err)
			}

			// 該当プランIDの取り出し
			flavorRef, err := conoha.GetFlavorRef(memSize)
			if err != nil {
				log.Fatalf("Flavor cloud not be found: %v", err)
			}

			// イメージID,プランIDを基にVM作成
			log.Println("Opening server...")
			log.Printf("imageRed:%v\nflavorRef:%v\n", imageRef, flavorRef)
			err = conoha.OpenServer(imageRef, flavorRef)
			if err != nil {
				log.Fatalf("Failed to stop server: %v", err)
			}
			log.Println("Opened server.")

			//TODO IPの取得

			//TODO discordにIPの出力

			//TODO 使用したイメージの削除
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
