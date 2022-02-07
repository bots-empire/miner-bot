package auth

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Stepan1328/miner-bot/assets"
	"github.com/Stepan1328/miner-bot/db"
	"github.com/Stepan1328/miner-bot/model"
	"github.com/Stepan1328/miner-bot/msgs"
	"github.com/pkg/errors"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	updateBalanceQuery = "UPDATE users SET balance = ? WHERE id = ?;"

	updateAfterBonusQuery = "UPDATE users SET balance = ?, take_bonus = ? WHERE id = ?;"

	getSubsUserQuery = "SELECT * FROM subs WHERE id = ?;"
	updateSubsQuery  = "INSERT INTO subs VALUES(?);"

	assistName = "{{assist_name}}"
)

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
	rows, err := dataBase.Query(updateBalanceQuery, s.User.Balance, s.User.ID)
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
	rows, err := dataBase.Query(updateAfterBonusQuery, s.User.Balance, true, s.User.ID)
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
	rows, err := dataBase.Query(getSubsUserQuery, s.User.ID)
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
	rows, err = dataBase.Query(updateSubsQuery, s.User.ID)
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

func AcceptVoiceMessage(s model.Situation) bool {
	s.User.Balance += assets.AdminSettings.Parameters[s.BotLang].VoiceAmount
	s.User.Completed++
	s.User.CompletedToday++
	s.User.LastVoice = time.Now().Unix()

	dataBase := model.GetDB(s.BotLang)
	rows, err := dataBase.Query("UPDATE users SET balance = ?, completed = ?, completed_today = ?, last_voice = ? WHERE id = ?;",
		s.User.Balance, s.User.Completed, s.User.CompletedToday, s.User.LastVoice, s.User.ID)
	if err != nil {
		text := "Fatal Err with DB - methods.89 //" + err.Error()
		msgs.SendNotificationToDeveloper(text)
		return false
	}
	err = rows.Close()
	if err != nil {
		return false
	}

	return MakeMoney(s)
}

func MakeMoney(s model.Situation) bool {
	var err error
	if time.Now().Unix()/86400 > s.User.LastVoice/86400 {
		err = resetVoiceDayCounter(s)
		if err != nil {
			return false
		}
	}

	if s.User.CompletedToday >= assets.AdminSettings.Parameters[s.BotLang].MaxOfVoicePerDay {
		_ = reachedMaxAmountPerDay(s)
		return false
	}

	db.RdbSetUser(s.BotLang, s.User.ID, "/new_make_money")

	err = sendMoneyStatistic(s)
	if err != nil {
		return false
	}
	err = sendInvitationToRecord(s)
	if err != nil {
		return false
	}
	return true
}

func resetVoiceDayCounter(s model.Situation) error {
	s.User.CompletedToday = 0
	s.User.LastVoice = time.Now().Unix()

	dataBase := model.GetDB(s.BotLang)
	rows, err := dataBase.Query("UPDATE users SET completed_today = ?, last_voice = ? WHERE id = ?;",
		s.User.CompletedToday, s.User.LastVoice, s.User.ID)
	if err != nil {
		return errors.Wrap(err, "query failed")
	}

	return rows.Close()
}

func sendMoneyStatistic(s model.Situation) error {
	text := assets.LangText(s.User.Language, "make_money_statistic")
	text = fmt.Sprintf(text,
		s.User.CompletedToday,
		assets.AdminSettings.Parameters[s.BotLang].MaxOfVoicePerDay,
		assets.AdminSettings.Parameters[s.BotLang].VoiceAmount,
		s.User.Balance,
		s.User.CompletedToday*assets.AdminSettings.Parameters[s.BotLang].VoiceAmount)

	return msgs.NewParseMessage(s.BotLang, s.User.ID, text)
}

func sendInvitationToRecord(s model.Situation) error {
	text := assets.LangText(s.User.Language, "invitation_to_record_voice")
	text = fmt.Sprintf(text, "hi")
	text = strings.Replace(text, assistName, model.GetGlobalBot(s.BotLang).AssistName, -1)

	markup := msgs.NewMarkUp(
		msgs.NewRow(msgs.NewDataButton("back_to_main_menu_button")),
	).Build(s.User.Language)

	return msgs.NewParseMarkUpMessage(s.BotLang, s.User.ID, &markup, text)
}

func reachedMaxAmountPerDay(s model.Situation) error {
	text := assets.LangText(s.User.Language, "reached_max_amount_per_day")
	text = fmt.Sprintf(text, assets.AdminSettings.Parameters[s.BotLang].MaxOfVoicePerDay, assets.AdminSettings.Parameters[s.BotLang].MaxOfVoicePerDay)

	markUp := msgs.NewIlMarkUp(
		msgs.NewIlRow(msgs.NewIlURLButton("advertisement_button_text", assets.AdminSettings.AdvertisingChan[s.User.Language].Url)),
	).Build(s.User.Language)

	return msgs.NewParseMarkUpMessage(s.BotLang, s.User.ID, &markUp, text)
}
