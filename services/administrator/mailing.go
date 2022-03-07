package administrator

import (
	"github.com/Stepan1328/miner-bot/assets"
	"github.com/Stepan1328/miner-bot/db"
	"github.com/Stepan1328/miner-bot/model"
	"github.com/Stepan1328/miner-bot/msgs"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type StartMailingCommand struct {
}

func NewStartMailingCommand() *StartMailingCommand {
	return &StartMailingCommand{}
}

func (c *StartMailingCommand) Serve(s *model.Situation) error {
	go db.StartMailing(s.BotLang, s.User)
	_ = msgs.SendAdminAnswerCallback(s.BotLang, s.CallbackQuery, "mailing_successful")
	return resendAdvertisementMenuLevel(s.BotLang, s.User.ID)
}

func sendMailingMenu(botLang string, userID int64) error {
	lang := assets.AdminLang(userID)

	text := assets.AdminText(lang, "mailing_main_text")
	markUp := createMailingMarkUp(lang)

	if db.RdbGetAdminMsgID(botLang, userID) == 0 {
		msgID, err := msgs.NewIDParseMarkUpMessage(botLang, userID, &markUp, text)
		if err != nil {
			return err
		}

		db.RdbSetAdminMsgID(botLang, userID, msgID)
		return nil
	}
	return msgs.NewEditMarkUpMessage(botLang, userID, db.RdbGetAdminMsgID(botLang, userID), &markUp, text)
}

func createMailingMarkUp(lang string) tgbotapi.InlineKeyboardMarkup {
	markUp := &msgs.InlineMarkUp{}

	markUp.Rows = append(markUp.Rows,
		msgs.NewIlRow(msgs.NewIlAdminButton("start_mailing_button", "admin/start_mailing")),
		msgs.NewIlRow(msgs.NewIlAdminButton("back_to_advertisement_setting", "admin/advertisement")),
	)
	return markUp.Build(lang)
}

func resendAdvertisementMenuLevel(botLang string, userID int64) error {
	db.DeleteOldAdminMsg(botLang, userID)

	db.RdbSetUser(botLang, userID, "admin/advertisement")
	inlineMarkUp, text := getAdvertisementMenu(botLang, userID)
	msgID, err := msgs.NewIDParseMarkUpMessage(botLang, userID, inlineMarkUp, text)
	if err != nil {
		return err
	}
	db.RdbSetAdminMsgID(botLang, userID, msgID)
	return nil
}
