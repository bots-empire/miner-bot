package main

import (
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/roylee0704/gron"
	"github.com/roylee0704/gron/xtime"

	"github.com/Stepan1328/miner-bot/log"
	"github.com/Stepan1328/miner-bot/model"
	"github.com/Stepan1328/miner-bot/services"
	"github.com/Stepan1328/miner-bot/services/administrator"
	"github.com/Stepan1328/miner-bot/services/auth"
	"github.com/Stepan1328/miner-bot/utils"
	"github.com/bots-empire/base-bot/mailing"
	"github.com/bots-empire/base-bot/msgs"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	rand.Seed(time.Now().Unix())

	logger := log.NewDefaultLogger().Prefix("Miner Bot")
	log.PrintLogo("Miner Bot", []string{"FFD700"})

	model.FillBotsConfig()
	model.UploadAdminSettings()
	go startPrometheusHandler(logger)

	srvs := startAllBot(logger)
	model.UploadUpdateStatistic()

	startHandlers(srvs, logger)
}

func startAllBot(log log.Logger) []*services.Users {
	srvs := make([]*services.Users, 0)

	for lang, globalBot := range model.Bots {
		startBot(globalBot, log, lang)

		service := msgs.NewService(globalBot, []int64{872383555, 1418862576, -1001736803459})

		authSrv := auth.NewAuthService(globalBot, service)
		mail := mailing.NewService(service, 100)
		adminSrv := administrator.NewAdminService(globalBot, mail, service)
		userSrv := services.NewUsersService(globalBot, authSrv, adminSrv, service)

		globalBot.MessageHandler = NewMessagesHandler(userSrv, adminSrv)
		globalBot.CallbackHandler = NewCallbackHandler(userSrv)
		globalBot.AdminMessageHandler = NewAdminMessagesHandler(adminSrv)
		globalBot.AdminCallBackHandler = NewAdminCallbackHandler(adminSrv)

		srvs = append(srvs, userSrv)
	}

	log.Ok("All bots is running")
	return srvs
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

	b.ParseLangMap()
	b.ParseCommandsList()
	b.ParseAdminMap()
}

func startPrometheusHandler(logger log.Logger) {
	http.Handle("/metrics", promhttp.Handler())
	logger.Ok("Metrics can be read from %s port", "7011")
	metricErr := http.ListenAndServe(":7011", nil)
	if metricErr != nil {
		logger.Fatal("metrics stoped by metricErr: %s\n", metricErr.Error())
	}
}

func startHandlers(srvs []*services.Users, logger log.Logger) {
	wg := new(sync.WaitGroup)
	cron := gron.New()
	cron.AddFunc(gron.Every(1*xtime.Day).At("20:59"), srvs[0].SendTodayUpdateMsg)

	for _, service := range srvs {
		wg.Add(1)
		go func(handler *services.Users, wg *sync.WaitGroup, cron *gron.Cron) {
			defer wg.Done()
			handler.ActionsWithUpdates(logger, utils.NewSpreader(time.Minute), cron)
		}(service, wg, cron)

		service.Msgs.SendNotificationToDeveloper("Bot are restart", false)
	}

	go func() {
		time.Sleep(5 * time.Second)

		cron.Start()
	}()

	logger.Ok("All handlers are running")

	wg.Wait()
}

func NewMessagesHandler(userSrv *services.Users, adminSrv *administrator.Admin) *services.MessagesHandlers {
	handle := services.MessagesHandlers{
		Handlers: map[string]model.Handler{},
	}

	handle.Init(userSrv, adminSrv)
	return &handle
}

func NewCallbackHandler(userSrv *services.Users) *services.CallBackHandlers {
	handle := services.CallBackHandlers{
		Handlers: map[string]model.Handler{},
	}

	handle.Init(userSrv)
	return &handle
}

func NewAdminMessagesHandler(adminSrv *administrator.Admin) *administrator.AdminMessagesHandlers {
	handle := administrator.AdminMessagesHandlers{
		Handlers: map[string]model.Handler{},
	}

	handle.Init(adminSrv)
	return &handle
}

func NewAdminCallbackHandler(adminSrv *administrator.Admin) *administrator.AdminCallbackHandlers {
	handle := administrator.AdminCallbackHandlers{
		Handlers: map[string]model.Handler{},
	}

	handle.Init(adminSrv)
	return &handle
}
