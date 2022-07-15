package administrator

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Stepan1328/miner-bot/db"
	"github.com/Stepan1328/miner-bot/model"
	"github.com/bots-empire/base-bot/msgs"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	changeLeftBorder  = "change_left_border"
	changeRightBorder = "change_right_border"
	changeAmount      = "change_amount"

	maxLevel = 5
)

func (a *Admin) sendRewardSettings(s *model.Situation, reward *model.RewardsGap, resend bool) error {
	db.RdbSetUser(s.BotLang, s.User.ID, "admin")

	markUp, text := a.rewardsMarkUpAndText(s.User.ID, reward)

	if resend {
		db.DeleteOldAdminMsg(s.BotLang, s.User.ID)

		msgId, err := a.msgs.NewIDParseMarkUpMessage(s.User.ID, markUp, text)
		if err != nil {
			return err
		}

		db.RdbSetAdminMsgID(s.BotLang, s.User.ID, msgId)
	}

	msgID := db.RdbGetAdminMsgID(s.BotLang, s.User.ID)
	err := a.msgs.NewEditMarkUpMessage(s.User.ID, msgID, markUp, text)
	if err != nil {
		return err
	}

	return nil
}

func (a *Admin) rewardsMarkUpAndText(userID int64, gap *model.RewardsGap) (*tgbotapi.InlineKeyboardMarkup, string) {
	lang := model.AdminLang(userID)

	markUp := msgs.NewIlMarkUp(
		msgs.NewIlRow(
			msgs.NewIlCustomButton(strconv.Itoa(gap.LeftBorder), "admin/change_rewards_gap?"+changeLeftBorder),
			msgs.NewIlCustomButton(strconv.Itoa(gap.RightBorder), "admin/change_rewards_gap?"+changeRightBorder),
			msgs.NewIlCustomButton(strconv.Itoa(gap.Amount), "admin/change_rewards_gap?"+changeAmount),
		),
		msgs.NewIlRow(
			msgs.NewIlCustomButton("⬅️", "admin/change_gap?-1"),
			msgs.NewIlCustomButton("✅", "admin/apply_rewards"),
			msgs.NewIlCustomButton("➡️", "admin/change_gap?1"),
		),
		msgs.NewIlRow(
			msgs.NewIlCustomButton("⬅️ Lvl", "admin/change_level?-1"),
			msgs.NewIlCustomButton(strconv.Itoa(gap.Level), "admin/lvl_info"),
			msgs.NewIlCustomButton("Lvl ➡️", "admin/change_level?1"),
		),
		msgs.NewIlRow(
			msgs.NewIlAdminButton("delete_gap", "admin/delete_gap"),
			msgs.NewIlAdminButton("delete_level", "admin/delete_level"),
		),
		msgs.NewIlRow(msgs.NewIlAdminButton("check_ranges", "admin/check_rewards_ranges")),
		msgs.NewIlRow(msgs.NewIlAdminButton("back_to_make_money_setting", "admin/make_money_setting")),
	).Build(a.bot.AdminLibrary[lang])

	return &markUp, a.bot.AdminText(lang, "change_referral_rewards")
}

func (a *Admin) ChangeRewardsGapCommand(s *model.Situation) error {
	command := strings.Split(s.CallbackQuery.Data, "?")[1]

	reward, err := RdbGetRewardGap(s.BotLang, s.User.ID)
	if err != nil {
		return err
	}

	db.RdbSetUser(s.BotLang, s.User.ID, "admin/change_rewards_gap?"+command)

	var value int
	switch command {
	case changeLeftBorder:
		value = reward.LeftBorder
	case changeRightBorder:
		value = reward.RightBorder
	case changeAmount:
		value = reward.Amount
	}

	lang := model.AdminLang(s.User.ID)
	text := a.adminFormatText(lang, "set_new_gap_info", value)

	_ = a.msgs.SendAdminAnswerCallback(s.CallbackQuery, "type_the_text") //TODO: make back button
	//markUp := msgs.NewMarkUp(
	//	msgs.NewRow(msgs.NewAdminButton("back_to_reward_setting_setting")),
	//	msgs.NewRow(msgs.NewAdminButton("admin_log_out_text")),
	//).Build(a.bot.AdminLibrary[lang])
	//
	//return a.msgs.NewParseMarkUpMessage(s.User.ID, markUp, text)
	return a.msgs.NewParseMessage(s.User.ID, text)
}

func (a *Admin) UpdateRewardsGapCommand(s *model.Situation) error {
	command := strings.Split(s.Params.Level, "?")[1]

	reward, err := RdbGetRewardGap(s.BotLang, s.User.ID)
	if err != nil {
		return err
	}

	newValue, err := strconv.Atoi(s.Message.Text)
	if err != nil {
		return a.sendErrorInChangeParameter(s.User.ID, "incorrect_value")
	}

	switch command {
	case changeLeftBorder:
		if newValue > reward.RightBorder {
			return a.sendErrorInChangeParameter(s.User.ID, "value_higher")
		}
		if newValue < 1 {
			return a.sendErrorInChangeParameter(s.User.ID, "value_lower_one")
		}

		reward.LeftBorder = newValue
	case changeRightBorder:
		if newValue < reward.LeftBorder {
			return a.sendErrorInChangeParameter(s.User.ID, "value_lower")
		}

		reward.RightBorder = newValue
	case changeAmount:
		if newValue < 0 {
			return a.sendErrorInChangeParameter(s.User.ID, "value_lower_zero")
		}

		reward.Amount = newValue
	}

	if err = RdbSetRewardGap(s.BotLang, s.User.ID, reward); err != nil {
		return err
	}

	return a.sendRewardSettings(s, reward, true)
}

func (a *Admin) ApplyRewardCommand(s *model.Situation) error {
	reward, err := RdbGetRewardGap(s.BotLang, s.User.ID)
	if err != nil {
		return err
	}

	leftBorder := reward.LeftBorder
	model.AdminSettings.GetParams(s.BotLang).ReferralReward.UpdateGap(reward)

	reward = model.AdminSettings.GetParams(s.BotLang).ReferralReward.GetGapByCount(reward.Level, leftBorder)
	if err = RdbSetRewardGap(s.BotLang, s.User.ID, reward); err != nil {
		return err
	}

	_ = a.msgs.SendAdminAnswerCallback(s.CallbackQuery, "confirmed")
	return nil
}

func (a *Admin) ChangeGapCommand(s *model.Situation) error {
	reward, err := RdbGetRewardGap(s.BotLang, s.User.ID)
	if err != nil {
		return err
	}

	reward = model.AdminSettings.GetParams(s.BotLang).ReferralReward.GetGapByIndex(reward.Level, reward.Index)

	direction, _ := strconv.Atoi(strings.Split(s.CallbackQuery.Data, "?")[1])

	if direction == -1 && reward.Index == 1 {
		_ = a.msgs.SendAdminAnswerCallback(s.CallbackQuery, "already_first_gap")
		return nil
	}

	if direction == 1 && reward.Index == model.AdminSettings.GetParams(s.BotLang).ReferralReward.MaxIndexByLvl(reward.Level) {
		newGap := model.AdminSettings.GetParams(s.BotLang).ReferralReward.AddGap(reward.Level)
		err = RdbSetRewardGap(s.BotLang, s.User.ID, newGap)
		if err != nil {
			return err
		}

		_ = a.msgs.SendAdminAnswerCallback(s.CallbackQuery, "new_gap_added")

		return a.sendRewardSettings(s, newGap, false)
	}

	newGap := model.AdminSettings.GetParams(s.BotLang).ReferralReward.GetGapByIndex(reward.Level, reward.Index+direction)

	err = RdbSetRewardGap(s.BotLang, s.User.ID, newGap)
	if err != nil {
		return err
	}
	_ = a.msgs.SendAdminAnswerCallback(s.CallbackQuery, "gap_changed")
	return a.sendRewardSettings(s, newGap, false)
}

func (a *Admin) ChangeLevelCommand(s *model.Situation) error {
	reward, err := RdbGetRewardGap(s.BotLang, s.User.ID)
	if err != nil {
		return err
	}

	reward = model.AdminSettings.GetParams(s.BotLang).ReferralReward.GetGapByIndex(reward.Level, reward.Index)

	direction, _ := strconv.Atoi(strings.Split(s.CallbackQuery.Data, "?")[1])

	if direction == -1 && reward.Level == 1 {
		_ = a.msgs.SendAdminAnswerCallback(s.CallbackQuery, "already_first_lvl")
		return nil
	}

	if direction == 1 && reward.Level == model.AdminSettings.GetParams(s.BotLang).ReferralReward.MaxLevel() {
		if reward.Level == maxLevel {
			return a.sendErrorInChangeParameter(s.User.ID, "already_max_lvl_count")
		}

		newGap := model.AdminSettings.GetParams(s.BotLang).ReferralReward.AddLvl()
		err = RdbSetRewardGap(s.BotLang, s.User.ID, newGap)
		if err != nil {
			return err
		}

		_ = a.msgs.SendAdminAnswerCallback(s.CallbackQuery, "new_lvl_added")

		return a.sendRewardSettings(s, newGap, false)
	}

	newGap := model.AdminSettings.GetParams(s.BotLang).ReferralReward.GetGapByIndex(reward.Level+direction, 1)

	err = RdbSetRewardGap(s.BotLang, s.User.ID, newGap)
	if err != nil {
		return err
	}
	_ = a.msgs.SendAdminAnswerCallback(s.CallbackQuery, "level_changed")
	return a.sendRewardSettings(s, newGap, false)
}

func (a *Admin) LevelInfoCommand(s *model.Situation) error {
	_ = a.msgs.SendAdminAnswerCallback(s.CallbackQuery, "lvl_number_info")
	return nil
}

func (a *Admin) DeleteGapCommand(s *model.Situation) error {
	reward, err := RdbGetRewardGap(s.BotLang, s.User.ID)
	if err != nil {
		return err
	}

	reward = model.AdminSettings.GetParams(s.BotLang).ReferralReward.GetGapByIndex(reward.Level, reward.Index)

	if model.AdminSettings.GetParams(s.BotLang).ReferralReward.LastGapInLvl(reward.Level) {
		_ = a.msgs.SendAdminAnswerCallback(s.CallbackQuery, "cant_delete_last_elem")
		return nil
	}

	newGap := model.AdminSettings.GetParams(s.BotLang).ReferralReward.DeleteGap(reward.Level, reward.Index)

	err = RdbSetRewardGap(s.BotLang, s.User.ID, newGap)
	if err != nil {
		return err
	}
	_ = a.msgs.SendAdminAnswerCallback(s.CallbackQuery, "gap_deleted")
	return a.sendRewardSettings(s, newGap, false)
}

func (a *Admin) DeleteLevelCommand(s *model.Situation) error {
	reward, err := RdbGetRewardGap(s.BotLang, s.User.ID)
	if err != nil {
		return err
	}

	reward = model.AdminSettings.GetParams(s.BotLang).ReferralReward.GetGapByIndex(reward.Level, reward.Index)

	if model.AdminSettings.GetParams(s.BotLang).ReferralReward.MaxLevel() == 1 {
		_ = a.msgs.SendAdminAnswerCallback(s.CallbackQuery, "cant_delete_last_lvl")
		return nil
	}

	newGap := model.AdminSettings.GetParams(s.BotLang).ReferralReward.DeleteLvl(reward.Level)

	err = RdbSetRewardGap(s.BotLang, s.User.ID, newGap)
	if err != nil {
		return err
	}
	_ = a.msgs.SendAdminAnswerCallback(s.CallbackQuery, "level_deleted")
	return a.sendRewardSettings(s, newGap, false)
}

func (a *Admin) ViewLevelCommand(s *model.Situation) error {
	reward, err := RdbGetRewardGap(s.BotLang, s.User.ID)
	if err != nil {
		return err
	}

	reward = model.AdminSettings.GetParams(s.BotLang).ReferralReward.GetGapByIndex(reward.Level, reward.Index)

	lvl := model.AdminSettings.GetParams(s.BotLang).ReferralReward.GetLvl(reward.Level)
	if err = a.msgs.NewParseMessage(s.User.ID, rewardsLvlToString(lvl)); err != nil {
		return err
	}

	_ = a.msgs.SendAdminAnswerCallback(s.CallbackQuery, "check_ranges")
	return a.sendRewardSettings(s, reward, true)
}

func rewardsLvlToString(lvl model.RewardsLvl) string {
	rawLvl := fmt.Sprintf("%d Lvl:\n", lvl[0].Level)

	for _, gap := range lvl {
		rawLvl += fmt.Sprintf("%d - %d:  %d\n", gap.LeftBorder, gap.RightBorder, gap.Amount)
	}

	return rawLvl
}

func (a *Admin) sendErrorInChangeParameter(userID int64, textKey string) error {
	lang := model.AdminLang(userID)
	text := a.adminFormatText(lang, textKey)

	return a.msgs.NewParseMessage(userID, text)
}
