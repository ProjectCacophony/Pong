package main

import (
	"flag"
	"os"

	"fmt"

	"time"

	"strconv"
	"strings"

	"errors"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/bwmarrin/discordgo"
	"github.com/json-iterator/go"
	"gitlab.com/project-d-collab/dhelpers"
)

var (
	Token    string
	Dg       *discordgo.Session
	InitAt   time.Time
	HandleAt time.Time
)

func init() {
	var err error
	// set init time
	InitAt = time.Now()
	// Parse command line flags (-t DISCORD_BOT_TOKEN)
	flag.StringVar(&Token, "t", "", "Discord Bot Token")
	flag.Parse()
	// overwrite with environment variables if set DISCORD_BOT_TOKEN=…
	if os.Getenv("DISCORD_BOT_TOKEN") != "" {
		Token = os.Getenv("DISCORD_BOT_TOKEN")
	}
	// create a new Discordgo Bot Client
	fmt.Println("connecting to Discord, Token Length:", len(Token))
	Dg, err = discordgo.New("Bot " + Token)
	if err != nil {
		panic(err)
	}
}

func Handler(container dhelpers.EventContainer) error {
	var err error

	// set handle time
	HandleAt = time.Now()

	switch container.Type {
	case dhelpers.MessageCreateEventType:
		var event dhelpers.EventMessageCreate
		err = jsoniter.Unmarshal(container.Data, &event)
		if err != nil {
			return errors.New("error unmarshaling " + string(container.Type) + ": " + err.Error())
		}

		err = MessageCreate(container, event)
		if err != nil {
			return errors.New("error processing " + string(container.Type) + ": " + err.Error())
		}
	}

	return nil
}

func MessageCreate(container dhelpers.EventContainer, event dhelpers.EventMessageCreate) (err error) {
	// respond "pong!" to "ping"
	switch event.Alias {
	case "ping-myself":
		_, err = Dg.ChannelMessageSend(event.Event.ChannelID, "/ping")
		if err != nil {
			return err
		}
	case "ping":
		_, err = Dg.ChannelMessageSendComplex(event.Event.ChannelID, &discordgo.MessageSend{
			Embed: &discordgo.MessageEmbed{
				Title:     "Pong!",
				Timestamp: time.Now().Format(time.RFC3339),
				Footer: &discordgo.MessageEmbedFooter{
					Text:    "requested by " + event.Event.Author.Username + "#" + event.Event.Author.Discriminator,
					IconURL: event.Event.Author.AvatarURL("64"),
				},
				Author: &discordgo.MessageEmbedAuthor{
					Name:    event.BotUser.Username + "#" + event.BotUser.Discriminator,
					IconURL: event.BotUser.AvatarURL("64"),
				},
				Fields: []*discordgo.MessageEmbedField{
					{
						Name:   "Init",
						Value:  InitAt.Format(time.StampNano),
						Inline: false,
					},
					{
						Name:   "Handler",
						Value:  HandleAt.Format(time.StampNano),
						Inline: false,
					},
					{
						Name:   "Gateway => Lambda",
						Value:  HandleAt.Sub(container.ReceivedAt).String(),
						Inline: false,
					},
					{
						Name:   "Gateway Uptime",
						Value:  time.Now().Sub(container.GatewayStarted).String() + "\nStarted at " + strconv.FormatInt(container.GatewayStarted.Unix(), 10),
						Inline: false,
					},
					{
						Name:   "Args",
						Value:  "`" + strings.Join(event.Args, "`, `") + "`",
						Inline: false,
					},
					{
						Name:   "Used Prefix",
						Value:  event.Prefix,
						Inline: false,
					},
				},
			},
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func main() {
	lambda.Start(Handler)
}
