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
	"gitlab.com/project-d-collab/dhelpers"
)

var (
	token          string
	dg             *discordgo.Session
	initAt         time.Time
	initFinishedAt time.Time
)

func init() {
	// set init time
	initAt = time.Now()
	var err error
	// Parse command line flags (-t DISCORD_BOT_TOKEN)
	flag.StringVar(&token, "t", "", "Discord Bot Token")
	flag.Parse()
	// overwrite with environment variables if set DISCORD_BOT_TOKEN=…
	if os.Getenv("DISCORD_BOT_TOKEN") != "" {
		token = os.Getenv("DISCORD_BOT_TOKEN")
	}
	// create a new Discordgo Bot Client
	fmt.Println("connecting to Discord, Token Length:", len(token))
	dg, err = discordgo.New("Bot " + token)
	if err != nil {
		panic(err)
	}

	initFinishedAt = time.Now()
}

func Handler(container dhelpers.EventContainer) error {
	// benchmark
	handlerStart := time.Now()
	defer func() {
		fmt.Println("handler took", time.Now().Sub(handlerStart).String())
	}()

	var err error

	switch container.Type {
	case dhelpers.MessageCreateEventType:
		err = MessageCreate(handlerStart, container)
		if err != nil {
			return errors.New("error processing " + string(container.Type) + ": " + err.Error())
		}
	}

	return nil
}

func MessageCreate(handleAt time.Time, container dhelpers.EventContainer) (err error) {
	// benchmark
	messageCreateStart := time.Now()
	defer func() {
		fmt.Println("messagecreate took", time.Now().Sub(messageCreateStart).String())
	}()

	// respond "pong!" to "ping"
	switch container.Alias {
	case "ping-myself":
		_, err = dg.ChannelMessageSend(container.MessageCreate.ChannelID, "/ping")
		if err != nil {
			return err
		}
	case "ping":
		_, err = dg.ChannelMessageSendComplex(container.MessageCreate.ChannelID, &discordgo.MessageSend{
			Embed: &discordgo.MessageEmbed{
				Title:     "Pong!",
				Timestamp: time.Now().Format(time.RFC3339),
				Footer: &discordgo.MessageEmbedFooter{
					Text:    "requested by " + container.MessageCreate.Author.Username + "#" + container.MessageCreate.Author.Discriminator,
					IconURL: container.MessageCreate.Author.AvatarURL("64"),
				},
				Author: &discordgo.MessageEmbedAuthor{
					Name:    container.BotUser.Username + "#" + container.BotUser.Discriminator,
					IconURL: container.BotUser.AvatarURL("64"),
				},
				Fields: []*discordgo.MessageEmbedField{
					{
						Name:   "Init",
						Value:  "At " + initAt.Format(time.StampNano) + "\nTook " + initFinishedAt.Sub(initAt).String(),
						Inline: false,
					},
					{
						Name:   "Handler",
						Value:  handleAt.Format(time.StampNano),
						Inline: false,
					},
					{
						Name:   "Gateway => Lambda",
						Value:  handleAt.Sub(container.ReceivedAt).String(),
						Inline: false,
					},
					{
						Name:   "Gateway Uptime",
						Value:  time.Now().Sub(container.GatewayStarted).String() + "\nStarted at " + strconv.FormatInt(container.GatewayStarted.Unix(), 10),
						Inline: false,
					},
					{
						Name:   "Args",
						Value:  "`" + strings.Join(container.Args, "`, `") + "`",
						Inline: false,
					},
					{
						Name:   "Used Prefix",
						Value:  container.Prefix,
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
	wrappedHandler := newHandler(Handler)
	lambda.StartHandler(wrappedHandler)
}
