package msgs

import (
	"strings"

	"github.com/Stepan1328/miner-bot/assets"
	"github.com/Stepan1328/miner-bot/model"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	defaultNotificationBot = "it"
	currency               = "{{currency}}"
)

var (
	DeveloperID int64 = 872383555
)

func SendMessageToChat(botLang string, msg tgbotapi.MessageConfig) bool {
	if _, err := model.GetGlobalBot(botLang).Bot.Send(msg); err != nil {
		return false
	}
	return true
}

func SendMsgToChat(botLang string, msg tgbotapi.Chattable) bool {
	if _, err := model.GetGlobalBot(botLang).Bot.Send(msg); err != nil {
		return false
	}
	return true
}

func NewParseMessage(botLang string, chatID int64, text string) error {
	msg := tgbotapi.MessageConfig{
		BaseChat: tgbotapi.BaseChat{
			ChatID: chatID,
		},
		Text:      insertCurrency(botLang, text),
		ParseMode: "HTML",
	}

	return SendMsgToUser(botLang, msg)
}

func NewIDParseMessage(botLang string, chatID int64, text string) (int, error) {
	msg := tgbotapi.MessageConfig{
		BaseChat: tgbotapi.BaseChat{
			ChatID: chatID,
		},
		Text:      insertCurrency(botLang, text),
		ParseMode: "HTML",
	}

	message, err := model.GetGlobalBot(botLang).Bot.Send(msg)
	if err != nil {
		return 0, nil
	}
	return message.MessageID, nil
}

func NewParseMarkUpMessage(botLang string, chatID int64, markUp interface{}, text string) error {
	msg := tgbotapi.MessageConfig{
		BaseChat: tgbotapi.BaseChat{
			ChatID:      chatID,
			ReplyMarkup: markUp,
		},
		Text:      insertCurrency(botLang, text),
		ParseMode: "HTML",
	}

	return SendMsgToUser(botLang, msg)
}

func NewParseMarkUpPhotoMessage(botLang string, chatID int64, markUp interface{}, text string, photo tgbotapi.RequestFileData) error {
	msg := tgbotapi.PhotoConfig{
		BaseFile: tgbotapi.BaseFile{
			BaseChat: tgbotapi.BaseChat{
				ChatID:      chatID,
				ReplyMarkup: markUp,
			},
			File: photo},
		Caption:   insertCurrency(botLang, text),
		ParseMode: "HTML",
	}

	return SendMsgToUser(botLang, msg)
}

func NewParseMarkUpVideoMessage(botLang string, chatID int64, markUp interface{}, text string, video tgbotapi.RequestFileData) error {
	msg := tgbotapi.VideoConfig{
		BaseFile: tgbotapi.BaseFile{
			BaseChat: tgbotapi.BaseChat{
				ChatID:      chatID,
				ReplyMarkup: markUp,
			},
			File: video,
		},
		Caption:   insertCurrency(botLang, text),
		ParseMode: "HTML",
	}

	return SendMsgToUser(botLang, msg)
}

func NewIDParseMarkUpMessage(botLang string, chatID int64, markUp interface{}, text string) (int, error) {
	msg := tgbotapi.MessageConfig{
		BaseChat: tgbotapi.BaseChat{
			ChatID:      chatID,
			ReplyMarkup: markUp,
		},
		Text:                  insertCurrency(botLang, text),
		ParseMode:             "HTML",
		DisableWebPagePreview: true,
	}

	message, err := model.GetGlobalBot(botLang).Bot.Send(msg)
	if err != nil {
		return 0, err
	}
	return message.MessageID, nil
}

func NewEditMarkUpMessage(botLang string, userID int64, msgID int, markUp *tgbotapi.InlineKeyboardMarkup, text string) error {
	msg := tgbotapi.EditMessageTextConfig{
		BaseEdit: tgbotapi.BaseEdit{
			ChatID:      userID,
			MessageID:   msgID,
			ReplyMarkup: markUp,
		},
		Text:                  insertCurrency(botLang, text),
		ParseMode:             "HTML",
		DisableWebPagePreview: true,
	}

	return SendMsgToUser(botLang, msg)
}

func SendAnswerCallback(botLang string, callbackQuery *tgbotapi.CallbackQuery, lang, text string) error {
	answerCallback := tgbotapi.CallbackConfig{
		CallbackQueryID: callbackQuery.ID,
		Text:            assets.LangText(lang, text),
	}

	_ = SendMsgToUser(botLang, answerCallback)
	return nil
}

func SendAdminAnswerCallback(botLang string, callbackQuery *tgbotapi.CallbackQuery, text string) error {
	lang := assets.AdminLang(callbackQuery.From.ID)
	answerCallback := tgbotapi.CallbackConfig{
		CallbackQueryID: callbackQuery.ID,
		Text:            assets.AdminText(lang, text),
	}

	_ = SendMsgToUser(botLang, answerCallback)
	return nil
}

func SendSimpleMsg(botLang string, chatID int64, text string) error {
	msg := tgbotapi.NewMessage(chatID, insertCurrency(botLang, text))

	return SendMsgToUser(botLang, msg)
}

func SendMsgToUser(botLang string, msg tgbotapi.Chattable) error {
	if _, err := model.GetGlobalBot(botLang).Bot.Send(msg); err != nil {
		return err
	}
	return nil
}

func SendNotificationToDeveloper(text string) int {
	id, _ := NewIDParseMessage(defaultNotificationBot, DeveloperID, text)
	return id
}

func PinMsgToDeveloper(msgID int) {
	_ = SendMsgToUser(defaultNotificationBot, tgbotapi.PinChatMessageConfig{
		ChatID:    DeveloperID,
		MessageID: msgID,
	})
}

func insertCurrency(botLang string, text string) string {
	return strings.Replace(text, currency, assets.AdminSettings.GetCurrency(botLang), -1)
}
