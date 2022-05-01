package auth

import (
	"database/sql"
	"strconv"
	"strings"
	"time"

	"github.com/Stepan1328/miner-bot/model"
	"github.com/bots-empire/base-bot/msgs"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/pkg/errors"
)

const (
	oneSatoshi = 0.00000001
)

func (a *Auth) MakeClick(s *model.Situation) (error, bool) {
	if time.Now().Unix()/86400 > s.User.LastClick/86400 {
		if err := resetTodayMiningCounter(s, a.bot.DataBase); err != nil {
			return err, false
		}
	}

	if s.User.MiningToday >= model.AdminSettings.GetParams(s.BotLang).MaxOfClickPerDay {
		return a.reachedMaxAmountPerDay(s), true
	}

	return a.increaseBalanceAfterClick(s), false
}

func resetTodayMiningCounter(s *model.Situation, dataBase *sql.DB) error {
	s.User.MiningToday = 0
	s.User.LastClick = time.Now().Unix()

	rows, err := dataBase.Query(`
UPDATE users SET
      mining_today = 0, 
	last_click = ? 
WHERE id = ?;`,
		s.User.LastClick,
		s.User.ID)
	if err != nil {
		return errors.Wrap(err, "query failed")
	}
	rows.Close()

	return nil
}

func (a *Auth) reachedMaxAmountPerDay(s *model.Situation) error {
	text := a.bot.LangText(s.User.Language, "reached_max_amount_per_day",
		model.AdminSettings.GetParams(s.BotLang).MaxOfClickPerDay,
		model.AdminSettings.GetParams(s.BotLang).MaxOfClickPerDay)

	markUp := msgs.NewIlMarkUp(
		msgs.NewIlRow(msgs.NewIlURLButton("advertisement_button_text", model.AdminSettings.GetAdvertUrl(s.BotLang, 1))),
	).Build(a.bot.Language[s.User.Language])

	return a.msgs.NewParseMarkUpMessage(s.User.ID, &markUp, text)
}

func (a *Auth) increaseBalanceAfterClick(s *model.Situation) error {
	s.User.BalanceHash += getClickAmount(s.BotLang, s.User.MinerLevel)
	s.User.MiningToday++
	s.User.LastClick = time.Now().Unix()

	dataBase := a.bot.GetDataBase()
	_, err := dataBase.Exec(`
UPDATE users 
	SET balance_hash = balance_hash + ?, 
	    mining_today = mining_today + 1,
	    last_click = ?
WHERE id = ?;`,
		getClickAmount(s.BotLang, s.User.MinerLevel),
		s.User.LastClick,
		s.User.ID)
	if err != nil {
		text := "Failed increase balance after click: " + err.Error()
		a.msgs.SendNotificationToDeveloper(text, false)
		return err
	}

	return nil
}

func getClickAmount(botLang string, minerLevel int8) int {
	return model.AdminSettings.GetClickAmount(botLang, int(minerLevel-1))
}

func (a *Auth) ChangeHashToBTC(s *model.Situation) (error, float64) {
	count, err := extractAmountFromMsg(s.Message.Text)
	if err != nil {
		return nil, 0
	}

	if count <= 0 || count > s.User.BalanceHash {
		return nil, 0
	}

	amountBTC := count / model.AdminSettings.GetParams(s.BotLang).ExchangeHashToBTC
	clearAmount := amountBTC * model.AdminSettings.GetParams(s.BotLang).ExchangeHashToBTC
	amountToChange := oneSatoshi * float64(amountBTC)

	_, err = a.bot.GetDataBase().Exec(`
UPDATE users 
	SET balance_hash = balance_hash - ?, 
	    balance_btc = balance_btc + ?
WHERE id = ?;`,
		clearAmount,
		amountToChange,
		s.User.ID)
	if err != nil {
		text := "Failed exchange hash to btc: " + err.Error()
		a.msgs.SendNotificationToDeveloper(text, false)
		return err, 0
	}

	s.User.BalanceBTC += amountToChange
	return nil, amountToChange
}

func extractAmountFromMsg(text string) (int, error) {
	return strconv.Atoi(text)
}

func (a *Auth) ChangeBTCToCurrency(s *model.Situation) (error, int) {
	count, err := extractAmountFromMsg(s.Message.Text)
	if err != nil {
		return nil, 0
	}

	amountBTC := float64(count) * model.AdminSettings.GetParams(s.BotLang).ExchangeBTCToCurrency * oneSatoshi
	if count <= 0 || amountBTC > s.User.BalanceBTC {
		return nil, 0
	}

	dataBase := a.bot.GetDataBase()
	_, err = dataBase.Exec(`
UPDATE users 
	SET balance_btc = balance_btc - ?, 
	    balance = balance + ?
WHERE id = ?;`,
		amountBTC,
		count,
		s.User.ID)
	if err != nil {
		text := "Failed exchange btc to currency: " + err.Error()
		a.msgs.SendNotificationToDeveloper(text, false)
		return err, 0
	}

	s.User.Balance += count
	return nil, count
}

func (a *Auth) UpgradeMinerLevel(s *model.Situation) (bool, error) {
	var err error
	s.User, err = a.GetUser(s.User.ID)
	if err != nil {
		return false, err
	}

	if int8(len(model.AdminSettings.GetParams(s.BotLang).UpgradeMinerCost)) == s.User.MinerLevel {
		return false, model.ErrMaxLevelAlreadyCompleted
	}

	if s.User.BalanceHash < model.AdminSettings.GetParams(s.BotLang).UpgradeMinerCost[s.User.MinerLevel] {
		return true, nil
	}

	dataBase := a.bot.GetDataBase()
	_, err = dataBase.Exec(`
UPDATE users 
	SET balance_hash = balance_hash - ?, 
	    miner_level = miner_level + 1
WHERE id = ?;`,
		model.AdminSettings.GetParams(s.BotLang).UpgradeMinerCost[s.User.MinerLevel],
		s.User.ID)
	if err != nil {
		text := "Failed update miner level: " + err.Error()
		a.msgs.SendNotificationToDeveloper(text, false)
		return false, err
	}
	s.User.MinerLevel++

	return false, nil
}

func (a *Auth) WithdrawMoneyFromBalance(s *model.Situation, amount string) error {
	amount = strings.Replace(amount, " ", "", -1)
	amountInt, err := strconv.Atoi(amount)
	if err != nil {
		msg := tgbotapi.NewMessage(s.User.ID, a.bot.LangText(s.User.Language, "incorrect_amount"))
		return a.msgs.SendMsgToUser(msg)
	}

	if amountInt < model.AdminSettings.GetParams(s.BotLang).MinWithdrawalAmount {
		return a.minAmountNotReached(s.User, s.BotLang)
	}

	if s.User.Balance < amountInt {
		msg := tgbotapi.NewMessage(s.User.ID, a.bot.LangText(s.User.Language, "lack_of_funds"))
		return a.msgs.SendMsgToUser(msg)
	}

	if s.User.MinerLevel < 3 {
		msg := tgbotapi.NewMessage(s.User.ID, a.bot.LangText(s.User.Language, "insufficient_miner_level"))
		return a.msgs.SendMsgToUser(msg)
	}

	return a.sendInvitationToSubs(s, amount)
}

func (a *Auth) minAmountNotReached(u *model.User, botLang string) error {
	text := a.bot.LangText(u.Language, "minimum_amount_not_reached",
		model.AdminSettings.GetParams(botLang).MinWithdrawalAmount)

	return a.msgs.NewParseMessage(u.ID, text)
}

func (a *Auth) sendInvitationToSubs(s *model.Situation, amount string) error {
	text := a.bot.LangText(s.User.Language, "withdrawal_not_subs_text")

	msg := tgbotapi.NewMessage(s.User.ID, text)
	msg.ReplyMarkup = msgs.NewIlMarkUp(
		msgs.NewIlRow(msgs.NewIlURLButton("advertising_button", model.AdminSettings.GetAdvertUrl(s.BotLang, 1))),
		msgs.NewIlRow(msgs.NewIlDataButton("im_subscribe_button", "/withdrawal_money?"+amount)),
	).Build(a.bot.Language[s.User.Language])

	return a.msgs.SendMsgToUser(msg)
}

func (a *Auth) CheckSubscribeToWithdrawal(s *model.Situation, amount int) bool {
	if s.User.Balance < amount {
		return false
	}

	if !a.CheckSubscribe(s, "withdrawal") {
		_ = a.sendInvitationToSubs(s, strconv.Itoa(amount))
		return false
	}

	s.User.Balance -= amount
	dataBase := a.bot.GetDataBase()
	rows, err := dataBase.Query(`
UPDATE users 
	SET balance = ?
WHERE id = ?;`,
		s.User.Balance,
		s.User.ID)
	if err != nil {
		return false
	}
	_ = rows.Close()

	msg := tgbotapi.NewMessage(s.User.ID, a.bot.LangText(s.User.Language, "successfully_withdrawn"))
	_ = a.msgs.SendMsgToUser(msg)
	return true
}

func (a *Auth) GetABonus(s *model.Situation) error {
	if !a.CheckSubscribe(s, "get_bonus") {
		text := a.bot.LangText(s.User.Language, "user_dont_subscribe")
		return a.msgs.SendSimpleMsg(s.User.ID, text)
	}

	if s.User.TakeBonus {
		text := a.bot.LangText(s.User.Language, "bonus_already_have")
		return a.msgs.SendSimpleMsg(s.User.ID, text)
	}

	s.User.Balance += model.AdminSettings.GetParams(s.BotLang).BonusAmount
	dataBase := a.bot.GetDataBase()
	rows, err := dataBase.Query(`
UPDATE users 
	SET balance = ?, 
	    take_bonus = ? 
WHERE id = ?;`,
		s.User.Balance,
		true,
		s.User.ID)
	if err != nil {
		return err
	}
	_ = rows.Close()

	text := a.bot.LangText(s.User.Language, "bonus_have_received")
	return a.msgs.SendSimpleMsg(s.User.ID, text)
}

func (a *Auth) CheckSubscribe(s *model.Situation, source string) bool {
	model.CheckSubscribe.WithLabelValues(
		a.bot.BotLink,
		s.BotLang,
		model.AdminSettings.GetAdvertUrl(s.BotLang, model.MainAdvert),
		source,
	).Inc()

	member, err := model.Bots[s.BotLang].Bot.GetChatMember(tgbotapi.GetChatMemberConfig{
		ChatConfigWithUser: tgbotapi.ChatConfigWithUser{
			ChatID: model.AdminSettings.GetAdvertChannelID(s.BotLang, model.MainAdvert),
			UserID: s.User.ID,
		},
	})

	if err == nil {
		if err := a.addMemberToSubsBase(s); err != nil {
			return false
		}
		return checkMemberStatus(member)
	}
	return false
}

func checkMemberStatus(member tgbotapi.ChatMember) bool {
	if member.IsAdministrator() {
		return true
	}
	if member.IsCreator() {
		return true
	}
	if member.Status == "member" {
		return true
	}
	return false
}

func (a *Auth) addMemberToSubsBase(s *model.Situation) error {
	dataBase := a.bot.GetDataBase()
	rows, err := dataBase.Query(`
SELECT * FROM subs 
	WHERE id = ?;`,
		s.User.ID)
	if err != nil {
		return err
	}

	user, err := readUser(rows)
	if err != nil {
		return err
	}

	if user.ID != 0 {
		return nil
	}
	rows, err = dataBase.Query(`
INSERT INTO subs VALUES(?);`,
		s.User.ID)
	if err != nil {
		return err
	}
	_ = rows.Close()
	return nil
}

func readUser(rows *sql.Rows) (*model.User, error) {
	defer rows.Close()

	var users []*model.User

	for rows.Next() {
		var id int64

		if err := rows.Scan(&id); err != nil {
			return nil, model.ErrScanSqlRow
		}

		users = append(users, &model.User{
			ID: id,
		})
	}
	if len(users) == 0 {
		users = append(users, &model.User{
			ID: 0,
		})
	}
	return users[0], nil
}
