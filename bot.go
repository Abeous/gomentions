package main

import (
	"log"

	"github.com/MemeLabs/dggchat"

	"go.mongodb.org/mongo-driver/mongo"
)

type bot struct {
	log         []dggchat.Message
	maxLogLines int
	mongo       mongo.Client
	optedUsers  []string
	msgParsers  []func(m dggchat.Message, s *dggchat.Session)
	pmParsers   []func(m dggchat.PrivateMessage, s *dggchat.Session)
	lastSent    string
	authCookie  string
}

func newBot(authCookie string, maxLogLines int) *bot {
	if maxLogLines < 0 {
		maxLogLines = 0
	}

	b := bot{
		log:         make([]dggchat.Message, maxLogLines),
		maxLogLines: maxLogLines,
		authCookie:  authCookie,
	}
	return &b
}

func (b *bot) addMSGParser(p ...func(m dggchat.Message, s *dggchat.Session)) {
	b.msgParsers = append(b.msgParsers, p...)
}

func (b *bot) addPRIVMSGParser(p ...func(m dggchat.PrivateMessage, s *dggchat.Session)) {
	b.pmParsers = append(b.pmParsers, p...)
}

func (b *bot) onMessage(m dggchat.Message, s *dggchat.Session) {
	// remember maxLogLines messages
	if len(b.log) >= b.maxLogLines {
		b.log = b.log[1:]
	}
	b.log = append(b.log, m)

	log.Printf("%s: %s\n", m.Sender.Nick, m.Message)

	for _, p := range b.msgParsers {
		p(m, s)
	}
}

func (b *bot) onPMHandler(m dggchat.PrivateMessage, s *dggchat.Session) {
	log.Printf("[#] PM: %s: %s\n", m.User.Nick, m.Message)
	for _, p := range b.pmParsers {
		p(m, s)
	}
}

func (b *bot) onError(e string, s *dggchat.Session) {
	log.Printf("[#] error: '%s'\n", e)
}

func (b *bot) onMute(m dggchat.Mute, s *dggchat.Session) {
	log.Printf("[#] mute: '%s' by '%s'\n", m.Target.Nick, m.Sender.Nick)
}

func (b *bot) onUnmute(m dggchat.Mute, s *dggchat.Session) {
	log.Printf("[#] unmute: '%s' by '%s'\n", m.Target.Nick, m.Sender.Nick)
}

func (b *bot) onBan(m dggchat.Ban, s *dggchat.Session) {
	log.Printf("[#] ban: '%s' by '%s'\n", m.Target.Nick, m.Sender.Nick)
}

func (b *bot) onUnban(m dggchat.Ban, s *dggchat.Session) {
	log.Printf("[#] unban: '%s' by '%s'\n", m.Target.Nick, m.Sender.Nick)
}

func (b *bot) onSocketError(err error, s *dggchat.Session) {
	log.Printf("[#] socket error: '%s'\n", err.Error())
}

func (b *bot) isUserOpted(nick string) bool {
	for _, user := range b.optedUsers {
		if user == nick {
			return true
		}
	}
	return false
}
