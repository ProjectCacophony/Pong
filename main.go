package main

import (
	"flag"
	"fmt"
	"os"

	"strings"

	"errors"

	"encoding/base64"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/bwmarrin/discordgo"
	"github.com/vmihailenco/msgpack"
)

var (
	Token       string
	SqsQueueUrl string
	Svc         *sqs.SQS
	Dg          *discordgo.Session
)

func init() {
	// Parse command line flags (-t DISCORD_BOT_TOKEN -sqs SQS_QUEUE_URL)
	flag.StringVar(&Token, "t", "", "Discord Bot Token")
	flag.StringVar(&SqsQueueUrl, "sqs", "", "Amazon SQS Queue URL")
	flag.Parse()
	// overwrite with environment variables if set DISCORD_BOT_TOKEN=… SQS_QUEUE_URL=…
	if os.Getenv("DISCORD_BOT_TOKEN") != "" {
		Token = os.Getenv("DISCORD_BOT_TOKEN")
	}
	if os.Getenv("SQS_QUEUE_URL") != "" {
		SqsQueueUrl = os.Getenv("SQS_QUEUE_URL")
	}
}

func Handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var err error
	// setup Amazon Session
	fmt.Println("connecting to Amazon SQS, URL:", SqsQueueUrl)
	awsSession := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	// setup Amazon SQS queue
	Svc = sqs.New(awsSession)

	// create a new Discordgo Bot Client
	fmt.Println("connecting to Discord, Token Length:", len(Token))
	Dg, err = discordgo.New("Bot " + Token)
	if err != nil {
		fmt.Println("error creating discord session,", err.Error())
		return events.APIGatewayProxyResponse{Body: request.Body, StatusCode: 500},
			errors.New("error creating discord session: " + err.Error())
	}

	// Receive last Amazon SQS message
	result, err := Svc.ReceiveMessage(&sqs.ReceiveMessageInput{
		QueueUrl: aws.String(SqsQueueUrl),
		AttributeNames: aws.StringSlice([]string{
			"SentTimestamp",
		}),
		MaxNumberOfMessages: aws.Int64(1),
		MessageAttributeNames: aws.StringSlice([]string{
			"All",
		}),
		WaitTimeSeconds: aws.Int64(60),
	})
	if err != nil {
		fmt.Println("unable to receive message from queue:", err.Error())
		return events.APIGatewayProxyResponse{Body: request.Body, StatusCode: 500},
			errors.New("unable to receive message from queue: " + err.Error())
	}

	if result == nil || result.Messages == nil || len(result.Messages) <= 0 {
		fmt.Println("received nothing from queue")
		return events.APIGatewayProxyResponse{Body: request.Body, StatusCode: 500},
			errors.New("received nothing from queue")
	}

	// check if Message is valid
	receivedEvent := result.Messages[0]

	if receivedEvent.MessageAttributes == nil {
		fmt.Println("received message with no event type from queue")
		return events.APIGatewayProxyResponse{Body: request.Body, StatusCode: 500},
			errors.New("received message with no event type from queue")
	}

	if eventType, ok := receivedEvent.MessageAttributes["EventType"]; !ok || eventType == nil || eventType.String() == "" {
		fmt.Println("received message with no event type from queue")
		return events.APIGatewayProxyResponse{Body: request.Body, StatusCode: 500},
			errors.New("received message with no event type from queue")
	}

	msgpackData, err := base64.StdEncoding.DecodeString(*receivedEvent.Body)
	if err != nil {
		fmt.Println("unable to base64 decode the message data")
		return events.APIGatewayProxyResponse{Body: request.Body, StatusCode: 500},
			errors.New("unable to base64 decode the message data")
	}

	// match and unpack received event
	switch receivedEvent.MessageAttributes["EventType"].String() {
	case "discordgo.MessageCreate":
		var m *discordgo.MessageCreate
		err = msgpack.Unmarshal(msgpackData, &m)
		if err != nil {
			fmt.Println("error unpacking event:", err.Error())
			return events.APIGatewayProxyResponse{Body: request.Body, StatusCode: 500},
				errors.New("error unpacking event: " + err.Error())
		}
		err = MessageCreate(m)
		if err != nil {
			fmt.Println("error processing discordgo.MessageCreate:", err.Error())
			return events.APIGatewayProxyResponse{Body: request.Body, StatusCode: 500},
				errors.New("error processing discordgo.MessageCreate: " + err.Error())
		}
	default:
		fmt.Println("received invalid event type:", receivedEvent.MessageAttributes["EventType"].String())
		return events.APIGatewayProxyResponse{Body: request.Body, StatusCode: 500},
			errors.New("received invalid event type: " + receivedEvent.MessageAttributes["EventType"].String())
	}

	return events.APIGatewayProxyResponse{Body: request.Body, StatusCode: 200}, nil
}

func MessageCreate(m *discordgo.MessageCreate) (err error) {
	// respond "pong!" to "ping"
	if strings.ToLower(strings.TrimSpace(m.Content)) == "ping" {
		_, err = Dg.ChannelMessageSend(m.ChannelID, "pong!")
		if err != nil {
			return err
		}
	}
	return nil
}

func main() {
	lambda.Start(Handler)
}
