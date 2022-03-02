package administrator

import (
	"strings"

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

type SelectedLangCommand struct {
}

func NewSelectedLangCommand() *SelectedLangCommand {
	return &SelectedLangCommand{}
}

func (c *SelectedLangCommand) Serve(s *model.Situation) error {
	data := strings.Split(s.CallbackQuery.Data, "?")
	partition := data[1]
	lang := data[2]
	switch partition {
	case "switch_lang":
		switchLangOnKeyboard(lang)
		if err := sendMailingMenu(s.BotLang, s.User.ID); err != nil {
			return err
		}
		return msgs.SendAdminAnswerCallback(s.BotLang, s.CallbackQuery, "make_a_choice")
	case "switch_all":
		switchedSelectedLanguages()
		if err := sendMailingMenu(s.BotLang, s.User.ID); err != nil {
			return err
		}

		return msgs.SendAdminAnswerCallback(s.BotLang, s.CallbackQuery, "make_a_choice")
	}
	return nil
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

func switchLangOnKeyboard(lang string) {
	assets.AdminSettings.LangSelectedMap[lang] = !assets.AdminSettings.LangSelectedMap[lang]
	assets.SaveAdminSettings()
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

func switchedSelectedLanguages() {
	if selectedAllLanguage() {
		resetSelectedLang()
		return
	}
	chooseAllLanguages()
}

func resetSelectedLang() {
	for lang := range assets.AdminSettings.LangSelectedMap {
		assets.AdminSettings.LangSelectedMap[lang] = false
	}
	assets.SaveAdminSettings()
}

func chooseAllLanguages() {
	for lang := range assets.AdminSettings.LangSelectedMap {
		assets.AdminSettings.LangSelectedMap[lang] = true
	}
	assets.SaveAdminSettings()
}

func selectedAllLanguage() bool {
	for _, lang := range assets.AvailableLang {
		if !assets.AdminSettings.LangSelectedMap[lang] {
			return false
		}
	}
	return true
}
