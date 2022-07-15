package model

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/Stepan1328/miner-bot/cfg"
	"github.com/go-redis/redis"
	_ "github.com/go-sql-driver/mysql"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	tokensPath       = "./cfg/tokens.json"
	dbDriver         = "mysql"
	redisDefaultAddr = "127.0.0.1:6379"

	statusDeleted = "deleted"
)

var Bots = make(map[string]*GlobalBot)

type GlobalBot struct {
	BotLang string

	Bot      *tgbotapi.BotAPI
	Chanel   tgbotapi.UpdatesChannel
	Rdb      *redis.Client
	DataBase *sql.DB

	MessageHandler  GlobalHandlers
	CallbackHandler GlobalHandlers

	AdminMessageHandler  GlobalHandlers
	AdminCallBackHandler GlobalHandlers

	Commands     map[string]string
	Language     map[string]map[string]string
	AdminLibrary map[string]map[string]string

	BotToken      string   `json:"bot_token"`
	BotLink       string   `json:"bot_link"`
	LanguageInBot []string `json:"language_in_bot"`

	MaintenanceMode bool
}

type GlobalHandlers interface {
	GetHandler(command string) Handler
}

type Handler func(situation *Situation) error

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
	dataBase.Exec("CREATE TABLE IF NOT EXISTS top (" + cfg.Top + ");")
	dataBase.Exec("CREATE INDEX balanceindex ON users (balance);")

	dataBase.Close()

	dataBase, err = sql.Open(dbDriver, cfg.DBCfg.User+cfg.DBCfg.Password+"@/"+cfg.DBCfg.Names[dbLang]) //TODO: refactor
	if err != nil {
		log.Fatalf("Failed open database: %s\n", err.Error())
	}

	dataBase, err = sql.Open(dbDriver, cfg.DBCfg.User+cfg.DBCfg.Password+"@/"+cfg.DBCfg.Names[dbLang]) //TODO: refactor
	if err != nil {
		log.Fatalf("Failed open database: %s\n", err.Error())
	}

	_, err = dataBase.Exec("ALTER TABLE users ADD COLUMN father_id bigint NOT NULL AFTER miner_level;")
	if err != nil && err.Error() != "Error 1060: Duplicate column name 'father_id'" {
		log.Fatalln(err)
	}

	_, err = dataBase.Exec("ALTER TABLE users ADD COLUMN all_referrals text NOT NULL AFTER father_id;")
	if err != nil && err.Error() != "Error 1060: Duplicate column name 'all_referrals'" {
		log.Fatalln(err)
	}

	migrateReferralFriends(dataBase)

	_, err = dataBase.Exec("ALTER TABLE users DROP COLUMN referral_count;")
	if err != nil && err.Error() != "Error 1091: Can't DROP 'referral_count'; check that column/key exists" {
		log.Fatalln(err)
	}

	err = dataBase.Ping()
	if err != nil {
		log.Fatalf("Failed upload database: %s\n", err.Error())
	}

	return dataBase
}

func migrateReferralFriends(dataBase *sql.DB) {
	rows, err := dataBase.Query(`SELECT * FROM users WHERE referral_count != 0;`)
	if err != nil && err.Error() != "Error 1054: Unknown column 'referral_count' in 'where clause'" {
		log.Fatalln(err)
	}

	if rows == nil {
		return
	}

	users, err := readCustomUser(rows)
	if err != nil {
		log.Fatalln(err)
	}

	go func() {
		for i, user := range users {
			_, err = dataBase.Exec(`UPDATE users SET all_referrals = ? WHERE id = ?;`, strconv.Itoa(user.ReferralCount), user.ID)
			if err != nil {
				log.Fatalln(err)
			}
			fmt.Println(user.ID, "    UPDATED NUMBER = ", i)
		}
	}()
}

type customUser struct {
	ID              int64   `json:"id"`
	Balance         int     `json:"balance"`
	BalanceHash     int     `json:"balance_hash"`
	BalanceBTC      float64 `json:"balance_btc"`
	MiningToday     int     `json:"mining_today"`
	LastClick       int64   `json:"last_click"`
	MinerLevel      int8    `json:"miner_level"`
	FatherID        int64   `json:"father_id"`
	AllReferrals    string  `json:"all_referrals"` // 10/20/30/40
	ReferralCount   int     `json:"referral_count"`
	AdvertChannel   int     `json:"advert_channel"`
	TakeBonus       bool    `json:"take_bonus"`
	Language        string  `json:"language"`
	RegisterTime    int64   `json:"register_time"`
	MinWithdrawal   int     `json:"min_withdrawal"`
	FirstWithdrawal bool    `json:"first_withdrawal"`
	Status          string  `json:"status"`
}

func readCustomUser(rows *sql.Rows) ([]*customUser, error) {
	defer rows.Close()

	var users []*customUser

	for rows.Next() {
		user := &customUser{}

		if err := rows.Scan(
			&user.ID,
			&user.Balance,
			&user.BalanceHash,
			&user.BalanceBTC,
			&user.MiningToday,
			&user.LastClick,
			&user.MinerLevel,
			&user.FatherID,
			&user.AllReferrals,
			&user.ReferralCount,
			&user.AdvertChannel,
			&user.TakeBonus,
			&user.Language,
			&user.RegisterTime,
			&user.MinWithdrawal,
			&user.FirstWithdrawal,
			&user.Status); err != nil {
			log.Fatalln(err)
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

func FillBotsConfig() {
	bytes, err := os.ReadFile(tokensPath)
	if err != nil {
		log.Fatal(err)
	}

	err = json.Unmarshal(bytes, &Bots)
	if err != nil {
		log.Fatal(err)
	}

	for lang, bot := range Bots {
		bot.BotLang = lang
	}
}

func (b *GlobalBot) GetBotLang() string {
	return b.BotLang
}

func (b *GlobalBot) GetBot() *tgbotapi.BotAPI {
	return b.Bot
}

func (b *GlobalBot) GetDataBase() *sql.DB {
	return b.DataBase
}

func (b *GlobalBot) AvailableLang() []string {
	return b.LanguageInBot
}

func (b *GlobalBot) GetCurrency() string {
	return AdminSettings.GetCurrency(b.BotLang)
}

func (b *GlobalBot) LangText(lang, key string, values ...interface{}) string {
	formatText := b.Language[lang][key]
	return fmt.Sprintf(formatText, values...)
}

func (b *GlobalBot) GetTexts(lang string) map[string]string {
	return b.Language[lang]
}

func (b *GlobalBot) CheckAdmin(userID int64) bool {
	_, exist := AdminSettings.AdminID[userID]
	return exist
}

func (b *GlobalBot) AdminLang(userID int64) string {
	return AdminSettings.AdminID[userID].Language
}

func (b *GlobalBot) AdminText(adminLang, key string) string {
	return b.AdminLibrary[adminLang][key]
}

func (b *GlobalBot) UpdateBlockedUsers(channel int) {
}

func (b *GlobalBot) GetAdvertURL(userLang string, channel int) string {
	return AdminSettings.GetAdvertUrl(userLang, channel)
}

func (b *GlobalBot) GetAdvertText(userLang string, channel int) string {
	return AdminSettings.GetAdvertText(userLang, channel)
}

func (b *GlobalBot) GetAdvertisingPhoto(lang string, channel int) string {
	return AdminSettings.GlobalParameters[lang].AdvertisingPhoto[channel]
}

func (b *GlobalBot) GetAdvertisingVideo(lang string, channel int) string {
	return AdminSettings.GlobalParameters[lang].AdvertisingVideo[channel]
}

func (b *GlobalBot) ButtonUnderAdvert() bool {
	return AdminSettings.GlobalParameters[b.BotLang].Parameters.ButtonUnderAdvert
}

func (b *GlobalBot) AdvertisingChoice(channel int) string {
	return AdminSettings.GlobalParameters[b.BotLang].AdvertisingChoice[channel]
}

func (b *GlobalBot) BlockUser(userID int64) error {
	_, err := b.GetDataBase().Exec(`
UPDATE users
	SET status = ?
WHERE id = ?`,
		statusDeleted,
		userID)

	return err
}

func (b *GlobalBot) GetMetrics(metricKey string) *prometheus.CounterVec {
	metricsByKey := map[string]*prometheus.CounterVec{
		"total_mailing_users": MailToUser,
		"total_block_users":   BlockUser,
	}

	return metricsByKey[metricKey]
}
