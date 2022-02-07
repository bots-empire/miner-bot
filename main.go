package main

import (
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/Stepan1328/miner-bot/assets"
	"github.com/Stepan1328/miner-bot/log"
	"github.com/Stepan1328/miner-bot/model"
	"github.com/Stepan1328/miner-bot/msgs"
	"github.com/Stepan1328/miner-bot/services"
	"github.com/Stepan1328/miner-bot/services/administrator"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	rand.Seed(time.Now().Unix())

	logger := log.NewDefaultLogger().Prefix("Miner Bot")
	log.PrintLogo("Miner Bot", []string{"FFD700"})

	startServices(logger)
	startAllBot(logger)
	assets.UploadUpdateStatistic()

	go startPrometheusHandler(logger)

	startHandlers(logger)
}

func startAllBot(log log.Logger) {
	for lang, globalBot := range model.Bots {
		startBot(globalBot, log, lang)
		model.Bots[lang].MessageHandler = NewMessagesHandler()
		model.Bots[lang].CallbackHandler = NewCallbackHandler()
		model.Bots[lang].AdminMessageHandler = NewAdminMessagesHandler()
		model.Bots[lang].AdminCallBackHandler = NewAdminCallbackHandler()
	}

	log.Ok("All bots is running")
}

func startServices(log log.Logger) {
	model.FillBotsConfig()
	assets.ParseLangMap()
	assets.ParseAdminMap()
	assets.UploadAdminSettings()
	assets.ParseCommandsList()

	log.Ok("All services are running successfully")
}

func startBot(b *model.GlobalBot, log log.Logger, lang string) {
	var err error
	b.Bot, err = tgbotapi.NewBotAPI(b.BotToken)
	if err != nil {
		log.Fatal("error start bot: %s", err.Error())
	}

	u := tgbotapi.NewUpdate(0)

	b.Chanel = b.Bot.GetUpdatesChan(u)

	b.Rdb = model.StartRedis()
	b.DataBase = model.UploadDataBase(lang)
}

func startPrometheusHandler(logger log.Logger) {
	http.Handle("/metrics", promhttp.Handler())
	logger.Ok("Metrics can be read from %s port", "7011")
	metricErr := http.ListenAndServe(":7011", nil)
	if metricErr != nil {
		logger.Fatal("metrics stoped by metricErr: %s\n", metricErr.Error())
	}
}

func startHandlers(logger log.Logger) {
	wg := new(sync.WaitGroup)

	for botLang, handler := range model.Bots {
		wg.Add(1)
		go func(botLang string, handler *model.GlobalBot, wg *sync.WaitGroup) {
			defer wg.Done()
			services.ActionsWithUpdates(botLang, handler.Chanel, logger)
		}(botLang, handler, wg)
	}

	logger.Ok("All handlers are running")
	msgs.SendNotificationToDeveloper("All bots are restart")
	wg.Wait()
}

func NewMessagesHandler() *services.MessagesHandlers {
	handle := services.MessagesHandlers{
		Handlers: map[string]model.Handler{},
	}

	handle.Init()
	return &handle
}

func NewCallbackHandler() *services.CallBackHandlers {
	handle := services.CallBackHandlers{
		Handlers: map[string]model.Handler{},
	}

	handle.Init()
	return &handle
}

func NewAdminMessagesHandler() *administrator.AdminMessagesHandlers {
	handle := administrator.AdminMessagesHandlers{
		Handlers: map[string]model.Handler{},
	}

	handle.Init()
	return &handle
}

func NewAdminCallbackHandler() *administrator.AdminCallbackHandlers {
	handle := administrator.AdminCallbackHandlers{
		Handlers: map[string]model.Handler{},
	}

	handle.Init()
	return &handle
}
