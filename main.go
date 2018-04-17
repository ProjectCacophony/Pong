package main

import (
	"flag"
	"os"

	"fmt"

	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/bwmarrin/discordgo"
	"github.com/json-iterator/go"
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

func Handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var err error
	// set handle time
	HandleAt = time.Now()

	eventType, ok := request.QueryStringParameters["type"]
	if !ok {
		fmt.Println("error processing event without event type")
		return events.APIGatewayProxyResponse{
				Body:       "",
				StatusCode: 200,
			},
			nil
	}

	switch eventType {
	case MessageCreateEventType:
		var event DDiscordEventMessageCreate

		err = jsoniter.Unmarshal([]byte(request.Body), &event)
		if err != nil {
			fmt.Println("error unpacking event:", err.Error())
			return events.APIGatewayProxyResponse{
					Body:       "",
					StatusCode: 200,
				},
				nil
		}

		err = MessageCreate(event)
		if err != nil {
			fmt.Println("error processing", event.Type, ":", err.Error())
			return events.APIGatewayProxyResponse{
					Body:       "",
					StatusCode: 200,
				},
				nil
		}
	}
	return events.APIGatewayProxyResponse{Body: "", StatusCode: 200}, nil
}

func MessageCreate(event DDiscordEventMessageCreate) (err error) {
	// respond "pong!" to "ping"
	switch event.Alias {
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
						Value:  HandleAt.Sub(event.GatewayReceivedAt).String(),
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
