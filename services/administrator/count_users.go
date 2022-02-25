package administrator

import (
	"database/sql"
	"log"
	"strconv"

	"github.com/Stepan1328/miner-bot/assets"
	"github.com/Stepan1328/miner-bot/model"
	"github.com/Stepan1328/miner-bot/msgs"
	"github.com/pkg/errors"
)

func countUsers(botLang string) int {
	dataBase := model.GetDB(botLang)
	rows, err := dataBase.Query(`
SELECT COUNT(*) FROM users;`)
	if err != nil {
		log.Println(err.Error())
	}
	return readRows(rows)
}

func readRows(rows *sql.Rows) int {
	defer rows.Close()

	var count int

	for rows.Next() {
		if err := rows.Scan(&count); err != nil {
			msgs.SendNotificationToDeveloper(errors.Wrap(err, "failed to scan row").Error())
		}
	}

	return count
}

func countAllUsers() int {
	var sum int
	for _, handler := range model.Bots {
		rows, err := handler.DataBase.Query(`
SELECT COUNT(*) FROM users;`)
		if err != nil {
			log.Println(err.Error())
		}
		sum += readRows(rows)
	}
	return sum
}

func countReferrals(botLang string, amountUsers int) string {
	var refText string
	rows, err := model.Bots[botLang].DataBase.Query("SELECT SUM(referral_count) FROM users;")
	if err != nil {
		log.Println(err.Error())
	}

	count := readRows(rows)
	refText = strconv.Itoa(count) + " (" + strconv.Itoa(int(float32(count)*100.0/float32(amountUsers))) + "%)"
	return refText
}

func countBlockedUsers(botLang string) int {
	//var count int
	//for _, value := range assets.AdminSettings.BlockedUsers {
	//	count += value
	//}
	//return count
	return assets.AdminSettings.BlockedUsers[botLang]
}

func countSubscribers(botLang string) int {
	rows, err := model.Bots[botLang].DataBase.Query(`
SELECT COUNT(DISTINCT id) FROM subs;`)
	if err != nil {
		log.Println(err.Error())
	}

	return readRows(rows)
}
