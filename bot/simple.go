package bot

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/context"

	"github.com/Kane-Sendgrid/chgk-bot/dbchgk"
	"github.com/nlopes/slack"
)

const (
	sleepReasonNormal = iota
	sleepReasonStop
	sleepReasonAnswer
)

//StartGame ...
func (b *Bot) StartGame(command string) {
	defer func() {
		if r := recover(); r != nil {
			b.BotMessage("Ошибка обработки запроса", "")
		}
	}()
	b.BotMessage("Начинаю новую игру...", "")
	b.scoreRight = 0
	b.scoreWrong = 0
	b.delay1 = 60 * time.Second
	b.delay2 = 40 * time.Second
	b.delay3 = 20 * time.Second
	b.answerDelay = 15 * time.Second
	b.questionDelay = 10 * time.Second

	// b.delay1 = 1 * time.Second
	// b.delay2 = 1 * time.Second
	// b.delay3 = 1 * time.Second
	// b.answerDelay = 1 * time.Second
	// b.questionDelay = 1 * time.Second

	options := strings.Split(command, " ")
	url := options[1]
	url = url[1 : len(url)-1]
	fmt.Println(options, url)

	s, err := dbchgk.LoadSuite(url)
	if err != nil {
		b.BotMessage("Ошибка: "+err.Error(), "")
		return
	}
	for i, q := range s.Questions {
		qNum := strconv.Itoa(i + 1)
		select {
		case <-b.ctx.Done():
			return
		default:
		}
		b.answerChan = make(chan bool)
		fmt.Println(">>> Q", q.Question)
		fmt.Println(">>> Q", q.Picture)
		fmt.Println(">>> A", q.Answer)
		b.StartTimer("ВОПРОС №"+qNum+". "+q.Question, q.Picture)
		<-b.answerChan
		b.answerChan = make(chan bool)
		b.BotMessage("ОТВЕТ "+q.Answer, "")
		if len(q.Comments) > 0 {
			b.BotMessage("Комментарий "+q.Comments, "")
		}
		b.BotMessage("Засчитать? ++ или --", "")
		b.Sleep(b.questionDelay)
	}

}

//StartTimer ...
func (b *Bot) StartTimer(question, picture string) {
	b.BotMessage(question, picture)
	b.BotMessage("Время!", "")
	reason := b.Sleep(b.delay1)
	if reason != sleepReasonNormal {
		return
	}
	b.BotMessage("Осталась 1 минута", "")
	reason = b.Sleep(b.delay2)
	if reason != sleepReasonNormal {
		return
	}
	b.BotMessage("Осталось 20 секунд", "")
	b.BotMessage("Повторяю вопрос: "+question, picture)
	reason = b.Sleep(b.delay3)
	if reason != sleepReasonNormal {
		return
	}
	b.BotMessage("Ваш ответ?", "")
}

//BotMessage ...
func (b *Bot) BotMessage(title, image string) {
	params := slack.PostMessageParameters{
		Text: title,
		Attachments: []slack.Attachment{
			slack.Attachment{
				Color: "#ff0000",
				Title: title,
				// Text:       "*" + title + "*",
				ImageURL:   image,
				MarkdownIn: []string{"text", "pretext", "fields"},
			},
		},
	}
	b.rtm.PostMessage(b.channel, "", params)
}

//Sleep ...
func (b *Bot) Sleep(t time.Duration) int {
	select {
	case <-b.ctx.Done():
		return sleepReasonNormal
	case <-b.answerChan:
		return sleepReasonAnswer
	case <-time.After(t):
		return sleepReasonNormal
	}
}

//Answer ...
func (b *Bot) Answer() {
	close(b.answerChan)
}

//Cancel ...
func (b *Bot) Cancel() {
	b.BotMessage("Останавливаю", "")
	b.cancel()
	time.Sleep(2 * time.Second)
	ctx, cancel := context.WithCancel(context.Background())
	b.ctx = ctx
	b.cancel = cancel
}

//TellScore ...
func (b *Bot) TellScore() {
	b.BotMessage(fmt.Sprintf("Счет (верно/неверно) %d / %d", b.scoreRight, b.scoreWrong), "")
}

//IncScoreRight ...
func (b *Bot) IncScoreRight() {
	b.scoreRight++
	b.TellScore()
}

//IncScoreWrong ...
func (b *Bot) IncScoreWrong() {
	b.scoreWrong++
	b.TellScore()
}

//SaveDoc ...
func (b *Bot) SaveDoc(token, outputdir string, file *slack.File) {
	fmt.Println(file.Title)
	fmt.Println(file.URLPrivateDownload)
	name := file.Title[0 : len(file.Title)-5]
	outputdir = path.Join(outputdir, name)
	os.MkdirAll(outputdir, 0777)
	outputFileName := path.Join(outputdir, name+".docx")
	cmdArgs := []string{file.URLPrivateDownload,
		"-O", outputFileName,
		"--header=Authorization: Bearer " + token,
	}
	out, err := exec.Command("wget", cmdArgs...).Output()
	if err != nil {
		fmt.Println(cmdArgs)
		fmt.Println(string(out))
		b.BotMessage("Ошибка: "+err.Error(), "")
		return
	}

	cmdArgs = []string{outputFileName,
		"--output-dir", outputdir,
	}
	out, err = exec.Command("mammoth", cmdArgs...).Output()
	if err != nil {
		fmt.Println(cmdArgs)
		fmt.Println(string(out))
		b.BotMessage("Ошибка: "+err.Error(), "")
		return
	}

	b.BotMessage(fmt.Sprintf("Файл сохранен. Игру можно начать командой: !начать http://kane1.ipq.co/chgk-files/%s/%s.html", name, name), "")
}
