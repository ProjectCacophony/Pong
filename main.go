package main

import (
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/bwmarrin/discordgo"
	"gitlab.com/project-d-collab/dhelpers"
	"gitlab.com/project-d-collab/dhelpers/cache"
	"gitlab.com/project-d-collab/dhelpers/components"
)

var (
	serviceName = "lambda/pong"
)

func init() {
	// init components
	components.InitLogger(serviceName)
	err := components.InitSentry()
	dhelpers.CheckErr(err)
}

// Handler is the lambda entry point when event is triggered
// error handling is built in, just panic (dhelpers.CheckErr)
func Handler(event dhelpers.EventContainer) {
	// pass on event to the correct method
	switch event.Type {
	case dhelpers.MessageCreateEventType:

		for _, destination := range event.Destinations {

			switch destination.Alias {
			case "ping":

				ping(event)
				return
			}
		}
	}
}

// ping is triggered when we receive a ping command
func ping(container dhelpers.EventContainer) {
	// measure time
	pingStart := time.Now()
	defer func() {
		cache.GetLogger().Infoln("ping took", time.Since(pingStart).String())
	}()

	// send ping response
	_, err := container.SendComplex(container.MessageCreate.ChannelID, &discordgo.MessageSend{
		Embed: &discordgo.MessageEmbed{
			Title: "Pong!",
			Fields: []*discordgo.MessageEmbedField{
				{
					Name:   "Gateway => Lambda",
					Value:  pingStart.Sub(container.ReceivedAt).String(),
					Inline: false,
				},
				{
					Name:   "Gateway Uptime",
					Value:  time.Since(container.GatewayStarted).String(),
					Inline: false,
				},
			},
		},
	})
	dhelpers.CheckErr(err)
}

func main() {
	// register handler
	lambda.StartHandler(dhelpers.NewLambdaHandler(serviceName, Handler))
}
