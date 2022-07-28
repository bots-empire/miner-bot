package administrator

import (
	"github.com/Stepan1328/miner-bot/model"
	"github.com/pkg/errors"
)

func (a *Admin) getIncomeInfo(userID int64) (*model.IncomeInfo, error) {
	info := &model.IncomeInfo{UserID: userID}
	err := a.bot.GetDataBase().QueryRow(`
SELECT (source)
	FROM income_info
WHERE user_id = ?`,
		userID).
		Scan(&info.Source)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return nil, nil
		}

		return nil, errors.Wrap(err, "failed query row")
	}

	return info, nil
}
