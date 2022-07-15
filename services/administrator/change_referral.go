package administrator

import (
	"encoding/json"
	"strconv"

	"github.com/Stepan1328/miner-bot/model"
)

func RdbSetRewardGap(botLang string, userID int64, rewardsGap *model.RewardsGap) error {
	refCounter := referralCounterToRdb(userID)
	value, err := json.Marshal(rewardsGap)
	if err != nil {
		return err
	}

	_, err = model.Bots[botLang].Rdb.Set(refCounter, value, 0).Result()
	if err != nil {
		return err
	}

	return nil
}

func referralCounterToRdb(userID int64) string {
	return "reward_counter:" + strconv.FormatInt(userID, 10)
}

func RdbGetRewardGap(botLang string, userID int64) (*model.RewardsGap, error) {
	RefCounter := referralCounterToRdb(userID)
	result, err := model.Bots[botLang].Rdb.Get(RefCounter).Result()
	if err != nil {
		if err.Error() == model.ErrRedisNil.Error() {
			return nil, nil
		}
		return nil, err
	}

	ref := &model.RewardsGap{}

	err = json.Unmarshal([]byte(result), ref)
	if err != nil {
		return nil, err
	}

	return ref, nil
}
