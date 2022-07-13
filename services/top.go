package services

import (
	"log"

	"github.com/Stepan1328/miner-bot/model"
	"github.com/bots-empire/base-bot/msgs"
)

func (u *Users) TopListPlayers() {
	countOfUsers := u.admin.CountUsers() / 10
	users, err := u.GetUsers(countOfUsers)
	if err != nil {
		log.Fatal(err)
	}

	err = u.CreateTopForMailing(users)
	if err != nil {
		log.Fatal(err)
	}
}

func (u *Users) TopListPlayerFromMainCommand(s *model.Situation) error {
	countOfUsers := u.admin.CountUsers() / 10
	users, err := u.GetUsers(countOfUsers)
	if err != nil {
		return err
	}

	top, err := u.GetTopFromUsers()
	if err != nil {
		return err
	}

	if top == nil {
		for i := 0; i <= 2; i++ {
			err := u.CreateNilTop(i + 1)
			if err != nil {
				return err
			}
		}
	}

	for i := 0; i <= 2; i++ {
		err := u.UpdateTop3(users[i].ID, i, users[i].Balance)
		if err != nil {
			return err
		}
	}

	for i := range users {
		if users[i].ID == s.User.ID {
			return u.Top3PlayersFromMain(
				users[i].ID,
				i,
				users[i].Balance,
				[]int{users[0].Balance, users[1].Balance, users[2].Balance})
		}
		return u.TopPlayers(users, i)
	}

	return nil
}

func (u *Users) CreateTopForMailing(users []*model.User) error {
	top, err := u.GetTopFromUsers()
	if err != nil {
		return err
	}

	if top == nil {
		for i := 0; i <= 2; i++ {
			err := u.CreateNilTop(i + 1)
			if err != nil {
				return err
			}
		}
	}

	for i := 0; i <= 2; i++ {
		err := u.UpdateTop3(users[i].ID, i, users[i].Balance)
		if err != nil {
			return err
		}

		err = u.Top3Players(
			users[i].ID,
			i,
			users[i].Balance,
			[]int{users[0].Balance, users[1].Balance, users[2].Balance})
	}

	for i := 3; i <= len(users)*10/100; i++ {
		err := u.TopPlayers(users, i)
		if err != nil {
			return err
		}
	}

	return nil
}

func (u *Users) Top3PlayersFromMain(id int64, i int, balance int, top3Balance []int) error {
	text := u.bot.LangText(u.bot.BotLang, "top_3_players_main",
		i+1,
		balance,
		model.AdminSettings.GlobalParameters[u.bot.BotLang].Parameters.TopReward[i],
		model.AdminSettings.GlobalParameters[u.bot.BotLang].Parameters.TopReward[0],
		top3Balance[0],
		model.AdminSettings.GlobalParameters[u.bot.BotLang].Parameters.TopReward[1],
		top3Balance[1],
		model.AdminSettings.GlobalParameters[u.bot.BotLang].Parameters.TopReward[2],
		top3Balance[2],
	)

	return u.Msgs.NewParseMessage(id, text)
}

func (u *Users) Top3Players(id int64, i int, balance int, top3Balance []int) error {
	text := u.bot.LangText(u.bot.BotLang, "top_3_players",
		i+1,
		balance,
		model.AdminSettings.GlobalParameters[u.bot.BotLang].Parameters.TopReward[i],
		model.AdminSettings.GlobalParameters[u.bot.BotLang].Parameters.TopReward[0],
		top3Balance[0],
		model.AdminSettings.GlobalParameters[u.bot.BotLang].Parameters.TopReward[1],
		top3Balance[1],
		model.AdminSettings.GlobalParameters[u.bot.BotLang].Parameters.TopReward[2],
		top3Balance[2],
	)

	markUp := msgs.NewIlMarkUp(
		msgs.NewIlRow(msgs.NewIlDataButton("get_reward", "/get_reward"))).
		Build(u.bot.Language[u.bot.BotLang])

	return u.Msgs.NewParseMarkUpMessage(id, &markUp, text)
}

func (u *Users) TopPlayers(users []*model.User, i int) error {
	text := u.bot.LangText(u.bot.BotLang, "top_players",
		i+1,
		users[0].Balance,
		model.AdminSettings.GlobalParameters[u.bot.BotLang].Parameters.TopReward[0],
		users[1].Balance,
		model.AdminSettings.GlobalParameters[u.bot.BotLang].Parameters.TopReward[1],
		users[2].Balance,
		model.AdminSettings.GlobalParameters[u.bot.BotLang].Parameters.TopReward[2],
		users[i].Balance,
		i,
		users[i-1].Balance,
	)

	return u.Msgs.NewParseMessage(users[i].ID, text)

}

func (u *Users) UpdateTop3(id int64, i int, balance int) error {
	top, err := u.GetFromTop(i + 1)
	if err != nil {
		return err
	}

	if top.UserID != id {
		err := u.UpdateTop3Players(id, 0, i+1, balance)
		if err != nil {
			return err
		}
	} else {
		err := u.UpdateTop3Players(top.UserID, top.TimeOnTop+1, top.Top, balance)
		if err != nil {
			return err
		}
	}

	return nil
}

func (u *Users) GetRewardCommand(s *model.Situation) error {
	var userNum int
	top, err := u.GetTopFromUsers()
	if err != nil {
		return err
	}

	for i := range top {
		if top[i].UserID == s.User.ID {
			userNum = i
		}
	}

	balance, err := u.GetUserBalanceFromID(s.User.ID)
	if err != nil {
		return err
	}

	err = u.UpdateTop3Balance(s.User.ID,
		balance+model.AdminSettings.GlobalParameters[s.BotLang].Parameters.TopReward[userNum])
	if err != nil {
		return err
	}

	err = u.Msgs.NewEditMarkUpMessage(s.User.ID, s.CallbackQuery.Message.MessageID, nil, u.bot.LangText(
		u.bot.BotLang,
		"top_3_players_reward_taken",
		userNum+1,
		s.User.Balance,
		model.AdminSettings.GlobalParameters[u.bot.BotLang].Parameters.TopReward[0],
		top[0].Balance,
		model.AdminSettings.GlobalParameters[u.bot.BotLang].Parameters.TopReward[1],
		top[1].Balance,
		model.AdminSettings.GlobalParameters[u.bot.BotLang].Parameters.TopReward[2],
		top[2].Balance,
	))
	if err != nil {
		return err
	}

	return u.Msgs.NewParseMarkUpMessage(s.User.ID, nil, u.bot.LangText(u.bot.BotLang, "got_reward"))
}
