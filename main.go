package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/Kane-Sendgrid/chgk-bot/dbchgk"
	"github.com/danott/envflag"
	"github.com/nlopes/slack"
)

func botMessage(rtm *slack.RTM, title, image string) {
	params := slack.PostMessageParameters{
		Attachments: []slack.Attachment{
			slack.Attachment{
				Color:    "#ff0000",
				Title:    title,
				ImageURL: image,
			},
		},
	}
	rtm.PostMessage("C1AL8DNMT", "", params)
}

func main() {
	var token string
	flag.StringVar(&token, "token", "", "Slack token")
	envflag.Parse()
	api := slack.New(token)
	// api.GetChannelHistory("", slack.HistoryParameters)
	api.SetDebug(true)

	rtm := api.NewRTM()
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
				fmt.Printf("Message: %v\n", ev)
				if ev.Msg.Text == "test" {
					botMessage(rtm, "testimage", "")
				}
				if ev.Msg.Text == "!начать" {
					botMessage(rtm, "Начинаю новую игру...", "")
					s := dbchgk.LoadSuite()
					for _, q := range s.Questions {
						fmt.Println(">>> Q", q.Question)
						fmt.Println(">>> Q", q.Picture)
						fmt.Println(">>> A", q.Answer)
						botMessage(rtm, "ВОПРОС "+q.Question, q.Picture)
						time.Sleep(5 * time.Second)
						botMessage(rtm, "ОТВЕТ "+q.Answer, "")
						time.Sleep(5 * time.Second)
					}
				}
				if ev.Msg.Text == "!время" {
					rtm.SendMessage(rtm.NewOutgoingMessage("Отсчет каждые 5 секунд", "C1AL8DNMT"))
					go func() {
						for {
							time.Sleep(5 * time.Second)
							rtm.SendMessage(rtm.NewOutgoingMessage("Тик...", "C1AL8DNMT"))
						}
					}()
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
