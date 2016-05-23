package main

import (
	"flag"
	"fmt"
	"strings"

	"github.com/Kane-Sendgrid/chgk-bot/bot"
	"github.com/danott/envflag"
	"github.com/nlopes/slack"
)

func main() {
	var token string
	var adminToken string
	var outputdir string
	flag.StringVar(&token, "token", "", "Slack token")
	flag.StringVar(&adminToken, "admintoken", "", "Slack token")
	flag.StringVar(&outputdir, "outputdir", "", "Output dir")
	envflag.Parse()
	api := slack.New(token)
	adminAPI := slack.New(adminToken)
	// api.GetChannelHistory("", slack.HistoryParameters)
	// api.SetDebug(true)
	rtm := api.NewRTM()
	bots := bot.NewChannelBots()
	go rtm.ManageConnection()

Loop:
	for {
		select {
		case msg := <-rtm.IncomingEvents:
			fmt.Print("Event Received: ")
			switch ev := msg.Data.(type) {
			case *slack.HelloEvent:
				// Ignore hello

			case *slack.ConnectedEvent:
				// fmt.Println("Infos:", ev.Info)
				// fmt.Println("Connection counter:", ev.ConnectionCount)
				// // Replace #general with your Channel ID
				// rtm.SendMessage(rtm.NewOutgoingMessage("Hello world", "#general"))

			case *slack.MessageEvent:
				fmt.Println(ev.Channel)
				b := bots.GetOrCreate(rtm, ev.Channel)
				fmt.Printf("Message: %v\n", ev)
				loText := strings.ToLower(ev.Msg.Text)
				if ev.Msg.Text == "test" {
					adminAPI.DeleteMessage(ev.Channel, ev.Msg.Timestamp)
					b.BotMessage("testimage", "")
				}
				if strings.HasPrefix(loText, "!начать") {
					adminAPI.DeleteMessage(ev.Channel, ev.Msg.Timestamp)
					go b.StartGame(ev.Msg.Text)
				}
				if strings.HasPrefix(loText, "!счет") {
					b.TellScore()
				}
				if strings.HasPrefix(loText, "!стоп") {
					b.Cancel()
				}
				if strings.HasPrefix(loText, "!ответ") {
					b.Answer()
				}
				if loText == "++" {
					b.IncScoreRight()
				}
				if loText == "--" {
					b.IncScoreWrong()
				}
				if ev.File != nil && strings.Contains(ev.File.Title, ".docx") {
					b.SaveDoc(token, outputdir, ev.File)
				}

				if strings.HasPrefix(loText, "!вопрос") ||
					strings.HasPrefix(loText, "?вопрос") ||
					strings.HasPrefix(loText, "вопрос ") {
					go b.StartTimer(ev.Msg.Text, "")
				}
			case *slack.PresenceChangeEvent:
				// fmt.Printf("Presence Change: %v\n", ev)

			case *slack.LatencyReport:
				// fmt.Printf("Current latency: %v\n", ev.Value)

			case *slack.RTMError:
				// fmt.Printf("Error: %s\n", ev.Error())

			case *slack.InvalidAuthEvent:
				// fmt.Printf("Invalid credentials")
				break Loop

			default:

				// Ignore other events..
				// fmt.Printf("Unexpected: %v\n", msg.Data)
			}
		}
	}
}
