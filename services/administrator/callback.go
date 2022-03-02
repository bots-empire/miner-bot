package administrator

import (
	"fmt"
	"strings"

	"github.com/Stepan1328/miner-bot/assets"
	"github.com/Stepan1328/miner-bot/db"
	"github.com/Stepan1328/miner-bot/model"
	"github.com/Stepan1328/miner-bot/msgs"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type AdminCallbackHandlers struct {
	Handlers map[string]model.Handler
}

func (h *AdminCallbackHandlers) GetHandler(command string) model.Handler {
	return h.Handlers[command]
}

func (h *AdminCallbackHandlers) Init() {
	//Admin Setting command
	h.OnCommand("/send_menu", NewAdminMenuCommand())
	h.OnCommand("/admin_setting", NewAdminSettingCommand())
	h.OnCommand("/change_language", NewChangeLangCommand())
	h.OnCommand("/set_language", NewSetNewLangCommand())
	h.OnCommand("/send_admin_list", NewAdminListCommand())
	h.OnCommand("/add_admin_msg", NewNewAdminToListCommand())
	h.OnCommand("/delete_admin", NewDeleteAdminCommand())
	h.OnCommand("/send_advert_source_menu", NewAdvertSourceMenuCommand())
	h.OnCommand("/add_new_source", NewAddNewSourceCommand())

	//Make Money Setting command
	h.OnCommand("/make_money_setting", NewMakeMoneySettingCommand())
	h.OnCommand("/make_money", NewChangeParameterCommand())

	//Mailing command
	h.OnCommand("/advertisement", NewAdvertisementMenuCommand())
	h.OnCommand("/change_url_menu", NewChangeUrlMenuCommand())
	h.OnCommand("/change_text_menu", NewChangeTextMenuCommand())
	h.OnCommand("/mailing_menu", NewMailingMenuCommand())
	//h.OnCommand("/change_advert_button_status", NewChangeUnderAdvertButtonCommand())
	h.OnCommand("/send_advertisement", NewSelectedLangCommand())
	h.OnCommand("/start_mailing", NewStartMailingCommand())

	//Send Statistic command
	h.OnCommand("/send_statistic", NewStatisticCommand())
}

func (h *AdminCallbackHandlers) OnCommand(command string, handler model.Handler) {
	h.Handlers[command] = handler
}

func CheckAdminCallback(s *model.Situation) error {
	if !ContainsInAdmin(s.User.ID) {
		return notAdmin(s.BotLang, s.User)
	}

	s.Command = strings.TrimLeft(s.Command, "admin")

	Handler := model.Bots[s.BotLang].AdminCallBackHandler.GetHandler(s.Command)
	if Handler != nil {
		return Handler.Serve(s)
	}
	return model.ErrCommandNotConverted
}

type AdminLoginCommand struct {
}

func NewAdminCommand() *AdminLoginCommand {
	return &AdminLoginCommand{}
}

func (c *AdminLoginCommand) Serve(s *model.Situation) error {
	if !ContainsInAdmin(s.User.ID) {
		return notAdmin(s.BotLang, s.User)
	}

	updateFirstNameInfo(s.Message)
	db.DeleteOldAdminMsg(s.BotLang, s.User.ID)

	if err := setAdminBackButton(s.BotLang, s.User.ID, "admin_log_in"); err != nil {
		return err
	}
	s.Command = "/send_menu"
	return NewAdminMenuCommand().Serve(s)
}

func ContainsInAdmin(userID int64) bool {
	for key := range assets.AdminSettings.AdminID {
		if key == userID {
			return true
		}
	}
	return false
}

func notAdmin(botLang string, user *model.User) error {
	text := assets.LangText(user.Language, "not_admin")
	return msgs.SendSimpleMsg(botLang, user.ID, text)
}

func updateFirstNameInfo(message *tgbotapi.Message) {
	userID := message.From.ID
	assets.AdminSettings.AdminID[userID].FirstName = message.From.FirstName
	assets.SaveAdminSettings()
}

func setAdminBackButton(botLang string, userID int64, key string) error {
	lang := assets.AdminLang(userID)
	text := assets.AdminText(lang, key)

	markUp := msgs.NewMarkUp(
		msgs.NewRow(msgs.NewAdminButton("admin_log_out_text")),
	).Build(lang)

	return msgs.NewParseMarkUpMessage(botLang, userID, markUp, text)
}

type AdminMenuCommand struct {
}

func NewAdminMenuCommand() *AdminMenuCommand {
	return &AdminMenuCommand{}
}

func (c *AdminMenuCommand) Serve(s *model.Situation) error {
	db.RdbSetUser(s.BotLang, s.User.ID, "admin")
	lang := assets.AdminLang(s.User.ID)
	text := assets.AdminText(lang, "admin_main_menu_text")

	markUp := msgs.NewIlMarkUp(
		msgs.NewIlRow(msgs.NewIlAdminButton("setting_admin_button", "admin/admin_setting")),
		msgs.NewIlRow(msgs.NewIlAdminButton("setting_make_money_button", "admin/make_money_setting")),
		msgs.NewIlRow(msgs.NewIlAdminButton("setting_advertisement_button", "admin/advertisement")),
		msgs.NewIlRow(msgs.NewIlAdminButton("setting_statistic_button", "admin/send_statistic")),
	).Build(lang)

	if db.RdbGetAdminMsgID(s.BotLang, s.User.ID) != 0 {
		_ = msgs.SendAdminAnswerCallback(s.BotLang, s.CallbackQuery, "make_a_choice")
		return msgs.NewEditMarkUpMessage(
			s.BotLang,
			s.User.ID,
			db.RdbGetAdminMsgID(s.BotLang, s.User.ID),
			&markUp,
			text,
		)
	}
	msgID, err := msgs.NewIDParseMarkUpMessage(s.BotLang, s.User.ID, markUp, text)
	if err != nil {
		return err
	}
	db.RdbSetAdminMsgID(s.BotLang, s.User.ID, msgID)
	return nil
}

type AdminSettingCommand struct {
}

func NewAdminSettingCommand() *AdminSettingCommand {
	return &AdminSettingCommand{}
}

func (c *AdminSettingCommand) Serve(s *model.Situation) error {
	if strings.Contains(s.Params.Level, "delete_admin") {
		if err := setAdminBackButton(s.BotLang, s.User.ID, "operation_canceled"); err != nil {
			return err
		}
		db.DeleteOldAdminMsg(s.BotLang, s.User.ID)
	}

	db.RdbSetUser(s.BotLang, s.User.ID, "admin/mailing")
	lang := assets.AdminLang(s.User.ID)
	text := assets.AdminText(lang, "admin_setting_text")

	markUp := msgs.NewIlMarkUp(
		msgs.NewIlRow(msgs.NewIlAdminButton("setting_language_button", "admin/change_language")),
		msgs.NewIlRow(msgs.NewIlAdminButton("admin_list_button", "admin/send_admin_list")),
		msgs.NewIlRow(msgs.NewIlAdminButton("advertisement_source_button", "admin/send_advert_source_menu")),
		msgs.NewIlRow(msgs.NewIlAdminButton("back_to_main_menu", "admin/send_menu")),
	).Build(lang)
	if err := sendMsgAdnAnswerCallback(s, &markUp, text); err != nil {
		return err
	}
	return msgs.SendAdminAnswerCallback(s.BotLang, s.CallbackQuery, "make_a_choice")
}

type AdvertisementMenuCommand struct {
}

func NewAdvertisementMenuCommand() *AdvertisementMenuCommand {
	return &AdvertisementMenuCommand{}
}

func (c *AdvertisementMenuCommand) Serve(s *model.Situation) error {
	if strings.Contains(s.Params.Level, "change_text_url?") {
		if err := setAdminBackButton(s.BotLang, s.User.ID, "operation_canceled"); err != nil {
			return err
		}
		db.DeleteOldAdminMsg(s.BotLang, s.User.ID)
	}

	markUp, text := getAdvertisementMenu(s.BotLang, s.User.ID)
	msgID := db.RdbGetAdminMsgID(s.BotLang, s.User.ID)
	if msgID == 0 {
		var err error
		msgID, err = msgs.NewIDParseMarkUpMessage(s.BotLang, s.User.ID, markUp, text)
		if err != nil {
			return err
		}

		db.RdbSetAdminMsgID(s.BotLang, s.User.ID, msgID)
	} else {
		if err := msgs.NewEditMarkUpMessage(s.BotLang, s.User.ID, msgID, markUp, text); err != nil {
			return err
		}
	}

	if s.CallbackQuery != nil {
		if s.CallbackQuery.ID != "" {
			if err := msgs.SendAdminAnswerCallback(s.BotLang, s.CallbackQuery, "make_a_choice"); err != nil {
				return err
			}
		}
	}
	return nil
}

func getAdvertisementMenu(botLang string, userID int64) (*tgbotapi.InlineKeyboardMarkup, string) {
	lang := assets.AdminLang(userID)
	text := assets.AdminText(lang, "advertisement_setting_text")

	markUp := msgs.NewIlMarkUp(
		msgs.NewIlRow(msgs.NewIlAdminButton("change_url_button", "admin/change_url_menu")),
		msgs.NewIlRow(msgs.NewIlAdminButton("change_text_button", "admin/change_text_menu")),
		msgs.NewIlRow(msgs.NewIlAdminButton("distribute_button", "admin/mailing_menu")),
		msgs.NewIlRow(msgs.NewIlAdminButton("back_to_main_menu", "admin/send_menu")),
	).Build(lang)

	db.RdbSetUser(botLang, userID, "admin/advertisement")
	return &markUp, text
}

type ChangeUrlMenuCommand struct {
}

func NewChangeUrlMenuCommand() *ChangeUrlMenuCommand {
	return &ChangeUrlMenuCommand{}
}

func (c *ChangeUrlMenuCommand) Serve(s *model.Situation) error {
	key := "set_new_url_text"
	value := assets.AdminSettings.AdvertisingChan[s.BotLang].Url

	db.RdbSetUser(s.BotLang, s.User.ID, "admin/change_text_url?change_url")
	if err := promptForInput(s.BotLang, s.User.ID, key, value); err != nil {
		return err
	}
	return msgs.SendAdminAnswerCallback(s.BotLang, s.CallbackQuery, "type_the_text")
}

type ChangeTextMenuCommand struct {
}

func NewChangeTextMenuCommand() *ChangeTextMenuCommand {
	return &ChangeTextMenuCommand{}
}

func (c *ChangeTextMenuCommand) Serve(s *model.Situation) error {
	key := "set_new_advertisement_text"
	value := assets.AdminSettings.AdvertisingText[s.BotLang]

	db.RdbSetUser(s.BotLang, s.User.ID, "admin/change_text_url?change_text")
	if err := promptForInput(s.BotLang, s.User.ID, key, value); err != nil {
		return err
	}
	return msgs.SendAdminAnswerCallback(s.BotLang, s.CallbackQuery, "type_the_text")
}

type MailingMenuCommand struct {
}

func NewMailingMenuCommand() *MailingMenuCommand {
	return &MailingMenuCommand{}
}

func (c *MailingMenuCommand) Serve(s *model.Situation) error {
	db.RdbSetUser(s.BotLang, s.User.ID, "admin/mailing")
	resetSelectedLang()
	_ = msgs.SendAdminAnswerCallback(s.BotLang, s.CallbackQuery, "make_a_choice")
	return sendMailingMenu(s.BotLang, s.User.ID)
}

func promptForInput(botLang string, userID int64, key string, values ...interface{}) error {
	lang := assets.AdminLang(userID)

	text := adminFormatText(lang, key, values...)
	markUp := msgs.NewMarkUp(
		msgs.NewRow(msgs.NewAdminButton("back_to_advertisement_setting")),
		msgs.NewRow(msgs.NewAdminButton("exit")),
	).Build(lang)

	return msgs.NewParseMarkUpMessage(botLang, userID, markUp, text)
}

type StatisticCommand struct {
}

func NewStatisticCommand() *StatisticCommand {
	return &StatisticCommand{}
}

func (c *StatisticCommand) Serve(s *model.Situation) error {
	lang := assets.AdminLang(s.User.ID)

	count := countUsers(s.BotLang)
	allCount := countAllUsers()
	referrals := countReferrals(s.BotLang, count)
	//lastDayUsers := countUserFromLastDay(s.BotLang)
	blocked := countBlockedUsers(s.BotLang)
	subscribers := countSubscribers(s.BotLang)
	text := adminFormatText(lang, "statistic_text",
		allCount, count, referrals, blocked, subscribers, count-blocked)

	if err := msgs.NewParseMessage(s.BotLang, s.User.ID, text); err != nil {
		return err
	}
	db.DeleteOldAdminMsg(s.BotLang, s.User.ID)
	if err := NewAdminMenuCommand().Serve(s); err != nil {
		return err
	}

	return msgs.SendAdminAnswerCallback(s.BotLang, s.CallbackQuery, "make_a_choice")
}

func adminFormatText(lang, key string, values ...interface{}) string {
	formatText := assets.AdminText(lang, key)
	return fmt.Sprintf(formatText, values...)
}

func sendMsgAdnAnswerCallback(s *model.Situation, markUp *tgbotapi.InlineKeyboardMarkup, text string) error {
	if db.RdbGetAdminMsgID(s.BotLang, s.User.ID) != 0 {
		return msgs.NewEditMarkUpMessage(s.BotLang, s.User.ID, db.RdbGetAdminMsgID(s.BotLang, s.User.ID), markUp, text)
	}
	msgID, err := msgs.NewIDParseMarkUpMessage(s.BotLang, s.User.ID, markUp, text)
	if err != nil {
		return err
	}
	db.RdbSetAdminMsgID(s.BotLang, s.User.ID, msgID)

	if s.CallbackQuery != nil {
		if s.CallbackQuery.ID != "" {
			return msgs.SendAdminAnswerCallback(s.BotLang, s.CallbackQuery, "make_a_choice")
		}
	}
	return nil
}
