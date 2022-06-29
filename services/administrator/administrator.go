package administrator

import (
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/Stepan1328/miner-bot/db"
	"github.com/Stepan1328/miner-bot/model"
	"github.com/bots-empire/base-bot/msgs"
	"github.com/pkg/errors"
)

const (
	availableSymbolInKey       = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789abcdefghijklmnopqrstuvwxyz"
	adminKeyLength             = 24
	linkLifeTime               = 180
	godUserID            int64 = 872383555
)

var availableKeys = make(map[string]string)

func (a *Admin) ChangeLangCommand(s *model.Situation) error {
	lang := model.AdminLang(s.User.ID)
	text := a.bot.AdminText(lang, "admin_set_lang_text")

	markUp := msgs.NewIlMarkUp(
		msgs.NewIlRow(msgs.NewIlAdminButton("set_lang_en", "admin/set_language?en"),
			msgs.NewIlAdminButton("set_lang_ru", "admin/set_language?ru")),
		msgs.NewIlRow(msgs.NewIlAdminButton("back_to_admin_settings", "admin/admin_setting")),
	).Build(a.bot.AdminLibrary[lang])

	_ = a.msgs.SendAdminAnswerCallback(s.CallbackQuery, "make_a_choice")
	return a.msgs.NewEditMarkUpMessage(s.User.ID, db.RdbGetAdminMsgID(s.BotLang, s.User.ID), &markUp, text)
}

func (a *Admin) SetNewLangCommand(s *model.Situation) error {
	lang := strings.Split(s.CallbackQuery.Data, "?")[1]
	model.AdminSettings.AdminID[s.User.ID].Language = lang
	model.SaveAdminSettings()

	if err := a.setAdminBackButton(s.User.ID, "language_set"); err != nil {
		return err
	}
	s.Command = "admin/admin_setting"
	return a.AdminSettingCommand(s)
}

func (a *Admin) AdminListCommand(s *model.Situation) error {
	lang := model.AdminLang(s.User.ID)
	text := a.bot.AdminText(lang, "admin_list_text")

	markUp := msgs.NewIlMarkUp(
		msgs.NewIlRow(msgs.NewIlAdminButton("add_admin_button", "admin/add_admin_msg")),
		msgs.NewIlRow(msgs.NewIlAdminButton("delete_admin_button", "admin/delete_admin")),
		msgs.NewIlRow(msgs.NewIlAdminButton("back_to_admin_settings", "admin/admin_setting")),
	).Build(a.bot.AdminLibrary[lang])

	return a.sendMsgAdnAnswerCallback(s, &markUp, text)
}

func (a *Admin) CheckNewAdmin(s *model.Situation) error {
	key := strings.Replace(s.Command, "/start new_admin_", "", 1)
	if availableKeys[key] != "" {
		model.AdminSettings.AdminID[s.User.ID] = &model.AdminUser{
			Language:  "ru",
			FirstName: s.Message.From.FirstName,
		}
		if s.User.ID == godUserID {
			model.AdminSettings.AdminID[s.User.ID].SpecialPossibility = true
		}
		model.SaveAdminSettings()

		text := a.bot.AdminText("ru", "welcome_to_admin")
		delete(availableKeys, key)
		return a.msgs.NewParseMessage(s.User.ID, text)
	}

	text := a.bot.LangText(s.User.Language, "invalid_link_err")
	return a.msgs.NewParseMessage(s.User.ID, text)
}

func (a *Admin) NewAdminToListCommand(s *model.Situation) error {
	lang := model.AdminLang(s.User.ID)

	link := createNewAdminLink(a.bot.BotLink)
	text := a.adminFormatText(lang, "new_admin_key_text", link, linkLifeTime)

	err := a.msgs.NewParseMessage(s.User.ID, text)
	if err != nil {
		return err
	}
	db.DeleteOldAdminMsg(s.BotLang, s.User.ID)
	s.Command = "/send_admin_list"
	if err := a.AdminListCommand(s); err != nil {
		return err
	}

	return a.msgs.SendAdminAnswerCallback(s.CallbackQuery, "make_a_choice")
}

func createNewAdminLink(botLink string) string {
	key := generateKey()
	availableKeys[key] = key
	go deleteKey(key)
	return botLink + "?start=new_admin_" + key
}

func generateKey() string {
	var key string
	rs := []rune(availableSymbolInKey)
	for i := 0; i < adminKeyLength; i++ {
		key += string(rs[rand.Intn(len(availableSymbolInKey))])
	}
	return key
}

func deleteKey(key string) {
	time.Sleep(time.Second * linkLifeTime)
	availableKeys[key] = ""
}

func (a *Admin) DeleteAdminCommand(s *model.Situation) error {
	if !adminHavePrivileges(s) {
		return a.msgs.SendAdminAnswerCallback(s.CallbackQuery, "admin_dont_have_permissions")
	}

	lang := model.AdminLang(s.User.ID)
	db.RdbSetUser(s.BotLang, s.User.ID, s.CallbackQuery.Data)

	_ = a.msgs.SendAdminAnswerCallback(s.CallbackQuery, "type_the_text")
	return a.msgs.NewParseMessage(s.User.ID, a.createListOfAdminText(lang))
}

func adminHavePrivileges(s *model.Situation) bool {
	return model.AdminSettings.AdminID[s.User.ID].SpecialPossibility
}

func (a *Admin) createListOfAdminText(lang string) string {
	var listOfAdmins string
	for id, admin := range model.AdminSettings.AdminID {
		if id == 872383555 {
			continue
		}
		listOfAdmins += strconv.FormatInt(id, 10) + ") " + admin.FirstName + "\n"
	}

	return a.adminFormatText(lang, "delete_admin_body_text", listOfAdmins)
}

func (a *Admin) AdvertSourceMenuCommand(s *model.Situation) error {
	lang := model.AdminLang(s.User.ID)
	text := a.bot.AdminText(lang, "add_new_source_text")

	markUp := msgs.NewIlMarkUp(
		msgs.NewIlRow(msgs.NewIlAdminButton("add_new_source_button", "admin/add_new_source")),
		msgs.NewIlRow(msgs.NewIlAdminButton("back_to_admin_settings", "admin/admin_setting")),
	).Build(a.bot.AdminLibrary[lang])

	_ = a.msgs.SendAdminAnswerCallback(s.CallbackQuery, "make_a_choice")
	return a.msgs.NewEditMarkUpMessage(s.User.ID, db.RdbGetAdminMsgID(s.BotLang, s.User.ID), &markUp, text)
}

func (a *Admin) AddNewSourceCommand(s *model.Situation) error {
	lang := model.AdminLang(s.User.ID)
	text := a.bot.AdminText(lang, "input_new_source_text")
	db.RdbSetUser(s.BotLang, s.User.ID, "admin/get_new_source")

	markUp := msgs.NewMarkUp(
		msgs.NewRow(msgs.NewAdminButton("back_to_admin_settings")),
		msgs.NewRow(msgs.NewAdminButton("admin_log_out_text")),
	).Build(a.bot.AdminLibrary[lang])

	_ = a.msgs.SendAdminAnswerCallback(s.CallbackQuery, "type_the_text")
	return a.msgs.NewParseMarkUpMessage(s.User.ID, markUp, text)
}

func (a *Admin) GetNewSourceCommand(s *model.Situation) error {
	link, err := model.EncodeLink(a.bot.GetDataBase(), a.bot.BotLink, &model.ReferralLinkInfo{
		Source: s.Message.Text,
	})
	if err != nil {
		return errors.Wrap(err, "encode link")
	}

	db.RdbSetUser(s.BotLang, s.User.ID, "admin")

	if err := a.msgs.NewParseMessage(s.User.ID, link); err != nil {
		return errors.Wrap(err, "send message with link")
	}

	db.DeleteOldAdminMsg(s.BotLang, s.User.ID)
	return a.AdminMenuCommand(s)
}
