package services

import (
	"encoding/json"
	"fmt"
	"runtime/debug"

	"github.com/Stepan1328/miner-bot/msgs"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/Stepan1328/miner-bot/log"
)

var (
	panicLogger = log.NewDefaultLogger().Prefix("panic cather")
)

func panicCather(botLang string, update *tgbotapi.Update) {
	msg := recover()
	if msg == nil {
		return
	}

	panicText := fmt.Sprintf("%s\npanic in backend: message = %s\n%s",
		botLang,
		msg,
		string(debug.Stack()),
	)
	panicLogger.Warn(panicText)

	msgs.SendNotificationToDeveloper(panicText)

	data, err := json.MarshalIndent(update, "", "  ")
	if err != nil {
		return
	}

	msgs.SendNotificationToDeveloper(string(data))
}
