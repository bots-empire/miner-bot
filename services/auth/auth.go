package auth

import (
	"database/sql"
	"math/rand"
	"strings"
	"time"

	"github.com/Stepan1328/miner-bot/model"
	"github.com/Stepan1328/miner-bot/services/administrator"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/pkg/errors"
)

func (a *Auth) CheckingTheUser(message *tgbotapi.Message) (*model.User, error) {
	dataBase := a.bot.GetDataBase()
	rows, err := dataBase.Query(`
SELECT * FROM users 
	WHERE id = ?;`,
		message.From.ID)
	if err != nil {
		return nil, errors.Wrap(err, "get user")
	}

	users, err := a.ReadUsers(rows)
	if err != nil {
		return nil, errors.Wrap(err, "read user")
	}

	switch len(users) {
	case 0:
		user := createSimpleUser(a.bot.LanguageInBot[0], message)
		if len(a.bot.LanguageInBot) > 1 && !administrator.ContainsInAdmin(message.From.ID) {
			user.Language = "not_defined" // TODO: refactor
		}
		referralID := a.pullReferralID(message)
		if err := a.addNewUser(user, a.bot.LanguageInBot[0], referralID); err != nil {
			return nil, errors.Wrap(err, "add new user")
		}

		model.TotalIncome.WithLabelValues(
			a.bot.BotLink,
			a.bot.BotLang,
		).Inc()

		if user.Language == "not_defined" {
			return user, model.ErrNotSelectedLanguage
		}
		return user, nil
	case 1:
		if users[0].Language == "not_defined" {
			return users[0], model.ErrNotSelectedLanguage
		}
		return users[0], nil
	default:
		return nil, model.ErrFoundTwoUsers
	}
}

func (a *Auth) SetStartLanguage(callback *tgbotapi.CallbackQuery) error {
	data := strings.Split(callback.Data, "?")[1]
	dataBase := a.bot.GetDataBase()
	_, err := dataBase.Exec("UPDATE users SET lang = ? WHERE id = ?", data, callback.From.ID)
	if err != nil {
		return err
	}
	return nil
}

func (a *Auth) addNewUser(user *model.User, botLang string, referralID int64) error {
	user.MinerLevel = 1
	user.RegisterTime = time.Now().Unix()
	user.Language = botLang

	dataBase := a.bot.GetDataBase()
	rows, err := dataBase.Query(`
INSERT INTO users
	VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);`,
		user.ID,
		user.Balance,
		user.BalanceHash,
		user.BalanceBTC,
		user.MiningToday,
		user.LastClick,
		user.MinerLevel,
		user.AdvertChannel,
		user.ReferralCount,
		user.TakeBonus,
		user.Language,
		user.RegisterTime,
		user.MinWithdrawal,
		user.FirstWithdrawal)
	if err != nil {
		return errors.Wrap(err, "query failed")
	}
	_ = rows.Close()

	if referralID == user.ID || referralID == 0 {
		return nil
	}

	baseUser, err := a.GetUser(referralID)
	if err != nil {
		return errors.Wrap(err, "get user")
	}
	baseUser.Balance += model.AdminSettings.GetParams(botLang).ReferralAmount
	rows, err = dataBase.Query("UPDATE users SET balance = ?, referral_count = ? WHERE id = ?;",
		baseUser.Balance, baseUser.ReferralCount+1, baseUser.ID)
	if err != nil {
		text := "Fatal Err with DB - auth.85 //" + err.Error()
		a.msgs.SendNotificationToDeveloper(text, false)
		return err
	}
	_ = rows.Close()

	return nil
}

func (a *Auth) pullReferralID(message *tgbotapi.Message) int64 {
	readParams := strings.Split(message.Text, " ")
	if len(readParams) < 2 {
		return 0
	}

	linkInfo, err := model.DecodeLink(a.bot.GetDataBase(), readParams[1])
	if err != nil || linkInfo == nil {
		if err != nil {
			a.msgs.SendNotificationToDeveloper("some err in decode link: "+err.Error(), false)
		}

		model.IncomeBySource.WithLabelValues(
			a.bot.BotLink,
			a.bot.BotLang,
			"unknown",
		).Inc()

		return 0
	}

	model.IncomeBySource.WithLabelValues(
		a.bot.BotLink,
		a.bot.BotLang,
		linkInfo.Source,
	).Inc()

	return linkInfo.ReferralID
}

func createSimpleUser(lang string, message *tgbotapi.Message) *model.User {
	return &model.User{
		ID:            message.From.ID,
		Language:      lang,
		AdvertChannel: rand.Intn(3) + 1,
	}
}

func (a *Auth) GetUser(id int64) (*model.User, error) {
	dataBase := a.bot.GetDataBase()
	rows, err := dataBase.Query(`
SELECT * FROM users
	WHERE id = ?;`,
		id)
	if err != nil {
		return nil, err
	}

	users, err := a.ReadUsers(rows)
	if err != nil || len(users) == 0 {
		return nil, model.ErrUserNotFound
	}
	return users[0], nil
}

func (a *Auth) ReadUsers(rows *sql.Rows) ([]*model.User, error) {
	defer rows.Close()

	var users []*model.User

	for rows.Next() {
		user := &model.User{}

		if err := rows.Scan(&user.ID,
			&user.Balance,
			&user.BalanceHash,
			&user.BalanceBTC,
			&user.MiningToday,
			&user.LastClick,
			&user.MinerLevel,
			&user.AdvertChannel,
			&user.ReferralCount,
			&user.TakeBonus,
			&user.Language,
			&user.RegisterTime,
			&user.MinWithdrawal,
			&user.FirstWithdrawal); err != nil {
			a.msgs.SendNotificationToDeveloper(errors.Wrap(err, "failed to scan row").Error(), false)
		}

		users = append(users, user)
	}

	return users, nil
}
