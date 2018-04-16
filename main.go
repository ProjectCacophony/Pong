package main

import (
	"flag"
	"fmt"
	"os"

	"strings"

	"errors"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/bwmarrin/discordgo"
	"github.com/vmihailenco/msgpack"
)

var (
	Token string
)

func init() {
	// Parse command line flags (-t DISCORD_BOT_TOKEN)
	flag.StringVar(&Token, "t", "", "Discord Bot Token")
	flag.Parse()
	// overwrite with environment variables if set DISCORD_BOT_TOKEN=…
	if os.Getenv("DISCORD_BOT_TOKEN") != "" {
		Token = os.Getenv("DISCORD_BOT_TOKEN")
	}

}
func Handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// unpack received event
	// TODO: check which type of event we are receiving (where are my MessageAttributes?)
	var m *discordgo.MessageCreate
	err := msgpack.Unmarshal([]byte(request.Body), &m)
	if err != nil {
		fmt.Println("error unpacking event:", err.Error())
		return events.APIGatewayProxyResponse{Body: request.Body, StatusCode: 500}, errors.New("error unpacking event: " + err.Error())
	}

	// create a new Discordgo Bot Client
	fmt.Println("connecting to Discord, Token Length:", len(Token))
	dg, err := discordgo.New("Bot " + Token)
	if err != nil {
		fmt.Println("error creating discord session,", err.Error())
		return events.APIGatewayProxyResponse{Body: request.Body, StatusCode: 500}, errors.New("error creating discord session: " + err.Error())
	}

	// respond "pong!" to "ping"
	if strings.ToLower(m.Content) == "ping" {
		dg.ChannelMessageSend(m.ChannelID, "pong!")
	}

	return events.APIGatewayProxyResponse{Body: request.Body, StatusCode: 200}, nil
}

func main() {
	lambda.Start(Handler)
}
