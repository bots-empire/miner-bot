package db

import (
	"log"
	"strconv"
	"time"

	"github.com/Stepan1328/miner-bot/model"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	emptyLevelName = "empty"
)

func RdbSetUser(botLang string, ID int64, level string) {
	userID := userIDToRdb(botLang, ID)
	_, err := model.Bots[botLang].Rdb.Set(userID, level, 0).Result()
	if err != nil {
		log.Println(err)
	}
}

func userIDToRdb(botLang string, userID int64) string {
	return botLang + ":user:" + strconv.FormatInt(userID, 10)
}

func GetLevel(botLang string, id int64) string {
	userID := userIDToRdb(botLang, id)
	have, err := model.Bots[botLang].Rdb.Exists(userID).Result()
	if err != nil {
		log.Println(err)
	}
	if have == 0 {
		return emptyLevelName
	}

	value, err := model.Bots[botLang].Rdb.Get(userID).Result()
	if err != nil {
		log.Println(err)
	}
	return value
}

func SaveUserClickerMsgID(botLang string, userID int64, msgID int) {
	userClickerID := userClickerMsgIDToRdb(botLang, userID)
	_, err := model.Bots[botLang].Rdb.Set(userClickerID, strconv.Itoa(msgID), 0).Result()
	if err != nil {
		log.Println(err)
	}

	UpdateClickerMsgTTL(botLang, userID)
}

func UpdateClickerMsgTTL(botLang string, userID int64) {
	userClickerID := userClickerMsgIDToRdb(botLang, userID)

	_, err := model.Bots[botLang].Rdb.Expire(userClickerID, time.Minute*3).Result()
	if err != nil {
		log.Println(err)
	}
}

func userClickerMsgIDToRdb(botLang string, userID int64) string {
	return botLang + ":user_clicker_id:" + strconv.FormatInt(userID, 10)
}

func GetUserClickerMsgID(botLang string, userID int64) int {
	userClickerID := userClickerMsgIDToRdb(botLang, userID)
	result, err := model.Bots[botLang].Rdb.Get(userClickerID).Result()
	if err != nil {
		log.Println(err)
	}
	msgID, _ := strconv.Atoi(result)
	return msgID
}

func RdbSetAdminMsgID(botLang string, userID int64, msgID int) {
	adminMsgID := adminMsgIDToRdb(botLang, userID)
	_, err := model.Bots[botLang].Rdb.Set(adminMsgID, strconv.Itoa(msgID), 0).Result()
	if err != nil {
		log.Println(err)
	}
}

func adminMsgIDToRdb(botLang string, userID int64) string {
	return botLang + ":admin_msg_id:" + strconv.FormatInt(userID, 10)
}

func RdbGetAdminMsgID(botLang string, userID int64) int {
	adminMsgID := adminMsgIDToRdb(botLang, userID)
	result, err := model.Bots[botLang].Rdb.Get(adminMsgID).Result()
	if err != nil {
		log.Println(err)
	}
	msgID, _ := strconv.Atoi(result)
	return msgID
}

func DeleteOldAdminMsg(botLang string, userID int64) {
	adminMsgID := adminMsgIDToRdb(botLang, userID)
	result, err := model.Bots[botLang].Rdb.Get(adminMsgID).Result()
	if err != nil {
		log.Println(err)
	}

	if oldMsgID, _ := strconv.Atoi(result); oldMsgID != 0 {
		msg := tgbotapi.NewDeleteMessage(userID, oldMsgID)

		if _, err = model.Bots[botLang].Bot.Send(msg); err != nil {
			log.Println(err)
		}
		RdbSetAdminMsgID(botLang, userID, 0)
	}
}

func RdbSetMinerLevelSetting(botLang string, userID int64, level int) {
	minerLevel := minerLevelSettingToRdb(botLang, userID)
	_, err := model.Bots[botLang].Rdb.Set(minerLevel, strconv.Itoa(level), 0).Result()
	if err != nil {
		log.Println(err)
	}
}

func minerLevelSettingToRdb(botLang string, userID int64) string {
	return botLang + ":miner_level_setting:" + strconv.FormatInt(userID, 10)
}

func RdbGetMinerLevelSetting(botLang string, userID int64) int {
	minerLevel := minerLevelSettingToRdb(botLang, userID)
	result, err := model.Bots[botLang].Rdb.Get(minerLevel).Result()
	if err != nil {
		log.Println(err)
	}
	level, _ := strconv.Atoi(result)
	return level
}
