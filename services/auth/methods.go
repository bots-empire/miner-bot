package auth

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Stepan1328/miner-bot/assets"
	"github.com/Stepan1328/miner-bot/model"
	"github.com/Stepan1328/miner-bot/msgs"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/pkg/errors"
)

func MakeClick(s model.Situation) (error, bool) {
	if time.Now().Unix()/86400 > s.User.LastClick/86400 {
		if err := resetTodayMiningCounter(s); err != nil {
			return err, false
		}
	}

	if s.User.MiningToday >= assets.AdminSettings.Parameters[s.BotLang].MaxOfClickPerDay {
		return reachedMaxAmountPerDay(s), true
	}

	return increaseBalanceAfterClick(s), false
}

func resetTodayMiningCounter(s model.Situation) error {
	s.User.MiningToday = 0
	s.User.LastClick = time.Now().Unix()

	dataBase := model.GetDB(s.BotLang)
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

func reachedMaxAmountPerDay(s model.Situation) error {
	text := assets.LangText(s.User.Language, "reached_max_amount_per_day")
	text = fmt.Sprintf(text,
		assets.AdminSettings.Parameters[s.BotLang].MaxOfClickPerDay,
		assets.AdminSettings.Parameters[s.BotLang].MaxOfClickPerDay)

	markUp := msgs.NewIlMarkUp(
		msgs.NewIlRow(msgs.NewIlURLButton("advertisement_button_text", assets.AdminSettings.AdvertisingChan[s.User.Language].Url)),
	).Build(s.User.Language)

	return msgs.NewParseMarkUpMessage(s.BotLang, s.User.ID, &markUp, text)
}

func increaseBalanceAfterClick(s model.Situation) error {
	s.User.BalanceHash += getClickAmount(s.BotLang, s.User.MinerLevel)
	s.User.MiningToday++
	s.User.LastClick = time.Now().Unix()

	dataBase := model.GetDB(s.BotLang)
	rows, err := dataBase.Query(`
UPDATE users 
	SET balance_hash = balance_hash + ?, 
	    mining_today = mining_today + 1,
	    last_click = ?
WHERE id = ?;`,
		getClickAmount(s.BotLang, s.User.MinerLevel),
		s.User.LastClick,
		s.User.ID)
	if err != nil {
		text := "Fatal Err with DB - methods.89 //" + err.Error()
		msgs.SendNotificationToDeveloper(text)
		return err
	}
	rows.Close()

	return nil
}

func getClickAmount(botLang string, minerLevel int8) int {
	return assets.AdminSettings.Parameters[botLang].ClickAmount[minerLevel-1]
}

func WithdrawMoneyFromBalance(s model.Situation, amount string) error {
	amount = strings.Replace(amount, " ", "", -1)
	amountInt, err := strconv.Atoi(amount)
	if err != nil {
		msg := tgbotapi.NewMessage(s.User.ID, assets.LangText(s.User.Language, "incorrect_amount"))
		return msgs.SendMsgToUser(s.BotLang, msg)
	}

	if amountInt < assets.AdminSettings.Parameters[s.BotLang].MinWithdrawalAmount {
		return minAmountNotReached(s.User, s.BotLang)
	}

	if s.User.Balance < amountInt {
		msg := tgbotapi.NewMessage(s.User.ID, assets.LangText(s.User.Language, "lack_of_funds"))
		return msgs.SendMsgToUser(s.BotLang, msg)
	}

	return sendInvitationToSubs(s, amount)
}

func minAmountNotReached(u *model.User, botLang string) error {
	text := assets.LangText(u.Language, "minimum_amount_not_reached")
	text = fmt.Sprintf(text, assets.AdminSettings.Parameters[botLang].MinWithdrawalAmount)

	return msgs.NewParseMessage(botLang, u.ID, text)
}

func sendInvitationToSubs(s model.Situation, amount string) error {
	text := msgs.GetFormatText(s.User.Language, "withdrawal_not_subs_text")

	msg := tgbotapi.NewMessage(s.User.ID, text)
	msg.ReplyMarkup = msgs.NewIlMarkUp(
		msgs.NewIlRow(msgs.NewIlURLButton("advertising_button", assets.AdminSettings.AdvertisingChan[s.User.Language].Url)),
		msgs.NewIlRow(msgs.NewIlDataButton("im_subscribe_button", "/withdrawal_money?"+amount)),
	).Build(s.User.Language)

	return msgs.SendMsgToUser(s.BotLang, msg)
}

func CheckSubscribeToWithdrawal(s model.Situation, amount int) bool {
	if s.User.Balance < amount {
		return false
	}

	if !CheckSubscribe(s, "withdrawal") {
		_ = sendInvitationToSubs(s, strconv.Itoa(amount))
		return false
	}

	s.User.Balance -= amount
	dataBase := model.GetDB(s.BotLang)
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

	msg := tgbotapi.NewMessage(s.User.ID, assets.LangText(s.User.Language, "successfully_withdrawn"))
	_ = msgs.SendMsgToUser(s.BotLang, msg)
	return true
}

func GetABonus(s model.Situation) error {
	if !CheckSubscribe(s, "get_bonus") {
		text := assets.LangText(s.User.Language, "user_dont_subscribe")
		return msgs.SendSimpleMsg(s.BotLang, s.User.ID, text)
	}

	if s.User.TakeBonus {
		text := assets.LangText(s.User.Language, "bonus_already_have")
		return msgs.SendSimpleMsg(s.BotLang, s.User.ID, text)
	}

	s.User.Balance += assets.AdminSettings.Parameters[s.BotLang].BonusAmount
	dataBase := model.GetDB(s.BotLang)
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

	text := assets.LangText(s.User.Language, "bonus_have_received")
	return msgs.SendSimpleMsg(s.BotLang, s.User.ID, text)
}

func CheckSubscribe(s model.Situation, source string) bool {
	model.CheckSubscribe.WithLabelValues(
		model.GetGlobalBot(s.BotLang).BotLink,
		s.BotLang,
		assets.AdminSettings.AdvertisingChan[s.BotLang].Url,
		source,
	).Inc()

	member, err := model.Bots[s.BotLang].Bot.GetChatMember(tgbotapi.GetChatMemberConfig{
		ChatConfigWithUser: tgbotapi.ChatConfigWithUser{
			ChatID: assets.AdminSettings.AdvertisingChan[s.BotLang].ChannelID,
			UserID: s.User.ID,
		},
	})

	if err == nil {
		if err := addMemberToSubsBase(s); err != nil {
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

func addMemberToSubsBase(s model.Situation) error {
	dataBase := model.GetDB(s.BotLang)
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
