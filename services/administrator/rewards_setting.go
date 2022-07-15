package administrator

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Stepan1328/miner-bot/db"
	"github.com/Stepan1328/miner-bot/model"
	"github.com/bots-empire/base-bot/msgs"
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

func (a *Admin) MakeMoneySettingCommand(s *model.Situation) error {

	markUp, text := a.sendMakeMoneyMenu(s.BotLang, s.User.ID)

	if db.RdbGetAdminMsgID(s.BotLang, s.User.ID) != 0 {
		err := a.msgs.NewEditMarkUpMessage(s.User.ID, db.RdbGetAdminMsgID(s.BotLang, s.User.ID), markUp, text)
		if err != nil {
			return errors.Wrap(err, "failed to edit markup")
		}
		err = a.msgs.SendAdminAnswerCallback(s.CallbackQuery, "make_a_choice")
		if err != nil {
			return errors.Wrap(err, "failed to send admin answer callback")
		}
		return nil
	}
	msgID, err := a.msgs.NewIDParseMarkUpMessage(s.User.ID, markUp, text)
	if err != nil {
		return errors.Wrap(err, "failed parse new id markup message")
	}
	db.RdbSetAdminMsgID(s.BotLang, s.User.ID, msgID)
	return nil
}

func (a *Admin) sendMakeMoneyMenu(botLang string, userID int64) (*tgbotapi.InlineKeyboardMarkup, string) {
	lang := model.AdminLang(userID)
	text := a.bot.AdminText(lang, "make_money_setting_text")

	markUp := msgs.NewIlMarkUp(
		msgs.NewIlRow(msgs.NewIlAdminButton("change_bonus_amount_button", "admin/make_money?"+bonusAmount)),
		msgs.NewIlRow(msgs.NewIlAdminButton("change_min_withdrawal_amount_button", "admin/make_money?"+minWithdrawalAmount)),
		msgs.NewIlRow(msgs.NewIlAdminButton("change_max_of_click_pd_button", "admin/make_money?"+maxOfClickPDAmount)),
		msgs.NewIlRow(msgs.NewIlAdminButton("change_miner_settings_button", "admin/miner_settings")),
		msgs.NewIlRow(msgs.NewIlAdminButton("change_exchange_rate_button", "admin/exchange_rate")),
		msgs.NewIlRow(msgs.NewIlAdminButton("change_change_top_amount_button", "admin/change_top_amount_settings")),
		msgs.NewIlRow(msgs.NewIlAdminButton("change_referral_amount_button", "admin/make_money?"+referralAmount)),
		msgs.NewIlRow(msgs.NewIlAdminButton("change_currency_type_button", "admin/make_money?"+currencyType)),
		msgs.NewIlRow(msgs.NewIlAdminButton("back_to_main_menu", "admin/send_menu")),
	).Build(a.bot.AdminLibrary[lang])

	db.RdbSetUser(botLang, userID, "admin/make_money_settings")
	return &markUp, text
}

func (a *Admin) ChangeParameterCommand(s *model.Situation) error {
	changeParameter := strings.Split(s.CallbackQuery.Data, "?")[1]

	lang := model.AdminLang(s.User.ID)
	var parameter, text string
	var value interface{}

	db.RdbSetUser(s.BotLang, s.User.ID, "admin/make_money?"+changeParameter)

	switch changeParameter {
	case bonusAmount:
		parameter = a.bot.AdminText(lang, "change_bonus_amount_button")
		value = model.AdminSettings.GetParams(s.BotLang).BonusAmount
	case minWithdrawalAmount:
		parameter = a.bot.AdminText(lang, "change_min_withdrawal_amount_button")
		value = model.AdminSettings.GetParams(s.BotLang).MinWithdrawalAmount
	case maxOfClickPDAmount:
		parameter = a.bot.AdminText(lang, "change_max_of_click_pd_button")
		value = model.AdminSettings.GetParams(s.BotLang).MaxOfClickPerDay
	case referralAmount:
		db.RdbSetUser(s.BotLang, s.User.ID, "admin")

		reward, err := RdbGetRewardGap(s.BotLang, s.User.ID)
		if err != nil {
			return err
		}
		if reward == nil {
			reward = model.AdminSettings.GetParams(s.BotLang).ReferralReward.GetGapByIndex(1, 1)
			err = RdbSetRewardGap(s.BotLang, s.User.ID, reward)
			if err != nil {
				return err
			}
		}

		markUp, text := a.rewardsMarkUpAndText(s.User.ID, reward)
		return a.sendMsgAdnAnswerCallback(s, markUp, text)
	case currencyType:
		parameter = a.bot.AdminText(lang, "change_currency_type_button")
		value = model.AdminSettings.GetParams(s.BotLang).Currency
	}

	text = a.adminFormatText(lang, "set_new_amount_text", parameter, value)
	_ = a.msgs.SendAdminAnswerCallback(s.CallbackQuery, "type_the_text")
	markUp := msgs.NewMarkUp(
		msgs.NewRow(msgs.NewAdminButton("back_to_make_money_setting")),
		msgs.NewRow(msgs.NewAdminButton("exit")),
	).Build(a.bot.AdminLibrary[lang])

	return a.msgs.NewParseMarkUpMessage(s.User.ID, markUp, text)
}

func (a *Admin) MinerSettingCommand(s *model.Situation) error {
	return a.sendMinerSettingMenu(s)
}

func (a *Admin) sendMinerSettingMenu(s *model.Situation) error {
	lang := model.AdminLang(s.User.ID)
	text := a.adminFormatText(lang, "miner_setting_text")
	markUp := getMinerSettingMenu(s.BotLang, s.User.ID, a.bot.AdminLibrary[lang])

	msgID := db.RdbGetAdminMsgID(s.BotLang, s.User.ID)
	if msgID == 0 {
		id, err := a.msgs.NewIDParseMarkUpMessage(s.User.ID, markUp, text)
		if err != nil {
			return err
		}

		db.RdbSetAdminMsgID(s.BotLang, s.User.ID, id)
		return nil
	}

	return a.msgs.NewEditMarkUpMessage(s.User.ID, msgID, markUp, text)
}

func getMinerSettingMenu(botLang string, userID int64, texts map[string]string) *tgbotapi.InlineKeyboardMarkup {
	level := db.RdbGetMinerLevelSetting(botLang, userID)

	clickAmount := model.AdminSettings.GetClickAmount(botLang, level)
	upgradeCost := model.AdminSettings.GetUpgradeCost(botLang, level)

	markUp := msgs.NewIlMarkUp(
		msgs.NewIlRow(msgs.NewIlAdminButton("hash_per_click", "admin/not_clickable")),
		msgs.NewIlRow(
			msgs.NewIlCustomButton("-5", "admin/change_click_amount?dec&5"),
			msgs.NewIlCustomButton("-1", "admin/change_click_amount?dec&1"),
			msgs.NewIlCustomButton(strconv.Itoa(clickAmount), "admin/change_click_amount?set_hash"),
			msgs.NewIlCustomButton("+1", "admin/change_click_amount?inc&1"),
			msgs.NewIlCustomButton("+5", "admin/change_click_amount?inc&5")),

		msgs.NewIlRow(msgs.NewIlAdminButton("level_cost_button", "admin/not_clickable")),
		msgs.NewIlRow(
			msgs.NewIlCustomButton("-50", "admin/change_upgrade_amount?dec&50"),
			msgs.NewIlCustomButton("-10", "admin/change_upgrade_amount?dec&10"),
			msgs.NewIlCustomButton(strconv.Itoa(upgradeCost), "admin/change_upgrade_amount?set_price"),
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
	).Build(texts)

	return &markUp
}

func (a *Admin) ChangeClickAmountButton(s *model.Situation) error {
	level := db.RdbGetMinerLevelSetting(s.BotLang, s.User.ID)

	allParams := strings.Split(s.CallbackQuery.Data, "?")[1]
	changeParams := strings.Split(allParams, "&")
	operation := changeParams[0]

	switch operation {

	case "set_hash":
		db.RdbSetUser(s.BotLang, s.User.ID, "admin/set_count?hash")
		return a.msgs.NewParseMessage(s.User.ID, a.bot.AdminText(model.AdminLang(s.User.ID), "set_hash_value"))
	case "inc":
		value, _ := strconv.Atoi(changeParams[1])
		model.AdminSettings.GetParams(s.BotLang).ClickAmount[level] += value
	case "dec":
		value, _ := strconv.Atoi(changeParams[1])
		if model.AdminSettings.GetParams(s.BotLang).ClickAmount[level]-value < 1 {
			_ = a.msgs.SendAdminAnswerCallback(s.CallbackQuery, "already_min_value")
			return nil
		}
		model.AdminSettings.GetParams(s.BotLang).ClickAmount[level] -= value
	}

	model.SaveAdminSettings()
	return a.sendMinerSettingMenu(s)
}

func (a *Admin) ChangeUpgradeAmountButton(s *model.Situation) error {
	level := db.RdbGetMinerLevelSetting(s.BotLang, s.User.ID)

	allParams := strings.Split(s.CallbackQuery.Data, "?")[1]
	changeParams := strings.Split(allParams, "&")
	operation := changeParams[0]

	switch operation {
	case "set_price":
		db.RdbSetUser(s.BotLang, s.User.ID, "admin/set_count?price")
		return a.msgs.NewParseMessage(s.User.ID, a.bot.AdminText(model.AdminLang(s.User.ID), "set_price_value"))
	case "inc":
		value, _ := strconv.Atoi(changeParams[1])
		model.AdminSettings.GetParams(s.BotLang).UpgradeMinerCost[level] += value
	case "dec":
		value, _ := strconv.Atoi(changeParams[1])
		if model.AdminSettings.GetParams(s.BotLang).UpgradeMinerCost[level]-value < 1 {
			_ = a.msgs.SendAdminAnswerCallback(s.CallbackQuery, "already_min_value")
			return nil
		}
		model.AdminSettings.GetParams(s.BotLang).UpgradeMinerCost[level] -= value
	}

	model.SaveAdminSettings()
	return a.sendMinerSettingMenu(s)
}

func (a *Admin) ChangeMinerLvlButton(s *model.Situation) error {
	level := db.RdbGetMinerLevelSetting(s.BotLang, s.User.ID)

	operation := strings.Split(s.CallbackQuery.Data, "?")[1]
	switch operation {
	case "inc":
		if level == model.AdminSettings.GetMaxMinerLevel(s.BotLang)-1 {
			_ = a.msgs.SendAdminAnswerCallback(s.CallbackQuery, "already_max_level")
			return nil
		}
		level++
	case "dec":
		if level == 0 {
			_ = a.msgs.SendAdminAnswerCallback(s.CallbackQuery, "already_min_level")
			return nil
		}
		level--
	}

	db.RdbSetMinerLevelSetting(s.BotLang, s.User.ID, level)
	return a.sendMinerSettingMenu(s)
}

func (a *Admin) DeleteMinerLevelButton(s *model.Situation) error {
	level := db.RdbGetMinerLevelSetting(s.BotLang, s.User.ID)

	if model.AdminSettings.GetMaxMinerLevel(s.BotLang) == 1 {

		//TODO: нельзя удалить уровень
		fmt.Println("hello")

		return nil
	}

	model.AdminSettings.DeleteMinerLevel(s.BotLang, level)

	if level == model.AdminSettings.GetMaxMinerLevel(s.BotLang) {
		db.RdbSetMinerLevelSetting(s.BotLang, s.User.ID, level-1)
	}

	model.SaveAdminSettings()
	return a.sendMinerSettingMenu(s)
}

func (a *Admin) AddMinerLevelButton(s *model.Situation) error {
	level := db.RdbGetMinerLevelSetting(s.BotLang, s.User.ID)
	model.AdminSettings.AddMinerLevel(s.BotLang, level)

	db.RdbSetMinerLevelSetting(s.BotLang, s.User.ID, level+1)
	model.SaveAdminSettings()
	return a.sendMinerSettingMenu(s)
}

func (a *Admin) ExchangerSettingCommand(s *model.Situation) error {
	return a.sendExchangerSettingMenu(s)
}

func (a *Admin) sendExchangerSettingMenu(s *model.Situation) error {
	lang := model.AdminLang(s.User.ID)
	text := a.adminFormatText(lang, "exchanger_setting_text")
	markUp := getExchangerSettingMenu(s.BotLang, s.User.ID, a.bot.AdminLibrary[lang])

	msgID := db.RdbGetAdminMsgID(s.BotLang, s.User.ID)
	if msgID == 0 {
		id, err := a.msgs.NewIDParseMarkUpMessage(s.User.ID, markUp, text)
		if err != nil {
			return err
		}

		db.RdbSetAdminMsgID(s.BotLang, s.User.ID, id)
		return nil
	}

	return a.msgs.NewEditMarkUpMessage(s.User.ID, msgID, markUp, text)
}

func getExchangerSettingMenu(botLang string, userID int64, texts map[string]string) *tgbotapi.InlineKeyboardMarkup {
	hashToBTC := model.AdminSettings.GetParams(botLang).ExchangeHashToBTC
	btcToCurrency := int(model.AdminSettings.GetParams(botLang).ExchangeBTCToCurrency)

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
	).Build(texts)

	return &markUp
}

func (a *Admin) ChangeHashToBTCRateButton(s *model.Situation) error {
	allParams := strings.Split(s.CallbackQuery.Data, "?")[1]
	changeParams := strings.Split(allParams, "&")
	operation := changeParams[0]
	value, _ := strconv.Atoi(changeParams[1])

	switch operation {
	case "inc":
		model.AdminSettings.GetParams(s.BotLang).ExchangeHashToBTC += value
	case "dec":
		if model.AdminSettings.GetParams(s.BotLang).ExchangeHashToBTC-value < 1 {
			_ = a.msgs.SendAdminAnswerCallback(s.CallbackQuery, "already_min_value")
			return nil
		}
		model.AdminSettings.GetParams(s.BotLang).ExchangeHashToBTC -= value
	}

	model.SaveAdminSettings()
	return a.sendExchangerSettingMenu(s)
}

func (a *Admin) ChangeBTCToCurrencyRateButton(s *model.Situation) error {
	allParams := strings.Split(s.CallbackQuery.Data, "?")[1]
	changeParams := strings.Split(allParams, "&")
	operation := changeParams[0]
	value, _ := strconv.Atoi(changeParams[1])

	switch operation {
	case "inc":
		model.AdminSettings.GetParams(s.BotLang).ExchangeBTCToCurrency += float64(value)
	case "dec":
		if int(model.AdminSettings.GetParams(s.BotLang).ExchangeBTCToCurrency)-value < 1 {
			_ = a.msgs.SendAdminAnswerCallback(s.CallbackQuery, "already_min_value")
			return nil
		}
		model.AdminSettings.GetParams(s.BotLang).ExchangeBTCToCurrency -= float64(value)
	}

	model.SaveAdminSettings()
	return a.sendExchangerSettingMenu(s)
}

func (a *Admin) NotClickableButton(s *model.Situation) error {
	_ = a.msgs.SendAdminAnswerCallback(s.CallbackQuery, "not_clickable_button")
	return nil
}

//func (a *Admin) TopRewardSetting(s *model.Situation) {
//	for i := 0; i < 3; i++ {
//		value := 60000
//		model.AdminSettings.UpdateTopRewardSetting(a.bot.BotLang, i, value)
//		value /= 2
//	}
//}

func (a *Admin) SetTopAmountCommand(s *model.Situation) error {
	lang := model.AdminLang(s.User.ID)
	text := a.adminFormatText(lang, "change_top_settings_button")

	top := db.RdbGetTopLevelSetting(s.BotLang, s.User.ID)
	markUp := getTopSettingMenu(a.bot.AdminLibrary[lang], top+1, model.AdminSettings.GlobalParameters[s.BotLang].Parameters.TopReward[top])

	msgID := db.RdbGetAdminMsgID(s.BotLang, s.User.ID)
	if msgID == 0 {
		id, err := a.msgs.NewIDParseMarkUpMessage(s.User.ID, markUp, text)
		if err != nil {
			return err
		}

		db.RdbSetAdminMsgID(s.BotLang, s.User.ID, id)
		return nil
	}

	return a.msgs.NewEditMarkUpMessage(s.User.ID, msgID, markUp, text)
}

func getTopSettingMenu(texts map[string]string, top int, amount int) *tgbotapi.InlineKeyboardMarkup {
	markUp := msgs.NewIlMarkUp(
		msgs.NewIlRow(msgs.NewIlAdminButton("top_level_button", "admin/not_clickable")),
		msgs.NewIlRow(
			msgs.NewIlCustomButton("<<", "admin/change_top_level?dec"),
			msgs.NewIlCustomButton(strconv.Itoa(top), "admin/not_clickable"),
			msgs.NewIlCustomButton(">>", "admin/change_top_level?inc")),

		msgs.NewIlRow(msgs.NewIlAdminButton("top_amount_button", "admin/not_clickable")),
		msgs.NewIlRow(
			msgs.NewIlCustomButton("-5", "admin/change_top_amount?dec&5"),
			msgs.NewIlCustomButton("-1", "admin/change_top_amount?dec&1"),
			msgs.NewIlCustomButton(strconv.Itoa(amount), "admin/not_clickable"),
			msgs.NewIlCustomButton("+1", "admin/change_top_amount?inc&1"),
			msgs.NewIlCustomButton("+5", "admin/change_top_amount?inc&5")),

		msgs.NewIlRow(msgs.NewIlAdminButton("back_to_make_money_setting", "admin/make_money_setting")),
	).Build(texts)

	return &markUp
}

func (a *Admin) ChangeTopLevelCommand(s *model.Situation) error {
	level := db.RdbGetTopLevelSetting(s.BotLang, s.User.ID)
	operation := strings.Split(s.CallbackQuery.Data, "?")[1]

	switch operation {
	case "inc":
		if level == 2 {
			_ = a.msgs.SendAdminAnswerCallback(s.CallbackQuery, "already_max_level")
			return nil
		}
		level++
	case "dec":
		if level == 0 {
			_ = a.msgs.SendAdminAnswerCallback(s.CallbackQuery, "already_min_level")
			return nil
		}
		level--
	}

	db.RdbSetTopLevelSetting(s.BotLang, s.User.ID, level)
	return a.SetTopAmountCommand(s)
}

func (a *Admin) ChangeTopAmountButtonCommand(s *model.Situation) error {
	level := db.RdbGetTopLevelSetting(s.BotLang, s.User.ID)

	allParams := strings.Split(s.CallbackQuery.Data, "?")[1]
	changeParams := strings.Split(allParams, "&")
	operation := changeParams[0]

	switch operation {
	case "inc":
		value, _ := strconv.Atoi(changeParams[1])
		model.AdminSettings.GetParams(s.BotLang).TopReward[level] += value
	case "dec":
		value, _ := strconv.Atoi(changeParams[1])

		if model.AdminSettings.GetParams(s.BotLang).TopReward[level]-value < 1 {
			_ = a.msgs.SendAdminAnswerCallback(s.CallbackQuery, "already_min_value")
			return nil
		}

		model.AdminSettings.GetParams(s.BotLang).TopReward[level] -= value
	}

	model.SaveAdminSettings()
	return a.SetTopAmountCommand(s)
}
