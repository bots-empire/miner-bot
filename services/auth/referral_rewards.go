package auth

import (
	"database/sql"
	"strconv"
	"strings"

	"github.com/Stepan1328/miner-bot/model"
)

func (a *Auth) referralRewardSystem(botLang string, userID int64, lvl int) error {
	user, err := a.GetUser(userID)
	if err != nil {
		return err
	}

	refByLvl := allReferralsByLvl(user.AllReferrals)
	refByLvl = increaseReferralOnLvl(refByLvl, lvl)

	_, err = a.bot.GetDataBase().Exec(`
UPDATE users SET
	balance = balance + ?,
	all_referrals = ?
WHERE id = ?;`,
		model.AdminSettings.GetParams(botLang).ReferralReward.GetReward(lvl, refByLvl[lvl-1]),
		refByLvlToString(refByLvl),
		userID)
	if err != nil {
		return err
	}

	rows, err := a.bot.GetDataBase().Query(`
SELECT father_id 
	FROM users 
WHERE id = ?`,
		userID)
	if err != nil {
		return err
	}

	fatherID, err := getFatherIDFromRow(rows)
	if err != nil {
		return err
	}

	if lvl == model.AdminSettings.GetParams(botLang).ReferralReward.MaxLevel() {
		return nil
	}

	return a.referralRewardSystem(botLang, fatherID, lvl+1)
}

func allReferralsByLvl(rawReferrals string) []int {
	rawLvls := strings.Split(rawReferrals, "/")
	if len(rawLvls) == 1 && rawLvls[0] == "" {
		return []int{}
	}

	byLvl := make([]int, 0)
	for _, rawLvl := range rawLvls {
		count, _ := strconv.Atoi(rawLvl)
		byLvl = append(byLvl, count)
	}

	return byLvl
}

func increaseReferralOnLvl(refByLvl []int, lvl int) []int {
	if len(refByLvl) < lvl {
		return append(refByLvl, 1)
	}

	refByLvl[lvl-1]++
	return refByLvl
}

func refByLvlToString(refByLvl []int) string {
	var rawLvls []string

	for _, count := range refByLvl {
		rawLvls = append(rawLvls, strconv.Itoa(count))
	}

	return strings.Join(rawLvls, "/")
}

func getFatherIDFromRow(rows *sql.Rows) (int64, error) {
	defer rows.Close()

	var fatherID int64
	for rows.Next() {
		if err := rows.Scan(&fatherID); err != nil {
			return 0, err
		}
	}

	return fatherID, nil
}
