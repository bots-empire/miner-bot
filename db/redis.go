package db

import (
	"log"
	"strconv"

	"github.com/Stepan1328/miner-bot/model"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	emptyLevelName = "empty"
)

// TODO: make one redis client: add to key botLang

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
