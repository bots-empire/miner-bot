package assets

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/Stepan1328/miner-bot/model"
)

const (
	adminPath      = "assets/admin"
	jsonFormatName = ".json"

	oneSatoshi = 0.00000001
)

type Admin struct {
	AdminID          map[int64]*AdminUser         `json:"admin_id"`
	GlobalParameters map[string]*GlobalParameters `json:"global_parameters"`
}

type GlobalParameters struct {
	Parameters      *Params        `json:"parameters"`
	AdvertisingChan *AdvertChannel `json:"advertising_chan"`
	BlockedUsers    int            `json:"blocked_users"`
	AdvertisingText string         `json:"advertising_text"`
}

type AdminUser struct {
	Language           string `json:"language"`
	FirstName          string `json:"first_name"`
	SpecialPossibility bool   `json:"special_possibility"`
}

type Params struct {
	BonusAmount         int   `json:"bonus_amount"`
	MinWithdrawalAmount int   `json:"min_withdrawal_amount"`
	ClickAmount         []int `json:"click_amount"`
	UpgradeMinerCost    []int `json:"upgrade_miner_cost"`
	MaxOfClickPerDay    int   `json:"max_of_click_per_day"`
	ReferralAmount      int   `json:"referral_amount"`

	ExchangeHashToBTC     int     `json:"exchange_hash_to_btc"`     // 0.00000001 BTC = ExchangeHashToBTC hashes
	ExchangeBTCToCurrency float64 `json:"exchange_btc_to_currency"` // 0.00000001 * ExchangeBTCToCurrency BTC = 1 USD/EUR

	Currency string `json:"currency"`
}

type AdvertChannel struct {
	Url       string `json:"url"`
	ChannelID int64  `json:"channel_id"`
}

var AdminSettings *Admin

func UploadAdminSettings() {
	var settings *Admin
	data, err := os.ReadFile(adminPath + jsonFormatName)
	if err != nil {
		fmt.Println(err)
	}

	err = json.Unmarshal(data, &settings)
	if err != nil {
		fmt.Println(err)
	}

	for lang, globalBot := range model.Bots {
		validateSettings(settings, lang)
		for _, lang = range globalBot.LanguageInBot {
			validateSettings(settings, lang)
		}
	}

	AdminSettings = settings
	SaveAdminSettings()
}

func validateSettings(settings *Admin, lang string) {
	if settings.GlobalParameters == nil {
		settings.GlobalParameters = make(map[string]*GlobalParameters)
	}

	if settings.GlobalParameters[lang] == nil {
		settings.GlobalParameters[lang] = &GlobalParameters{}
	}

	if settings.GlobalParameters[lang].Parameters == nil {
		settings.GlobalParameters[lang].Parameters = &Params{
			ClickAmount:           []int{1},
			UpgradeMinerCost:      []int{0},
			ExchangeHashToBTC:     1,
			ExchangeBTCToCurrency: oneSatoshi,
		}
	}

	if settings.GlobalParameters[lang].AdvertisingChan == nil {
		settings.GlobalParameters[lang].AdvertisingChan = &AdvertChannel{
			Url: "https://google.com",
		}
	}
}

func SaveAdminSettings() {
	data, err := json.MarshalIndent(AdminSettings, "", "  ")
	if err != nil {
		panic(err)
	}

	if err = os.WriteFile(adminPath+jsonFormatName, data, 0600); err != nil {
		panic(err)
	}
}

func (a *Admin) GetCurrency(lang string) string {
	return a.GlobalParameters[lang].Parameters.Currency
}

func (a *Admin) GetAdvertText(lang string) string {
	return a.GlobalParameters[lang].AdvertisingText
}

func (a *Admin) UpdateAdvertText(lang string, value string) {
	a.GlobalParameters[lang].AdvertisingText = value
}

func (a *Admin) GetAdvertUrl(lang string) string {
	return a.GlobalParameters[lang].AdvertisingChan.Url
}

func (a *Admin) GetAdvertChannelID(lang string) int64 {
	return a.GlobalParameters[lang].AdvertisingChan.ChannelID
}

func (a *Admin) UpdateAdvertChan(lang string, newChan *AdvertChannel) {
	a.GlobalParameters[lang].AdvertisingChan = newChan
}

func (a *Admin) UpdateBlockedUsers(lang string, value int) {
	a.GlobalParameters[lang].BlockedUsers = value
}

func (a *Admin) GetParams(lang string) *Params {
	return a.GlobalParameters[lang].Parameters
}

func (a *Admin) GetClickAmount(lang string, level int) int {
	if level < 0 {
		level = 0
	}

	if level > len(a.GlobalParameters[lang].Parameters.ClickAmount)-1 {
		level = len(a.GlobalParameters[lang].Parameters.ClickAmount) - 1
	}

	return a.GlobalParameters[lang].Parameters.ClickAmount[level]
}

func (a *Admin) GetUpgradeCost(lang string, level int) int {
	if level < 0 {
		level = 0
	}

	if level > len(a.GlobalParameters[lang].Parameters.UpgradeMinerCost)-1 {
		level = len(a.GlobalParameters[lang].Parameters.UpgradeMinerCost) - 1
	}

	return a.GlobalParameters[lang].Parameters.UpgradeMinerCost[level]
}

func (a *Admin) GetMaxMinerLevel(lang string) int {
	clickLen := len(a.GlobalParameters[lang].Parameters.ClickAmount)
	upgradeLen := len(a.GlobalParameters[lang].Parameters.UpgradeMinerCost)

	return minFromTwo(clickLen, upgradeLen)
}

func minFromTwo(a, b int) int {
	if a < b {
		return a
	}

	return b
}

func (a *Admin) AddMinerLevel(lang string, level int) {
	if level == a.GetMaxMinerLevel(lang)-1 {
		a.GlobalParameters[lang].Parameters.ClickAmount = append(
			a.GlobalParameters[lang].Parameters.ClickAmount,
			a.GlobalParameters[lang].Parameters.ClickAmount[level])

		a.GlobalParameters[lang].Parameters.UpgradeMinerCost = append(
			a.GlobalParameters[lang].Parameters.UpgradeMinerCost,
			a.GlobalParameters[lang].Parameters.UpgradeMinerCost[level])
		return
	}

	if level == 0 {
		a.GlobalParameters[lang].Parameters.ClickAmount = append(
			[]int{a.GlobalParameters[lang].Parameters.ClickAmount[0]},
			a.GlobalParameters[lang].Parameters.ClickAmount...)

		a.GlobalParameters[lang].Parameters.UpgradeMinerCost = append(
			[]int{a.GlobalParameters[lang].Parameters.UpgradeMinerCost[0]},
			a.GlobalParameters[lang].Parameters.UpgradeMinerCost...)
		return
	}

	copyOfClick := a.GlobalParameters[lang].Parameters.ClickAmount
	copyOfClick = append(copyOfClick[:level+1], copyOfClick[level:]...)
	copyOfClick[level] = copyOfClick[level-1]
	a.GlobalParameters[lang].Parameters.ClickAmount = copyOfClick

	copyOfUpgrade := a.GlobalParameters[lang].Parameters.UpgradeMinerCost
	copyOfUpgrade = append(copyOfUpgrade[:level+1], copyOfUpgrade[level:]...)
	copyOfUpgrade[level] = copyOfUpgrade[level-1]
	a.GlobalParameters[lang].Parameters.UpgradeMinerCost = copyOfUpgrade
}

func (a *Admin) DeleteMinerLevel(lang string, level int) {
	a.GlobalParameters[lang].Parameters.ClickAmount = remove(
		a.GlobalParameters[lang].Parameters.ClickAmount,
		level)

	a.GlobalParameters[lang].Parameters.UpgradeMinerCost = remove(
		a.GlobalParameters[lang].Parameters.UpgradeMinerCost,
		level)
}

func remove(slice []int, s int) []int {
	return append(slice[:s], slice[s+1:]...)
}

// ----------------------------------------------------
//
// Update Statistic
//
// ----------------------------------------------------

type UpdateInfo struct {
	Mu      *sync.Mutex
	Counter int
	Day     int
}

var UpdateStatistic *UpdateInfo

func UploadUpdateStatistic() {
	info := &UpdateInfo{}
	info.Mu = new(sync.Mutex)
	strStatistic, err := model.Bots["it"].Rdb.Get("update_statistic").Result()
	if err != nil {
		UpdateStatistic = info
		return
	}

	data := strings.Split(strStatistic, "?")
	if len(data) != 2 {
		UpdateStatistic = info
		return
	}
	info.Counter, _ = strconv.Atoi(data[0])
	info.Day, _ = strconv.Atoi(data[1])
	UpdateStatistic = info
}

func SaveUpdateStatistic() {
	strStatistic := strconv.Itoa(UpdateStatistic.Counter) + "?" + strconv.Itoa(UpdateStatistic.Day)
	_, err := model.Bots["it"].Rdb.Set("update_statistic", strStatistic, 0).Result()
	if err != nil {
		log.Println(err)
	}
}
