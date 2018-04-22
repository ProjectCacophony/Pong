package main

import (
	"os"

	"fmt"

	"time"

	"strconv"
	"strings"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/bwmarrin/discordgo"
	"gitlab.com/project-d-collab/dhelpers"
	"gitlab.com/project-d-collab/dhelpers/cache"
	"gitlab.com/project-d-collab/dhelpers/components"
)

var (
	token           string
	discordEndpoint string
	dg              *discordgo.Session
	initAt          time.Time
	initFinishedAt  time.Time
)

func init() {
	// set init time
	initAt = time.Now()
	var err error
	// read environment variables if set DISCORD_BOT_TOKEN=… DISCORD_ENDPOINT=…
	token = os.Getenv("DISCORD_BOT_TOKEN")
	discordEndpoint = os.Getenv("DISCORD_ENDPOINT")
	// init components
	dhelpers.CheckErr(err)
	components.InitLogger("lambda/pong")
	err = components.InitSentry()
	dhelpers.CheckErr(err)
	// create a new Discordgo Bot Client
	dhelpers.SetDiscordEndpoints(discordEndpoint)
	fmt.Println("set Discord Endpoint API URL to", discordgo.EndpointAPI)
	fmt.Println("connecting to Discord, Token Length:", len(token))
	dg, err = discordgo.New("Bot " + token)
	dhelpers.CheckErr(err)

	initFinishedAt = time.Now()
}

// Handler is the lambda entry point when event is triggered
func Handler(event dhelpers.EventContainer) {
	// benchmark
	handlerStart := time.Now()
	defer func() {
		cache.GetLogger().Infoln("handler took", time.Since(handlerStart).String())
	}()

	switch event.Type {
	case dhelpers.MessageCreateEventType:

		switch event.Args[0] {
		case "ping", "pong":
			ping(event)
		}
	}
}

// MessageCreate is triggered when a MessageCreate event has been received
func ping(container dhelpers.EventContainer) {
	// benchmark
	messageCreateStart := time.Now()
	defer func() {
		cache.GetLogger().Infoln("messagecreate took", time.Since(messageCreateStart).String())
	}()

	var err error

	_, err = dg.ChannelMessageSendComplex(container.MessageCreate.ChannelID, &discordgo.MessageSend{
		Embed: &discordgo.MessageEmbed{
			Title:     "Pong!",
			Timestamp: time.Now().Format(time.RFC3339),
			Footer: &discordgo.MessageEmbedFooter{
				Text:    "requested by " + container.MessageCreate.Author.Username + "#" + container.MessageCreate.Author.Discriminator,
				IconURL: container.MessageCreate.Author.AvatarURL("64"),
			},
			/*
				Author: &discordgo.MessageEmbedAuthor{
					Name:    container.BotUser.Username + "#" + container.BotUser.Discriminator,
					IconURL: container.BotUser.AvatarURL("64"),
				},
			*/
			Fields: []*discordgo.MessageEmbedField{
				{
					Name:   "Init",
					Value:  "At " + initAt.Format(time.StampNano) + "\nTook " + initFinishedAt.Sub(initAt).String(),
					Inline: false,
				},
				{
					Name:   "Handler",
					Value:  messageCreateStart.Format(time.StampNano),
					Inline: false,
				},
				{
					Name:   "Gateway => Lambda",
					Value:  messageCreateStart.Sub(container.ReceivedAt).String(),
					Inline: false,
				},
				{
					Name:   "Gateway Uptime",
					Value:  time.Since(container.GatewayStarted).String() + "\nStarted at " + strconv.FormatInt(container.GatewayStarted.Unix(), 10),
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
	dhelpers.CheckErr(err)
}

func main() {
	lambda.StartHandler(dhelpers.NewLambdaHandler("lambda/pong", Handler))
}
