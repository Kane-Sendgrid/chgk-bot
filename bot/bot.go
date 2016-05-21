package bot

import (
	"sync"
	"time"

	"github.com/nlopes/slack"
	"golang.org/x/net/context"
)

//Bot ...
type Bot struct {
	ctx        context.Context
	cancel     context.CancelFunc
	rtm        *slack.RTM
	channel    string
	scoreRight int
	scoreWrong int

	delay1        time.Duration
	delay2        time.Duration
	delay3        time.Duration
	answerDelay   time.Duration
	questionDelay time.Duration
	answerChan    chan bool
}

//ChannelBots ...
type ChannelBots struct {
	lock sync.RWMutex
	bots map[string]*Bot
}

//NewBot ...
func NewBot(rtm *slack.RTM, channel string) *Bot {
	ctx, cancel := context.WithCancel(context.Background())
	return &Bot{
		ctx:           ctx,
		answerChan:    make(chan bool),
		cancel:        cancel,
		rtm:           rtm,
		channel:       channel,
		delay1:        60 * time.Second,
		delay2:        40 * time.Second,
		delay3:        20 * time.Second,
		answerDelay:   15 * time.Second,
		questionDelay: 10 * time.Second,
	}
}

//NewChannelBots ...
func NewChannelBots() *ChannelBots {
	return &ChannelBots{
		bots: make(map[string]*Bot),
	}
}

//GetOrCreate ...
func (cb *ChannelBots) GetOrCreate(rtm *slack.RTM, channel string) *Bot {
	cb.lock.RLock()
	bot, ok := cb.bots[channel]
	cb.lock.RUnlock()
	if !ok {
		bot = NewBot(rtm, channel)
		cb.lock.Lock()
		cb.bots[channel] = bot
		cb.lock.Unlock()
	}
	return bot
}
