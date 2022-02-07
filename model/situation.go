package model

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Situation struct {
	Message       *tgbotapi.Message
	CallbackQuery *tgbotapi.CallbackQuery
	BotLang       string
	User          *User
	Command       string
	Params        *Parameters
	Err           error
}

type Parameters struct {
	ReplyText string
	Level     string
	Partition string
	Link      *LinkInfo
}

type LinkInfo struct {
	Url             string
	FileID          string
	Duration        int
	Limited         bool
	ImpressionsLeft int
}
