package main

import (
	"flag"
	"os"

	"strings"

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

func Handler(request events.KinesisEvent) (events.APIGatewayProxyResponse, error) {
	var err error
	// set handle time
	HandleAt = time.Now()

	if request.Records == nil || len(request.Records) <= 0 {
		fmt.Println("received nothing from kinesis")
		return events.APIGatewayProxyResponse{
				Body:       "",
				StatusCode: 200,
			},
			nil
	}

	for _, record := range request.Records {
		var m DMessageCreateEvent

		err = jsoniter.Unmarshal(record.Kinesis.Data, &m)
		if err != nil {
			fmt.Println("error unpacking event:", err.Error())
			return events.APIGatewayProxyResponse{
					Body:       "",
					StatusCode: 200,
				},
				nil
		}

		err = MessageCreate(record, m)
		if err != nil {
			fmt.Println("error processing discordgo.MessageCreate:", err.Error())
			return events.APIGatewayProxyResponse{
					Body:       "",
					StatusCode: 200,
				},
				nil
		}
	}

	return events.APIGatewayProxyResponse{Body: "", StatusCode: 200}, nil
}

func MessageCreate(event events.KinesisEventRecord, messageCreate DMessageCreateEvent) (err error) {
	// respond "pong!" to "ping"
	if strings.ToLower(strings.TrimSpace(messageCreate.Event.Content)) == "ping" {
		_, err = Dg.ChannelMessageSendComplex(messageCreate.Event.ChannelID, &discordgo.MessageSend{
			Embed: &discordgo.MessageEmbed{
				Title:     "Pong!",
				Timestamp: time.Now().Format(time.RFC3339),
				Footer: &discordgo.MessageEmbedFooter{
					Text: "requested by " + messageCreate.Event.Author.Username + "#" + messageCreate.Event.Author.Discriminator + " | " +
						//"Region " + event.AwsRegion + " | " +
						//"Event #" + event.EventID + " | " +
						//"Source " + event.EventSource + " | " +
						//"Name " + event.EventName + " | " +
						"Sequence #" + event.Kinesis.SequenceNumber,
					IconURL: messageCreate.Event.Author.AvatarURL("64"),
				},
				Author: &discordgo.MessageEmbedAuthor{
					Name:    messageCreate.BotUser.Username + "#" + messageCreate.BotUser.Discriminator,
					IconURL: messageCreate.BotUser.AvatarURL("64"),
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
