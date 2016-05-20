package main

import (
	"fmt"

	"github.com/nlopes/slack"
)

func main() {
	api := slack.New("xoxb-44710107905-CaKMcJiN41xeykAzfYKItlMT")
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
				if ev.Msg.Text == "test" {
					fmt.Printf("Message: %v\n", ev)
					rtm.SendMessage(rtm.NewOutgoingMessage("Hello from bot!", "#random"))
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
