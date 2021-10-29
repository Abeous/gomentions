package main

import (
	"bytes"
	"context"
	"html/template"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/MemeLabs/dggchat"
)

type mongoUser struct {
	Username  string
	Timestamp int64
	Discord   string
}

func isMod(user dggchat.User) bool {
	return user.HasFeature("moderator") || user.HasFeature("admin") || user.Nick == "abesus"
}

func isLastMSG(new string, last string) string {
	if new == last {
		return " ComfyCat"
	}
	return ""
}

// TODO
func (b *bot) sendMessage(m string, s *dggchat.Session) {
	dupe := isLastMSG(m, b.lastSent)
	msg := m + dupe
	b.lastSent = msg

	err := s.SendMessage(msg)
	if err != nil {
		log.Printf("[##] send error: %s\n", err.Error())
	}
}

func (b *bot) sendPM(u string, m string, s *dggchat.Session) {
	dupe := isLastMSG(m, b.lastSent)
	msg := m + dupe
	b.lastSent = msg

	err := s.SendPrivateMessage(u, msg)
	if err != nil {
		log.Printf("[##] send error: %s\n", err.Error())
	}
}

// !say - say a message
func (b *bot) say(m dggchat.Message, s *dggchat.Session) {
	if !isMod(m.Sender) || !strings.HasPrefix(m.Message, "!say") {
		return
	}

	// message itself can contain spaces
	parts := strings.SplitN(m.Message, " ", 2)
	if len(parts) != 2 {
		return
	}
	b.sendMessage(parts[1], s)
}

func (b *bot) isUserMention(m dggchat.Message, s *dggchat.Session) {
	if len(m.Entities.Nicks) > 0 {
		for _, opted := range b.optedUsers {
			if m.IsUserMentioned(opted) {
				err := b.sendtoDB(m)
				if err != nil {
					log.Printf("[##] send error: %s\n", err.Error())
				}
			}
		}
	}
}

func (b *bot) optIn(m dggchat.Message, s *dggchat.Session) {
	args := strings.Split(m.Message, " ")
	if strings.HasPrefix(m.Message, "!mentions") && len(args) > 1 && args[1] == "enable" {
		if b.isUserOpted(m.Sender.Nick) {
			b.sendPM(m.Sender.Nick, "You're already opted in silly!", s)
			return
		}
		discordId := ""
		// doesnt work probably
		if len(args) == 3 {
			discordId = args[2]
		}

		user := mongoUser{
			Username:  m.Sender.Nick,
			Timestamp: time.Now().Unix(),
			Discord:   discordId,
		}

		b.sendPM(m.Sender.Nick, "Opted in!", s)

		b.optedUsers = append(b.optedUsers, m.Sender.Nick)

		_, err := b.mongo.Database("mentions").Collection("optedusers").InsertOne(context.TODO(), user)
		if err != nil {
			log.Printf("[##] send error: %s\n", err.Error())
		}
	}
}

func (b *bot) optOut(m dggchat.Message, s *dggchat.Session) {
	args := strings.Split(m.Message, " ")
	if strings.HasPrefix(m.Message, "!mentions") && len(args) > 1 && args[1] == "disable" {
		// delete logs, delete user file, send message
	}
}

func (b *bot) getMentions(m dggchat.Message, s *dggchat.Session) {
	args := strings.Split(m.Message, " ")
	if strings.HasPrefix(m.Message, "!mentions") {
		limit := 5
		if len(args) > 1 {
			arglimit, err := strconv.Atoi(args[1])
			if err != nil {
				return
			}
			limit = arglimit
		}

		results, err := b.requestMentions(m, int64(limit))
		if err != nil {
			log.Printf("[##] send error: %s\n", err.Error())
			b.sendPM(m.Sender.Nick, "Failed getting mentions", s)
		}

		tmpl, err := template.ParseFiles("template.html")
		if err != nil {
			log.Printf("[##] send error: %s\n", err.Error())
		}

		var tpl bytes.Buffer
		page := NewMentionsPage(results)
		tmpl.Execute(&tpl, page)

		resp, err := UploadToHost(tpl.Bytes())
		if err != nil {
			log.Printf("[##] send error: %s\n", err.Error())
		}
		b.sendPM(m.Sender.Nick, resp, s)
	}
}
