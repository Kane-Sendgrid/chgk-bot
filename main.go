package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
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

	for {
		select {
		case msg := <-rtm.IncomingEvents:
			func() {
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
					defer func() {
						if r := recover(); r != nil {
							bots.Delete(ev.Channel)
							b.BotMessage("Ошибка. Бот обнулен", "")
						}
					}()
					fmt.Printf("Message: %v\n", ev)
					loText := strings.TrimSpace(strings.ToLower(ev.Msg.Text))
					if !strings.HasPrefix(loText, "!начать") && !strings.HasPrefix(loText, "!капитан") &&
						loText != "!счет" && (strings.HasPrefix(loText, "!") || loText == "++" || loText == "--") {
						if !b.CheckCaptain(ev.User) {
							return
						}
					}
					if ev.Msg.Text == "test" {
						adminAPI.DeleteMessage(ev.Channel, ev.Msg.Timestamp)
						b.BotMessage("testimage", "")
					}
					if strings.HasPrefix(loText, "!начать") {
						adminAPI.DeleteMessage(ev.Channel, ev.Msg.Timestamp)
						go b.StartGame(ev.Msg.Text, "simple")
					}
					if strings.HasPrefix(loText, "!startregex") {
						adminAPI.DeleteMessage(ev.Channel, ev.Msg.Timestamp)
						go b.StartGame(ev.Msg.Text, "regex")
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
					if strings.HasPrefix(loText, "!капитан") {
						fmt.Println("text >>>", ev.Msg.Text)
						fmt.Println("user >>>", ev.User)
						fmt.Println("username >>>", ev.Username)
						b.SetCaptainCommand(ev.Msg.Text)
					}
					if strings.HasPrefix(loText, "!время") {
						tasks := strings.Split(loText, " ")
						if len(tasks) == 2 {
							delay, err := strconv.Atoi(tasks[1])
							if err != nil {
								b.BotMessage("Ошибка. Введите команду: !время ЧИСЛО", "")
								return
							}
							b.SetDelay(delay)
							b.BotMessage(fmt.Sprintf("Установлено время игры %d минут", delay), "")
						}
					}
					if loText == "++" {
						b.IncScoreRight()
						b.Answer()
					}
					if loText == "--" {
						b.IncScoreWrong()
						b.Answer()
					}
					if ev.File != nil && strings.Contains(ev.File.Title, ".docx") {
						b.SaveDoc(token, outputdir, ev.File)
					}

					if len(loText) > 20 && (strings.HasPrefix(loText, "!вопрос") ||
						strings.HasPrefix(loText, "?вопрос") ||
						strings.HasPrefix(loText, "?")) {
						b.Cancel()
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
					os.Exit(1)

				default:

					// Ignore other events..
					// fmt.Printf("Unexpected: %v\n", msg.Data)
				}
			}()
		}
	}
}
