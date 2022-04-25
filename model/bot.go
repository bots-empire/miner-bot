package model

import (
	"database/sql"
	"encoding/json"
	"log"
	"math/rand"
	"os"

	"github.com/Stepan1328/miner-bot/cfg"
	"github.com/go-redis/redis"
	_ "github.com/go-sql-driver/mysql"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	tokensPath       = "./cfg/tokens.json"
	dbDriver         = "mysql"
	redisDefaultAddr = "127.0.0.1:6379"
)

var Bots = make(map[string]*GlobalBot)

type GlobalBot struct {
	Bot      *tgbotapi.BotAPI
	Chanel   tgbotapi.UpdatesChannel
	Rdb      *redis.Client
	DataBase *sql.DB

	MessageHandler  GlobalHandlers
	CallbackHandler GlobalHandlers

	AdminMessageHandler  GlobalHandlers
	AdminCallBackHandler GlobalHandlers

	BotToken      string   `json:"bot_token"`
	BotLink       string   `json:"bot_link"`
	LanguageInBot []string `json:"language_in_bot"`

	MaintenanceMode bool
}

type GlobalHandlers interface {
	GetHandler(command string) Handler
}

type Handler interface {
	Serve(situation *Situation) error
}

func UploadDataBase(dbLang string) *sql.DB {
	dataBase, err := sql.Open(dbDriver, cfg.DBCfg.User+cfg.DBCfg.Password+"@/") //TODO: refactor
	if err != nil {
		log.Fatalf("Failed open database: %s\n", err.Error())
	}

	dataBase.Exec("CREATE DATABASE IF NOT EXISTS " + cfg.DBCfg.Names[dbLang] + ";")
	dataBase.Exec("USE " + cfg.DBCfg.Names[dbLang] + ";")
	dataBase.Exec("CREATE TABLE IF NOT EXISTS users (" + cfg.UserTable + ");")
	dataBase.Exec("CREATE TABLE IF NOT EXISTS links (" + cfg.Links + ");")
	dataBase.Exec("CREATE TABLE IF NOT EXISTS subs (" + cfg.Subs + ");")

	dataBase.Close()

	dataBase, err = sql.Open(dbDriver, cfg.DBCfg.User+cfg.DBCfg.Password+"@/"+cfg.DBCfg.Names[dbLang]) //TODO: refactor
	if err != nil {
		log.Fatalf("Failed open database: %s\n", err.Error())
	}

	TakeAllUsers(dataBase)

	err = dataBase.Ping()
	if err != nil {
		log.Fatalf("Failed upload database: %s\n", err.Error())
	}

	return dataBase
}

func TakeAllUsers(dataBase *sql.DB) {
	rows, err := dataBase.Query(`SELECT * FROM users WHERE advert_channel = 0;`)
	if err != nil {
	}

	if rows == nil {
		return
	}

	users, err := ReadUser(rows)
	if err != nil {
	}

	for i := range users {
		dataBase.Exec(`UPDATE users SET advert_channel = ? WHERE id = ?;`, rand.Intn(3)+1, users[i].ID)
	}
}

func ReadUser(rows *sql.Rows) ([]*User, error) {
	defer rows.Close()

	var users []*User

	for rows.Next() {
		user := &User{}

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
			//msgs.SendNotificationToDeveloper(errors.Wrap(err, "failed to scan row").Error())
		}

		users = append(users, user)
	}

	return users, nil
}

func StartRedis() *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     redisDefaultAddr,
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	return rdb
}

func GetDB(botLang string) *sql.DB {
	return Bots[botLang].DataBase
}

func FillBotsConfig() {
	bytes, err := os.ReadFile(tokensPath)
	if err != nil {
		log.Fatal(err)
	}

	err = json.Unmarshal(bytes, &Bots)
	if err != nil {
		panic(err)
	}
}

func GetGlobalBot(botLang string) *GlobalBot {
	return Bots[botLang]
}
