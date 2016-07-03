package bot

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"regexp"
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
func (b *Bot) StartGame(command, gameType string) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println(r)
			b.BotMessage("Ошибка обработки запроса", "")
		}
	}()
	b.BotMessage(fmt.Sprintf("Начинаю новую игру. Лимит времени %d минуты", int(b.delay1.Minutes()+1)), "")
	b.scoreRight = 0
	b.scoreWrong = 0
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
	startQuestion := 1
	url := options[1]
	if len(options) >= 3 {
		var err error
		startQuestion, err = strconv.Atoi(options[2])
		if err != nil {
			b.BotMessage("Ошибка: "+err.Error(), "")
		}
	}
	url = url[1 : len(url)-1]
	fmt.Println(options, url)

	s, err := dbchgk.LoadSuite(url, gameType, command)
	if err != nil {
		b.BotMessage("Ошибка: "+err.Error(), "")
		return
	}
	b.totalQuestions = len(s.Questions)
	for i, q := range s.Questions {
		qNum := strconv.Itoa(i + 1)
		if i+1 < startQuestion {
			continue
		}
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
		b.WaitForAnswer()
		b.BotColorMessage("ОТВЕТ "+q.Answer, "", "#ff0000")
		if len(q.Comments) > 0 {
			b.BotColorMessage("Комментарий "+q.Comments, "", "#ff0000")
		}
		b.BotColorMessage("Засчитать? ++ или --", "", "#ff0000")
		b.WaitForAnswer()
	}
	b.BotColorMessage("Тур закончен, окончательный счет:", "", "#ff0000")
	b.TellScore()
}

//SetDelay ...
func (b *Bot) SetDelay(delay int) {
	b.delay1 = time.Duration(delay-1) * time.Minute
}

//WaitForAnswer ...
func (b *Bot) WaitForAnswer() {
	<-b.answerChan
	b.answerChan = make(chan bool)
}

//StartTimer ...
func (b *Bot) StartTimer(question, picture string) {
	b.BotColorMessage("Задаю вопрос:", "", "#00ff00")
	b.BotColorMessage(question, picture, "#ff0000")
	b.BotColorMessage("Даю 20 секунд на чтение вопроса", "", "#00ff00")
	reason := b.Sleep(20 * time.Second)
	if reason != sleepReasonNormal {
		return
	}
	b.BotColorMessage("Время!", "", "#00ff00")
	reason = b.Sleep(b.delay1)
	if reason != sleepReasonNormal {
		return
	}
	b.BotColorMessage("Осталась 1 минута", "", "#00ff00")
	reason = b.Sleep(b.delay2)
	if reason != sleepReasonNormal {
		return
	}
	b.BotColorMessage("Осталось 20 секунд", "", "#00ff00")
	b.BotColorMessage("Повторяю вопрос:", "", "#00ff00")
	b.BotColorMessage(question, picture, "#ff0000")
	reason = b.Sleep(b.delay3)
	if reason != sleepReasonNormal {
		return
	}
	b.BotColorMessage("Ваш ответ?", "", "#00ff00")
}

//BotMessage ...
func (b *Bot) BotMessage(title, image string) {
	params := slack.NewPostMessageParameters()
	params.AsUser = true
	params.UnfurlMedia = true
	params.UnfurlLinks = true
	params.Markdown = true
	params.Attachments = []slack.Attachment{
		slack.Attachment{
			// Title: title,
			// Text:       "*" + title + "*",
			ImageURL:   image,
			MarkdownIn: []string{"text", "pretext", "fields"},
		},
	}
	b.rtm.PostMessage(b.channel, title, params)
}

//BotColorMessage ...
func (b *Bot) BotColorMessage(title, image, color string) {
	params := slack.NewPostMessageParameters()
	params.AsUser = true
	params.UnfurlMedia = true
	params.UnfurlLinks = true
	params.Markdown = true
	params.Attachments = []slack.Attachment{
		slack.Attachment{
			Color: color,
			Title: title,
			// Text:       "*" + title + "*",
			ImageURL:   image,
			MarkdownIn: []string{"text", "pretext", "fields"},
		},
	}
	b.rtm.PostMessage(b.channel, "", params)
}

//Sleep ...
func (b *Bot) Sleep(t time.Duration) int {
	select {
	case <-b.ctx.Done():
		return sleepReasonStop
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

//CheckCaptain ...
func (b *Bot) CheckCaptain(userID string) bool {
	if b.captain == "" {
		return true
	}
	if b.captain == userID {
		return true
	}
	b.BotMessage(fmt.Sprintf("Капитан игры - %s. Поменяйте капитана командой !капитан @notabene(или другой ник). Уберите капитана командой: !капитан (без ника)",
		b.captainUsername), "")
	return false
}

//SetCaptainCommand ...
func (b *Bot) SetCaptainCommand(cmd string) {
	cmd = strings.TrimSpace(cmd)
	if cmd == "!капитан" {
		b.captain = ""
		b.captainUsername = ""
		b.BotMessage("Капитан убран", "")
		return
	}
	captainRE := regexp.MustCompile(`<@(.*?)>`)
	match := captainRE.FindStringSubmatch(cmd)
	if match == nil {
		b.BotMessage("Неправильное имя юзера (например @notabene). !капитан - без ника, убирает капитана", "")
		return
	}
	fmt.Println(match)
	b.captain = match[1]
	user, err := b.rtm.GetUserInfo(b.captain)
	if err != nil {
		fmt.Println(err)
		b.BotMessage("Неправильное имя юзера (например @notabene). !капитан - без ника, убирает капитана", "")
		return
	}
	fmt.Println(user.Name)
	b.captainUsername = user.Name
	b.BotMessage(fmt.Sprintf("Капитан - %s", b.captainUsername), "")
}

//TellScore ...
func (b *Bot) TellScore() {
	b.BotMessage(fmt.Sprintf("Счет (верно/неверно) %d / %d. Всего вопросов: %d. Капитан - %s", b.scoreRight, b.scoreWrong, b.totalQuestions, b.captainUsername), "")
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
