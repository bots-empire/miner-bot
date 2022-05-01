package auth

import (
	"github.com/Stepan1328/miner-bot/model"
	"github.com/bots-empire/base-bot/msgs"
)

type Auth struct {
	bot *model.GlobalBot

	msgs *msgs.Service
}

func NewAuthService(bot *model.GlobalBot, msgs *msgs.Service) *Auth {
	return &Auth{
		bot:  bot,
		msgs: msgs,
	}
}
