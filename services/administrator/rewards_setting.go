package administrator

import (
	"strings"

	"github.com/Stepan1328/miner-bot/assets"
	"github.com/Stepan1328/miner-bot/db"
	"github.com/Stepan1328/miner-bot/model"
	"github.com/Stepan1328/miner-bot/msgs"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/pkg/errors"
)

const (
	bonusAmount         = "bonus_amount"
	minWithdrawalAmount = "min_withdrawal_amount"
	referralAmount      = "referral_amount"
	currencyType        = "currency_type"
)

type MakeMoneySettingCommand struct {
}

func NewMakeMoneySettingCommand() *MakeMoneySettingCommand {
	return &MakeMoneySettingCommand{}
}

func (c *MakeMoneySettingCommand) Serve(s *model.Situation) error {

	markUp, text := sendMakeMoneyMenu(s.BotLang, s.User.ID)

	if db.RdbGetAdminMsgID(s.BotLang, s.User.ID) != 0 {
		err := msgs.NewEditMarkUpMessage(s.BotLang, s.User.ID, db.RdbGetAdminMsgID(s.BotLang, s.User.ID), markUp, text)
		if err != nil {
			return errors.Wrap(err, "failed to edit markup")
		}
		err = msgs.SendAdminAnswerCallback(s.BotLang, s.CallbackQuery, "make_a_choice")
		if err != nil {
			return errors.Wrap(err, "failed to send admin answer callback")
		}
		return nil
	}
	msgID, err := msgs.NewIDParseMarkUpMessage(s.BotLang, s.User.ID, markUp, text)
	if err != nil {
		return errors.Wrap(err, "failed parse new id markup message")
	}
	db.RdbSetAdminMsgID(s.BotLang, s.User.ID, msgID)
	return nil
}

func sendMakeMoneyMenu(botLang string, userID int64) (*tgbotapi.InlineKeyboardMarkup, string) {
	lang := assets.AdminLang(userID)
	text := assets.AdminText(lang, "make_money_setting_text")

	markUp := msgs.NewIlMarkUp(
		msgs.NewIlRow(msgs.NewIlAdminButton("change_bonus_amount_button", "admin/make_money?bonus_amount")),
		msgs.NewIlRow(msgs.NewIlAdminButton("change_min_withdrawal_amount_button", "admin/make_money?min_withdrawal_amount")),
		msgs.NewIlRow(msgs.NewIlAdminButton("change_miner_settings_button", "admin/miner_settings")),
		msgs.NewIlRow(msgs.NewIlAdminButton("change_exchange_rate_button", "admin/exchange_rate")),
		msgs.NewIlRow(msgs.NewIlAdminButton("change_referral_amount_button", "admin/make_money?referral_amount")),
		msgs.NewIlRow(msgs.NewIlAdminButton("change_currency_type_button", "admin/make_money?currency_type")),
		msgs.NewIlRow(msgs.NewIlAdminButton("back_to_main_menu", "admin/send_menu")),
	).Build(lang)

	db.RdbSetUser(botLang, userID, "admin/make_money_settings")
	return &markUp, text
}

type ChangeParameterCommand struct {
}

func NewChangeParameterCommand() *ChangeParameterCommand {
	return &ChangeParameterCommand{}
}

func (c *ChangeParameterCommand) Serve(s *model.Situation) error {
	changeParameter := strings.Split(s.CallbackQuery.Data, "?")[1]

	lang := assets.AdminLang(s.User.ID)
	var parameter, text string
	var value interface{}

	db.RdbSetUser(s.BotLang, s.User.ID, "admin/make_money?"+changeParameter)

	switch changeParameter {
	case bonusAmount:
		parameter = assets.AdminText(lang, "change_bonus_amount_button")
		value = assets.AdminSettings.GetParams(s.BotLang).BonusAmount
	case minWithdrawalAmount:
		parameter = assets.AdminText(lang, "change_min_withdrawal_amount_button")
		value = assets.AdminSettings.GetParams(s.BotLang).MinWithdrawalAmount
	case referralAmount:
		parameter = assets.AdminText(lang, "change_referral_amount_button")
		value = assets.AdminSettings.GetParams(s.BotLang).ReferralAmount
	case currencyType:
		parameter = assets.AdminText(lang, "change_currency_type_button")
		value = assets.AdminSettings.GetParams(s.BotLang).Currency
	}

	text = adminFormatText(lang, "set_new_amount_text", parameter, value)
	_ = msgs.SendAdminAnswerCallback(s.BotLang, s.CallbackQuery, "type_the_text")
	markUp := msgs.NewMarkUp(
		msgs.NewRow(msgs.NewAdminButton("back_to_make_money_setting")),
		msgs.NewRow(msgs.NewAdminButton("exit")),
	).Build(lang)

	return msgs.NewParseMarkUpMessage(s.BotLang, s.User.ID, markUp, text)
}
