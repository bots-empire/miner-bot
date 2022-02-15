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
)

type Admin struct {
	AdminID         map[int64]*AdminUser
	Parameters      map[string]*Params
	AdvertisingChan map[string]*AdvertChannel
	BlockedUsers    map[string]int
	LangSelectedMap map[string]bool
	AdvertisingText map[string]string
}

type AdminUser struct {
	Language           string
	FirstName          string
	SpecialPossibility bool
}

type Params struct { // TODO: rename fields
	BonusAmount         int
	MinWithdrawalAmount int
	VoiceAmount         int
	MaxOfVoicePerDay    int
	ReferralAmount      int

	Currency string
}

type AdvertChannel struct {
	Url       string
	ChannelID int64
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
		nilSettings(settings, lang)
		for _, lang = range globalBot.LanguageInBot {
			nilSettings(settings, lang)
		}
	}

	AdminSettings = settings
	SaveAdminSettings()
}

func nilSettings(settings *Admin, lang string) {
	if settings.Parameters[lang] == nil {
		settings.Parameters[lang] = &Params{}
	}
	if settings.AdvertisingChan[lang] == nil {
		settings.AdvertisingChan[lang] = &AdvertChannel{
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
