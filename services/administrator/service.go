package administrator

import (
	"github.com/Stepan1328/miner-bot/model"
	"github.com/bots-empire/base-bot/mailing"
	"github.com/bots-empire/base-bot/msgs"
)

type Admin struct {
	bot *model.GlobalBot

	mailing *mailing.Service
	msgs    *msgs.Service
}

func NewAdminService(bot *model.GlobalBot, mailing *mailing.Service, msgs *msgs.Service) *Admin {
	return &Admin{
		bot:     bot,
		mailing: mailing,
		msgs:    msgs,
	}
}
