package model

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
)

const (
	adminPath      = "assets/admin"
	jsonFormatName = ".json"

	oneSatoshi    = 0.00000001
	GlobalMailing = 4
	MainAdvert    = 5
)

type Admin struct {
	AdminID          map[int64]*AdminUser         `json:"admin_id"`
	GlobalParameters map[string]*GlobalParameters `json:"global_parameters"`
}

type GlobalParameters struct {
	Parameters        *Params        `json:"parameters"`
	AdvertisingChan   *AdvertChannel `json:"advertising_chan"`
	BlockedUsers      int            `json:"blocked_users"`
	AdvertisingText   map[int]string `json:"advertising_text"`
	AdvertisingPhoto  map[int]string
	AdvertisingVideo  map[int]string
	AdvertisingChoice map[int]string
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

	ButtonUnderAdvert bool

	ExchangeHashToBTC     int     `json:"exchange_hash_to_btc"`     // 0.00000001 BTC = ExchangeHashToBTC hashes
	ExchangeBTCToCurrency float64 `json:"exchange_btc_to_currency"` // 0.00000001 * ExchangeBTCToCurrency BTC = 1 USD/EUR

	Currency string `json:"currency"`

	TopReward []int `json:"top_reward"`
}

type AdvertChannel struct {
	Url       map[int]string `json:"url"`
	ChannelID map[int]int64  `json:"channel_id"`
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

	for lang, globalBot := range Bots {
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
			TopReward:             []int{10, 10, 10},
		}
	}

	if settings.GlobalParameters[lang].AdvertisingChan == nil {
		settings.GlobalParameters[lang].AdvertisingChan = &AdvertChannel{
			Url: map[int]string{
				0: "https://google.com",
				1: "https://google.com",
				2: "https://google.com",
				5: "https://google.com"},
			ChannelID: make(map[int]int64),
		}
	}

	if settings.GlobalParameters[lang].AdvertisingChoice == nil {
		settings.GlobalParameters[lang].AdvertisingChoice = make(map[int]string)
	}

	if settings.GlobalParameters[lang].AdvertisingText == nil {
		settings.GlobalParameters[lang].AdvertisingText = make(map[int]string)
	}

	if settings.GlobalParameters[lang].AdvertisingChan == nil {
		settings.GlobalParameters[lang].AdvertisingChan = &AdvertChannel{}
	}

	if settings.GlobalParameters[lang].AdvertisingPhoto == nil {
		settings.GlobalParameters[lang].AdvertisingPhoto = make(map[int]string)
	}
	if settings.GlobalParameters[lang].AdvertisingVideo == nil {
		settings.GlobalParameters[lang].AdvertisingVideo = make(map[int]string)
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

func (a *Admin) GetAdvertText(lang string, channel int) string {
	return a.GlobalParameters[lang].AdvertisingText[channel]
}

func (a *Admin) UpdateAdvertUrl(lang string, channel int, value string) {
	a.GlobalParameters[lang].AdvertisingChan.Url[channel] = value
}

func (a *Admin) UpdateAdvertChannelID(lang string, value int64, channel int) {
	a.GlobalParameters[lang].AdvertisingChan.ChannelID[channel] = value
}

func (a *Admin) UpdateAdvertText(lang string, value string, channel int) {
	a.GlobalParameters[lang].AdvertisingText[channel] = value
}

func (a *Admin) UpdateAdvertPhoto(lang string, channel int, value string) {
	a.GlobalParameters[lang].AdvertisingPhoto[channel] = value
}

func (a *Admin) UpdateAdvertVideo(lang string, channel int, value string) {
	a.GlobalParameters[lang].AdvertisingVideo[channel] = value
}

func (a *Admin) UpdateAdvertChoice(lang string, channel int, value string) {
	a.GlobalParameters[lang].AdvertisingChoice[channel] = value
}

func (a *Admin) GetAdvertUrl(lang string, channel int) string {
	return a.GlobalParameters[lang].AdvertisingChan.Url[channel]
}

func (a *Admin) GetAdvertChannelID(lang string, channel int) int64 {
	return a.GlobalParameters[lang].AdvertisingChan.ChannelID[channel]
}

func (a *Admin) UpdateAdvertChan(lang string, newChan *AdvertChannel) {
	a.GlobalParameters[lang].AdvertisingChan = newChan
}

func (a *Admin) UpdateBlockedUsers(lang string, value int) {
	a.GlobalParameters[lang].BlockedUsers = value
}

func (a *Admin) UpdateTopRewardSetting(lang string, i int, value int) {
	a.GlobalParameters[lang].Parameters.TopReward[i] = value
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
	strStatistic, err := Bots["it"].Rdb.Get("update_statistic").Result()
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
	_, err := Bots["it"].Rdb.Set("update_statistic", strStatistic, 0).Result()
	if err != nil {
		log.Println(err)
	}
}
