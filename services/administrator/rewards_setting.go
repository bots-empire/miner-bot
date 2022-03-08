package administrator

import (
	"fmt"
	"strconv"
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
	maxOfClickPDAmount  = "max_click_pd"
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
		msgs.NewIlRow(msgs.NewIlAdminButton("change_bonus_amount_button", "admin/make_money?"+bonusAmount)),
		msgs.NewIlRow(msgs.NewIlAdminButton("change_min_withdrawal_amount_button", "admin/make_money?"+minWithdrawalAmount)),
		msgs.NewIlRow(msgs.NewIlAdminButton("change_max_of_click_pd_button", "admin/make_money?"+maxOfClickPDAmount)),
		msgs.NewIlRow(msgs.NewIlAdminButton("change_miner_settings_button", "admin/miner_settings")),
		msgs.NewIlRow(msgs.NewIlAdminButton("change_exchange_rate_button", "admin/exchange_rate")),
		msgs.NewIlRow(msgs.NewIlAdminButton("change_referral_amount_button", "admin/make_money?"+referralAmount)),
		msgs.NewIlRow(msgs.NewIlAdminButton("change_currency_type_button", "admin/make_money?"+currencyType)),
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
	case maxOfClickPDAmount:
		parameter = assets.AdminText(lang, "change_max_of_click_pd_button")
		value = assets.AdminSettings.GetParams(s.BotLang).MaxOfClickPerDay
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

type MinerSettingCommand struct {
}

func NewMinerSettingCommand() *MinerSettingCommand {
	return &MinerSettingCommand{}
}

func (c *MinerSettingCommand) Serve(s *model.Situation) error {
	return sendMinerSettingMenu(s)
}

func sendMinerSettingMenu(s *model.Situation) error {
	lang := assets.AdminLang(s.User.ID)
	text := adminFormatText(lang, "miner_setting_text")
	markUp := getMinerSettingMenu(s.BotLang, s.User.ID)

	msgID := db.RdbGetAdminMsgID(s.BotLang, s.User.ID)
	if msgID == 0 {
		id, err := msgs.NewIDParseMarkUpMessage(s.BotLang, s.User.ID, markUp, text)
		if err != nil {
			return err
		}

		db.RdbSetAdminMsgID(s.BotLang, s.User.ID, id)
		return nil
	}

	return msgs.NewEditMarkUpMessage(s.BotLang, s.User.ID, msgID, markUp, text)
}

func getMinerSettingMenu(botLang string, userID int64) *tgbotapi.InlineKeyboardMarkup {
	level := db.RdbGetMinerLevelSetting(botLang, userID)

	clickAmount := assets.AdminSettings.GetClickAmount(botLang, level)
	upgradeCost := assets.AdminSettings.GetUpgradeCost(botLang, level)

	markUp := msgs.NewIlMarkUp(
		msgs.NewIlRow(msgs.NewIlAdminButton("hash_per_click", "admin/not_clickable")),
		msgs.NewIlRow(
			msgs.NewIlCustomButton("-5", "admin/change_click_amount?dec&5"),
			msgs.NewIlCustomButton("-1", "admin/change_click_amount?dec&1"),
			msgs.NewIlCustomButton(strconv.Itoa(clickAmount), "admin/not_clickable"),
			msgs.NewIlCustomButton("+1", "admin/change_click_amount?inc&1"),
			msgs.NewIlCustomButton("+5", "admin/change_click_amount?inc&5")),

		msgs.NewIlRow(msgs.NewIlAdminButton("level_cost_button", "admin/not_clickable")),
		msgs.NewIlRow(
			msgs.NewIlCustomButton("-50", "admin/change_upgrade_amount?dec&50"),
			msgs.NewIlCustomButton("-10", "admin/change_upgrade_amount?dec&10"),
			msgs.NewIlCustomButton(strconv.Itoa(upgradeCost), "admin/not_clickable"),
			msgs.NewIlCustomButton("+10", "admin/change_upgrade_amount?inc&10"),
			msgs.NewIlCustomButton("+50", "admin/change_upgrade_amount?inc&50")),

		msgs.NewIlRow(
			msgs.NewIlCustomButton("<<", "admin/change_miner_level?dec"),
			msgs.NewIlCustomButton(strconv.Itoa(level+1), "admin/not_clickable"),
			msgs.NewIlCustomButton(">>", "admin/change_miner_level?inc")),

		msgs.NewIlRow(
			msgs.NewIlAdminButton("delete_miner_level", "admin/remove_miner_lvl"),
			msgs.NewIlAdminButton("add_new_miner_level", "admin/add_miner_lvl")),

		msgs.NewIlRow(msgs.NewIlAdminButton("back_to_make_money_setting", "admin/make_money_setting")),
	).Build(assets.AdminLang(userID))

	return &markUp
}

type ChangeClickAmountButton struct {
}

func NewChangeClickAmountButton() *ChangeClickAmountButton {
	return &ChangeClickAmountButton{}
}

func (c *ChangeClickAmountButton) Serve(s *model.Situation) error {
	level := db.RdbGetMinerLevelSetting(s.BotLang, s.User.ID)

	allParams := strings.Split(s.CallbackQuery.Data, "?")[1]
	changeParams := strings.Split(allParams, "&")
	operation := changeParams[0]
	value, _ := strconv.Atoi(changeParams[1])

	switch operation {
	case "inc":
		assets.AdminSettings.GetParams(s.BotLang).ClickAmount[level] += value
	case "dec":
		if assets.AdminSettings.GetParams(s.BotLang).ClickAmount[level]-value < 1 {
			_ = msgs.SendAdminAnswerCallback(s.BotLang, s.CallbackQuery, "already_min_value")
			return nil
		}
		assets.AdminSettings.GetParams(s.BotLang).ClickAmount[level] -= value
	}

	assets.SaveAdminSettings()
	return sendMinerSettingMenu(s)
}

type ChangeUpgradeAmountButton struct {
}

func NewChangeUpgradeAmountButton() *ChangeUpgradeAmountButton {
	return &ChangeUpgradeAmountButton{}
}

func (c *ChangeUpgradeAmountButton) Serve(s *model.Situation) error {
	level := db.RdbGetMinerLevelSetting(s.BotLang, s.User.ID)

	allParams := strings.Split(s.CallbackQuery.Data, "?")[1]
	changeParams := strings.Split(allParams, "&")
	operation := changeParams[0]
	value, _ := strconv.Atoi(changeParams[1])

	switch operation {
	case "inc":
		assets.AdminSettings.GetParams(s.BotLang).UpgradeMinerCost[level] += value
	case "dec":
		if assets.AdminSettings.GetParams(s.BotLang).UpgradeMinerCost[level]-value < 1 {
			_ = msgs.SendAdminAnswerCallback(s.BotLang, s.CallbackQuery, "already_min_value")
			return nil
		}
		assets.AdminSettings.GetParams(s.BotLang).UpgradeMinerCost[level] -= value
	}

	assets.SaveAdminSettings()
	return sendMinerSettingMenu(s)
}

type ChangeMinerLvlButton struct {
}

func NewChangeMinerLvlButton() *ChangeMinerLvlButton {
	return &ChangeMinerLvlButton{}
}

func (c *ChangeMinerLvlButton) Serve(s *model.Situation) error {
	level := db.RdbGetMinerLevelSetting(s.BotLang, s.User.ID)

	operation := strings.Split(s.CallbackQuery.Data, "?")[1]
	switch operation {
	case "inc":
		if level == assets.AdminSettings.GetMaxMinerLevel(s.BotLang)-1 {
			_ = msgs.SendAdminAnswerCallback(s.BotLang, s.CallbackQuery, "already_max_level")
			return nil
		}
		level++
	case "dec":
		if level == 0 {
			_ = msgs.SendAdminAnswerCallback(s.BotLang, s.CallbackQuery, "already_min_level")
			return nil
		}
		level--
	}

	db.RdbSetMinerLevelSetting(s.BotLang, s.User.ID, level)
	return sendMinerSettingMenu(s)
}

type DeleteMinerLevelButton struct {
}

func NewDeleteMinerLevelButton() *DeleteMinerLevelButton {
	return &DeleteMinerLevelButton{}
}

func (c *DeleteMinerLevelButton) Serve(s *model.Situation) error {
	level := db.RdbGetMinerLevelSetting(s.BotLang, s.User.ID)

	if assets.AdminSettings.GetMaxMinerLevel(s.BotLang) == 1 {

		//TODO: нельзя удалить уровень
		fmt.Println("hello")

		return nil
	}

	assets.AdminSettings.DeleteMinerLevel(s.BotLang, level)

	if level == assets.AdminSettings.GetMaxMinerLevel(s.BotLang) {
		db.RdbSetMinerLevelSetting(s.BotLang, s.User.ID, level-1)
	}

	assets.SaveAdminSettings()
	return sendMinerSettingMenu(s)
}

type AddMinerLevelButton struct {
}

func NewAddMinerLevelButton() *AddMinerLevelButton {
	return &AddMinerLevelButton{}
}

func (c *AddMinerLevelButton) Serve(s *model.Situation) error {
	level := db.RdbGetMinerLevelSetting(s.BotLang, s.User.ID)
	assets.AdminSettings.AddMinerLevel(s.BotLang, level)

	db.RdbSetMinerLevelSetting(s.BotLang, s.User.ID, level+1)
	assets.SaveAdminSettings()
	return sendMinerSettingMenu(s)
}

type ExchangerSettingCommand struct {
}

func NewExchangerSettingCommand() *ExchangerSettingCommand {
	return &ExchangerSettingCommand{}
}

func (c *ExchangerSettingCommand) Serve(s *model.Situation) error {
	return sendExchangerSettingMenu(s)
}

func sendExchangerSettingMenu(s *model.Situation) error {
	lang := assets.AdminLang(s.User.ID)
	text := adminFormatText(lang, "exchanger_setting_text")
	markUp := getExchangerSettingMenu(s.BotLang, s.User.ID)

	msgID := db.RdbGetAdminMsgID(s.BotLang, s.User.ID)
	if msgID == 0 {
		id, err := msgs.NewIDParseMarkUpMessage(s.BotLang, s.User.ID, markUp, text)
		if err != nil {
			return err
		}

		db.RdbSetAdminMsgID(s.BotLang, s.User.ID, id)
		return nil
	}

	return msgs.NewEditMarkUpMessage(s.BotLang, s.User.ID, msgID, markUp, text)
}

func getExchangerSettingMenu(botLang string, userID int64) *tgbotapi.InlineKeyboardMarkup {
	hashToBTC := assets.AdminSettings.GetParams(botLang).ExchangeHashToBTC
	btcToCurrency := int(assets.AdminSettings.GetParams(botLang).ExchangeBTCToCurrency)

	markUp := msgs.NewIlMarkUp(
		msgs.NewIlRow(msgs.NewIlAdminButton("hash_to_btc_button", "admin/not_clickable")),
		msgs.NewIlRow(
			msgs.NewIlCustomButton("-5", "admin/change_hash_to_btc_rate?dec&5"),
			msgs.NewIlCustomButton("-1", "admin/change_hash_to_btc_rate?dec&1"),
			msgs.NewIlCustomButton(strconv.Itoa(hashToBTC), "admin/not_clickable"),
			msgs.NewIlCustomButton("+1", "admin/change_hash_to_btc_rate?inc&1"),
			msgs.NewIlCustomButton("+5", "admin/change_hash_to_btc_rate?inc&5")),

		msgs.NewIlRow(msgs.NewIlAdminButton("btc_to_currency_button", "admin/not_clickable")),
		msgs.NewIlRow(
			msgs.NewIlCustomButton("-5", "admin/change_btc_to_currency_rate?dec&5"),
			msgs.NewIlCustomButton("-1", "admin/change_btc_to_currency_rate?dec&1"),
			msgs.NewIlCustomButton(strconv.Itoa(btcToCurrency), "admin/not_clickable"),
			msgs.NewIlCustomButton("+1", "admin/change_btc_to_currency_rate?inc&1"),
			msgs.NewIlCustomButton("+5", "admin/change_btc_to_currency_rate?inc&5")),

		msgs.NewIlRow(msgs.NewIlAdminButton("back_to_make_money_setting", "admin/make_money_setting")),
	).Build(assets.AdminLang(userID))

	return &markUp
}

type ChangeHashToBTCRateButton struct {
}

func NewChangeHashToBTCRateButton() *ChangeHashToBTCRateButton {
	return &ChangeHashToBTCRateButton{}
}

func (c *ChangeHashToBTCRateButton) Serve(s *model.Situation) error {
	allParams := strings.Split(s.CallbackQuery.Data, "?")[1]
	changeParams := strings.Split(allParams, "&")
	operation := changeParams[0]
	value, _ := strconv.Atoi(changeParams[1])

	switch operation {
	case "inc":
		assets.AdminSettings.GetParams(s.BotLang).ExchangeHashToBTC += value
	case "dec":
		if assets.AdminSettings.GetParams(s.BotLang).ExchangeHashToBTC-value < 1 {
			_ = msgs.SendAdminAnswerCallback(s.BotLang, s.CallbackQuery, "already_min_value")
			return nil
		}
		assets.AdminSettings.GetParams(s.BotLang).ExchangeHashToBTC -= value
	}

	assets.SaveAdminSettings()
	return sendExchangerSettingMenu(s)
}

type ChangeBTCToCurrencyRateButton struct {
}

func NewChangeBTCToCurrencyRateButton() *ChangeBTCToCurrencyRateButton {
	return &ChangeBTCToCurrencyRateButton{}
}

func (c *ChangeBTCToCurrencyRateButton) Serve(s *model.Situation) error {
	allParams := strings.Split(s.CallbackQuery.Data, "?")[1]
	changeParams := strings.Split(allParams, "&")
	operation := changeParams[0]
	value, _ := strconv.Atoi(changeParams[1])

	switch operation {
	case "inc":
		assets.AdminSettings.GetParams(s.BotLang).ExchangeBTCToCurrency += float64(value)
	case "dec":
		if int(assets.AdminSettings.GetParams(s.BotLang).ExchangeBTCToCurrency)-value < 1 {
			_ = msgs.SendAdminAnswerCallback(s.BotLang, s.CallbackQuery, "already_min_value")
			return nil
		}
		assets.AdminSettings.GetParams(s.BotLang).ExchangeBTCToCurrency -= float64(value)
	}

	assets.SaveAdminSettings()
	return sendExchangerSettingMenu(s)
}

type NotClickableButton struct {
}

func NewNotClickableButton() *NotClickableButton {
	return &NotClickableButton{}
}

func (c *NotClickableButton) Serve(s *model.Situation) error {
	_ = msgs.SendAdminAnswerCallback(s.BotLang, s.CallbackQuery, "not_clickable_button")
	return nil
}
