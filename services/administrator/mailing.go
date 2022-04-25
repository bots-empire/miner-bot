package administrator

import (
	"github.com/Stepan1328/miner-bot/assets"
	"github.com/Stepan1328/miner-bot/db"
	"github.com/Stepan1328/miner-bot/model"
	"github.com/Stepan1328/miner-bot/msgs"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"strconv"
	"strings"
)

type StartMailingCommand struct {
}

func NewStartMailingCommand() *StartMailingCommand {
	return &StartMailingCommand{}
}

func (c *StartMailingCommand) Serve(s *model.Situation) error {
	channel, _ := strconv.Atoi(strings.Split(s.CallbackQuery.Data, "?")[1])
	go db.StartMailing(s.BotLang, s.User, channel)

	_ = msgs.SendAdminAnswerCallback(s.BotLang, s.CallbackQuery, "mailing_successful")
	if channel == assets.GlobalMailing {
		return NewAdvertisementMenuCommand().Serve(s)
	}
	return resendAdvertisementMenuLevel(s.BotLang, s.User.ID, channel)
}

func sendMailingMenu(botLang string, userID int64, channel string) error {
	lang := assets.AdminLang(userID)

	text := assets.AdminText(lang, "mailing_main_text")
	markUp := createMailingMarkUp(botLang, lang, channel)

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

func createMailingMarkUp(botLang, lang, channel string) tgbotapi.InlineKeyboardMarkup {
	markUp := &msgs.InlineMarkUp{}

	if buttonUnderAdvertisementUnable(botLang) {
		markUp.Rows = append(markUp.Rows,
			msgs.NewIlRow(msgs.NewIlAdminButton("advert_button_on", "admin/change_advert_button_status?"+channel)),
		)
	} else {
		markUp.Rows = append(markUp.Rows,
			msgs.NewIlRow(msgs.NewIlAdminButton("advert_button_off", "admin/change_advert_button_status?"+channel)),
		)
	}

	if channel == "4" {
		markUp.Rows = append(markUp.Rows,
			msgs.NewIlRow(msgs.NewIlAdminButton("start_mailing_button", "admin/start_mailing?"+channel)),
			msgs.NewIlRow(msgs.NewIlAdminButton("back_to_chan_menu", "admin/advertisement")),
		)
	} else {
		markUp.Rows = append(markUp.Rows,
			msgs.NewIlRow(msgs.NewIlAdminButton("start_mailing_button", "admin/start_mailing?"+channel)),
			msgs.NewIlRow(msgs.NewIlAdminButton("back_to_advertisement_setting", "admin/change_advert_chan?"+channel)),
		)
	}

	return markUp.Build(lang)
}

func resendAdvertisementMenuLevel(botLang string, userID int64, channel int) error {
	db.DeleteOldAdminMsg(botLang, userID)

	db.RdbSetUser(botLang, userID, "admin/advertisement")
	inlineMarkUp, text := getAdvertisementMenu(botLang, userID, channel)
	msgID, err := msgs.NewIDParseMarkUpMessage(botLang, userID, inlineMarkUp, text)
	if err != nil {
		return err
	}
	db.RdbSetAdminMsgID(botLang, userID, msgID)
	return nil
}

func buttonUnderAdvertisementUnable(botLang string) bool {
	return assets.AdminSettings.GlobalParameters[botLang].Parameters.ButtonUnderAdvert
}
