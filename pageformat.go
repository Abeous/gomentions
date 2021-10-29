package main

import (
	"github.com/MemeLabs/dggchat"
)

type MentionsPage struct {
	Logs []dggchat.Message
}

func NewMentionsPage(logs []dggchat.Message) *MentionsPage {
	page := MentionsPage{
		Logs: logs,
	}

	return &page
}
