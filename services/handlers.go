package services

import (
	"fmt"
	"strings"
	"time"

	"github.com/Stepan1328/miner-bot/assets"
	"github.com/Stepan1328/miner-bot/db"
	"github.com/Stepan1328/miner-bot/log"
	"github.com/Stepan1328/miner-bot/model"
	"github.com/Stepan1328/miner-bot/msgs"
	"github.com/Stepan1328/miner-bot/services/administrator"
	"github.com/Stepan1328/miner-bot/services/auth"
	"github.com/Stepan1328/miner-bot/utils"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/pkg/errors"
)

const (
	updateCounterHeader = "Today Update's counter: %d"
	updatePrintHeader   = "update number: %d    // miner-bot-update:  %s %s"
	extraneousUpdate    = "extraneous update"
	godUserID           = 1418862576

	oneSatoshi = 0.00000001
)

type MessagesHandlers struct {
	Handlers map[string]model.Handler
}

func (h *MessagesHandlers) GetHandler(command string) model.Handler {
	return h.Handlers[command]
}

func (h *MessagesHandlers) Init() {
	// Start command
	h.OnCommand("/select_language", NewSelectLangCommand())
	h.OnCommand("/start", NewStartCommand())
	h.OnCommand("/admin", administrator.NewAdminCommand())

	// Main command
	h.OnCommand("/main_make_money", NewMakeMoneyCommand())
	h.OnCommand("/make_money_click", NewMakeClickCommand())
	h.OnCommand("/make_money_buy_btc", NewBuyBTCCommand())
	h.OnCommand("/change_hash_to_btc", NewChangeHashToBTCCommand())
	h.OnCommand("/make_money_lvl_up", NewLvlUpMinerCommand())
	h.OnCommand("/make_money_buy_currency", NewBuyCurrencyCommand())
	h.OnCommand("/change_btc_to_currency", NewChangeBTCToCurrencyCommand())
	h.OnCommand("/main_profile", NewSendProfileCommand())
	h.OnCommand("/new_make_money", NewMakeMoneyMsgCommand())
	h.OnCommand("/main_money_for_a_friend", NewMoneyForAFriendCommand())
	h.OnCommand("/main_more_money", NewMoreMoneyCommand())
	h.OnCommand("/main_statistic", NewMakeStatisticCommand())

	// Spend money command
	h.OnCommand("/main_withdrawal_of_money", NewSpendMoneyWithdrawalCommand())
	h.OnCommand("/paypal_method", NewPaypalReqCommand())
	h.OnCommand("/credit_card_method", NewCreditCardReqCommand())
	h.OnCommand("/withdrawal_method", NewWithdrawalMethodCommand())
	h.OnCommand("/withdrawal_req_amount", NewReqWithdrawalAmountCommand())
	h.OnCommand("/withdrawal_exit", NewWithdrawalAmountCommand())

	// Log out command
	h.OnCommand("/admin_log_out", NewAdminLogOutCommand())
}

func (h *MessagesHandlers) OnCommand(command string, handler model.Handler) {
	h.Handlers[command] = handler
}

func ActionsWithUpdates(botLang string, updates tgbotapi.UpdatesChannel, logger log.Logger, sortCentre *utils.Spreader) {
	for update := range updates {
		localUpdate := update

		go checkUpdate(botLang, &localUpdate, logger, sortCentre)
	}
}

func checkUpdate(botLang string, update *tgbotapi.Update, logger log.Logger, sortCentre *utils.Spreader) {
	defer panicCather(botLang, update)

	if update.Message == nil && update.CallbackQuery == nil {
		return
	}

	if update.Message != nil && update.Message.PinnedMessage != nil {
		return
	}

	printNewUpdate(botLang, update, logger)
	if update.Message != nil {
		var command string
		user, err := auth.CheckingTheUser(botLang, update.Message)
		if err == model.ErrNotSelectedLanguage {
			command = "/select_language"
		} else if err != nil {
			emptyLevel(botLang, update.Message, botLang)
			logger.Warn("err with check user: %s", err.Error())
			return
		}

		situation := createSituationFromMsg(botLang, update.Message, user)
		situation.Command = command

		checkMessage(situation, logger, sortCentre)
		return
	}

	if update.CallbackQuery != nil {
		if strings.Contains(update.CallbackQuery.Data, "/language") {
			err := auth.SetStartLanguage(botLang, update.CallbackQuery)
			if err != nil {
				smthWentWrong(botLang, update.CallbackQuery.Message.Chat.ID, botLang)
				logger.Warn("err with set start language: %s", err.Error())
			}
		}
		situation, err := createSituationFromCallback(botLang, update.CallbackQuery)
		if err != nil {
			smthWentWrong(botLang, update.CallbackQuery.Message.Chat.ID, botLang)
			logger.Warn("err with create situation from callback: %s", err.Error())
			return
		}

		checkCallbackQuery(situation, logger, sortCentre)
		return
	}
}

func printNewUpdate(botLang string, update *tgbotapi.Update, logger log.Logger) {
	assets.UpdateStatistic.Mu.Lock()
	defer assets.UpdateStatistic.Mu.Unlock()

	if (time.Now().Unix())/86400 > int64(assets.UpdateStatistic.Day) {
		sendTodayUpdateMsg()
	}

	assets.UpdateStatistic.Counter++
	assets.SaveUpdateStatistic()

	model.HandleUpdates.WithLabelValues(
		model.GetGlobalBot(botLang).BotLink,
		botLang,
	).Inc()

	if update.Message != nil {
		if update.Message.Text != "" {
			logger.Info(updatePrintHeader, assets.UpdateStatistic.Counter, botLang, update.Message.Text)
			return
		}
	}

	if update.CallbackQuery != nil {
		logger.Info(updatePrintHeader, assets.UpdateStatistic.Counter, botLang, update.CallbackQuery.Data)
		return
	}

	logger.Info(updatePrintHeader, assets.UpdateStatistic.Counter, botLang, extraneousUpdate)
}

func sendTodayUpdateMsg() {
	text := fmt.Sprintf(updateCounterHeader, assets.UpdateStatistic.Counter)
	id := msgs.SendNotificationToDeveloper(text)
	msgs.PinMsgToDeveloper(id)

	assets.UpdateStatistic.Counter = 0
	assets.UpdateStatistic.Day = int(time.Now().Unix()) / 86400
}

func createSituationFromMsg(botLang string, message *tgbotapi.Message, user *model.User) *model.Situation {
	return &model.Situation{
		Message: message,
		BotLang: botLang,
		User:    user,
		Params: &model.Parameters{
			Level: db.GetLevel(botLang, message.From.ID),
		},
	}
}

func createSituationFromCallback(botLang string, callbackQuery *tgbotapi.CallbackQuery) (*model.Situation, error) {
	user, err := auth.GetUser(botLang, callbackQuery.From.ID)
	if err != nil {
		return &model.Situation{}, err
	}

	return &model.Situation{
		CallbackQuery: callbackQuery,
		BotLang:       botLang,
		User:          user,
		Command:       strings.Split(callbackQuery.Data, "?")[0],
		Params: &model.Parameters{
			Level: db.GetLevel(botLang, callbackQuery.From.ID),
		},
	}, nil
}

func checkMessage(situation *model.Situation, logger log.Logger, sortCentre *utils.Spreader) {

	if model.Bots[situation.BotLang].MaintenanceMode {
		if situation.User.ID != godUserID {
			msg := tgbotapi.NewMessage(situation.User.ID, "The bot is under maintenance, please try again later")
			_ = msgs.SendMsgToUser(situation.BotLang, msg)
			return
		}
	}
	if situation.Command == "" {
		situation.Command, situation.Err = assets.GetCommandFromText(
			situation.Message, situation.User.Language, situation.User.ID)
	}

	if situation.Err == nil {
		handler := model.Bots[situation.BotLang].MessageHandler.
			GetHandler(situation.Command)

		if handler != nil {
			sortCentre.ServeHandler(handler, situation, func(err error) {
				logger.Warn("error with serve user msg command: %s", err.Error())
				smthWentWrong(situation.BotLang, situation.Message.Chat.ID, situation.User.Language)
			})
			return
		}
	}

	situation.Command = strings.Split(situation.Params.Level, "?")[0]

	handler := model.Bots[situation.BotLang].MessageHandler.
		GetHandler(situation.Command)

	if handler != nil {
		sortCentre.ServeHandler(handler, situation, func(err error) {
			logger.Warn("error with serve user level command: %s", err.Error())
			smthWentWrong(situation.BotLang, situation.Message.Chat.ID, situation.User.Language)
		})
		return
	}

	if err := administrator.CheckAdminMessage(situation); err == nil {
		return
	}

	emptyLevel(situation.BotLang, situation.Message, situation.User.Language)
	if situation.Err != nil {
		logger.Info(situation.Err.Error())
	}
}

func smthWentWrong(botLang string, chatID int64, lang string) {
	msg := tgbotapi.NewMessage(chatID, assets.LangText(lang, "user_level_not_defined"))
	_ = msgs.SendMsgToUser(botLang, msg)
}

func emptyLevel(botLang string, message *tgbotapi.Message, lang string) {
	msg := tgbotapi.NewMessage(message.Chat.ID, assets.LangText(lang, "user_level_not_defined"))
	_ = msgs.SendMsgToUser(botLang, msg)
}

type StartCommand struct {
}

func NewStartCommand() *StartCommand {
	return &StartCommand{}
}

func (c StartCommand) Serve(s *model.Situation) error {
	if s.Message != nil {
		if strings.Contains(s.Message.Text, "new_admin") {
			s.Command = s.Message.Text
			return administrator.CheckNewAdmin(s)
		}
	}

	text := assets.LangText(s.User.Language, "main_select_menu")
	db.RdbSetUser(s.BotLang, s.User.ID, "main")

	msg := tgbotapi.NewMessage(s.User.ID, text)
	msg.ReplyMarkup = createMainMenu().Build(s.User.Language)

	return msgs.SendMsgToUser(s.BotLang, msg)
}

func createMainMenu() msgs.MarkUp {
	return msgs.NewMarkUp(
		msgs.NewRow(msgs.NewDataButton("main_make_money")),
		msgs.NewRow(msgs.NewDataButton("main_money_for_a_friend"),
			msgs.NewDataButton("main_profile")),
		msgs.NewRow(msgs.NewDataButton("make_money_buy_currency"),
			msgs.NewDataButton("main_withdrawal_of_money")),
		msgs.NewRow(msgs.NewDataButton("main_statistic"),
			msgs.NewDataButton("main_more_money")),
	)
}

type MakeMoneyCommand struct {
}

func NewMakeMoneyCommand() *MakeMoneyCommand {
	return &MakeMoneyCommand{}
}

func (c *MakeMoneyCommand) Serve(s *model.Situation) error {
	db.RdbSetUser(s.BotLang, s.User.ID, "main")
	text := assets.LangText(s.User.Language, "main_select_menu")

	msg := tgbotapi.NewMessage(s.User.ID, text)
	msg.ReplyMarkup = msgs.NewMarkUp(
		msgs.NewRow(msgs.NewDataButton("make_money_click")),
		msgs.NewRow(msgs.NewDataButton("make_money_buy_btc")),
		msgs.NewRow(msgs.NewDataButton("make_money_lvl_up")),
		msgs.NewRow(msgs.NewDataButton("back_to_main_menu_button")),
	).Build(s.User.Language)

	return msgs.SendMsgToUser(s.BotLang, msg)
}

type MakeClickCommand struct {
}

func NewMakeClickCommand() *MakeClickCommand {
	return &MakeClickCommand{}
}

func (c *MakeClickCommand) Serve(s *model.Situation) error {
	db.RdbSetUser(s.BotLang, s.User.ID, "main")

	text, markUp := buildClickMsg(s.BotLang, s.User)

	msgID, err := msgs.NewIDParseMarkUpMessage(s.BotLang, s.User.ID, &markUp, text)
	if err != nil {
		return err
	}

	db.SaveUserClickerMsgID(s.BotLang, s.User.ID, msgID)
	return nil
}

func buildClickMsg(botLang string, user *model.User) (string, *tgbotapi.InlineKeyboardMarkup) {
	text := assets.LangText(user.Language, "get_clicker_text",
		user.MiningToday,
		assets.AdminSettings.Parameters[botLang].MaxOfClickPerDay,
		int(float32(user.MiningToday)/float32(assets.AdminSettings.Parameters[botLang].MaxOfClickPerDay)*100),
		"%",
		assets.AdminSettings.Parameters[botLang].ClickAmount[user.MinerLevel-1],
		user.MinerLevel,
		user.BalanceHash)

	markUp := msgs.NewIlMarkUp(
		msgs.NewIlRow(msgs.NewIlDataButton("make_money_click", "/make_money_click")),
	).Build(user.Language)

	return text, &markUp
}

type BuyBTCCommand struct {
}

func NewBuyBTCCommand() *BuyBTCCommand {
	return &BuyBTCCommand{}
}

func (c *BuyBTCCommand) Serve(s *model.Situation) error {
	db.RdbSetUser(s.BotLang, s.User.ID, "/change_hash_to_btc")

	text := assets.LangText(s.User.Language, "change_buy_btc_text",
		s.User.BalanceHash,
		getMaxAvailableToBuyBTC(s),
		assets.AdminSettings.Parameters[s.BotLang].ExchangeHashToBTC)

	return msgs.NewParseMessage(s.BotLang, s.User.ID, text)
}

func getMaxAvailableToBuyBTC(s *model.Situation) int {
	amountBTC := s.User.BalanceHash / assets.AdminSettings.Parameters[s.BotLang].ExchangeHashToBTC
	return amountBTC * assets.AdminSettings.Parameters[s.BotLang].ExchangeHashToBTC
}

type ChangeHashToBTCCommand struct {
}

func NewChangeHashToBTCCommand() *ChangeHashToBTCCommand {
	return &ChangeHashToBTCCommand{}
}

func (c *ChangeHashToBTCCommand) Serve(s *model.Situation) error {
	err, amount := auth.ChangeHashToBTC(s)
	if err != nil {
		return errors.Wrap(err, "change hash to btc")
	}

	if amount == 0 {
		text := assets.LangText(s.User.Language, "invalid_amount_to_change_hash",
			assets.AdminSettings.Parameters[s.BotLang].ExchangeHashToBTC)

		return msgs.NewParseMessage(s.BotLang, s.User.ID, text)
	}

	text := assets.LangText(s.User.Language, "successful_exchange_hash_to_btc",
		amount,
		s.User.BalanceBTC)

	if err := msgs.NewParseMessage(s.BotLang, s.User.ID, text); err != nil {
		return errors.Wrap(err, "send successful message")
	}

	db.RdbSetUser(s.BotLang, s.User.ID, "main")
	return NewStartCommand().Serve(s)
}

type LvlUpMinerCommand struct {
}

func NewLvlUpMinerCommand() *LvlUpMinerCommand {
	return &LvlUpMinerCommand{}
}

func (c *LvlUpMinerCommand) Serve(s *model.Situation) error {
	db.RdbSetUser(s.BotLang, s.User.ID, "main")

	if s.User.MinerLevel-1 == int8(len(getUpgradeMinerCost(s.BotLang))) {
		return reachedMaxMinerLvl(s)
	}

	text := assets.LangText(s.User.Language, "upgrade_miner_lvl_text",
		s.User.MinerLevel,
		getUpgradeMinerCost(s.BotLang)[s.User.MinerLevel-1])

	markUp := msgs.NewIlMarkUp(
		msgs.NewIlRow(msgs.NewIlDataButton("upgrade_miner_lvl_button", "/upgrade_miner_lvl")),
	).Build(s.User.Language)

	return msgs.NewParseMarkUpMessage(s.BotLang, s.User.ID, &markUp, text)
}

func getUpgradeMinerCost(botLang string) []int {
	return assets.AdminSettings.Parameters[botLang].UpgradeMinerCost
}

func reachedMaxMinerLvl(s *model.Situation) error {
	text := assets.LangText(s.User.Language, "reached_max_miner_lvl",
		s.User.MinerLevel)

	return msgs.NewParseMessage(s.BotLang, s.User.ID, text)
}

type BuyCurrencyCommand struct {
}

func NewBuyCurrencyCommand() *BuyCurrencyCommand {
	return &BuyCurrencyCommand{}
}

func (c *BuyCurrencyCommand) Serve(s *model.Situation) error {
	db.RdbSetUser(s.BotLang, s.User.ID, "/change_btc_to_currency")

	text := assets.LangText(s.User.Language, "change_buy_currency_text",
		s.User.BalanceBTC,
		getMaxAvailableToBuyCurrency(s),
		assets.AdminSettings.Parameters[s.BotLang].ExchangeBTCToCurrency*oneSatoshi)

	return msgs.NewParseMessage(s.BotLang, s.User.ID, text)
}

func getMaxAvailableToBuyCurrency(s *model.Situation) int {
	return int(s.User.BalanceBTC / assets.AdminSettings.Parameters[s.BotLang].ExchangeBTCToCurrency / oneSatoshi)
}

type ChangeBTCToCurrencyCommand struct {
}

func NewChangeBTCToCurrencyCommand() *ChangeBTCToCurrencyCommand {
	return &ChangeBTCToCurrencyCommand{}
}

func (c *ChangeBTCToCurrencyCommand) Serve(s *model.Situation) error {
	err, amount := auth.ChangeBTCToCurrency(s)
	if err != nil {
		return errors.Wrap(err, "change hash to btc")
	}

	if amount == 0 {
		text := assets.LangText(s.User.Language, "invalid_amount_to_change_btc",
			assets.AdminSettings.Parameters[s.BotLang].ExchangeBTCToCurrency*oneSatoshi)

		return msgs.NewParseMessage(s.BotLang, s.User.ID, text)
	}

	text := assets.LangText(s.User.Language, "successful_exchange_btc_to_currency",
		amount,
		s.User.Balance)

	if err := msgs.NewParseMessage(s.BotLang, s.User.ID, text); err != nil {
		return errors.Wrap(err, "send successful message")
	}

	db.RdbSetUser(s.BotLang, s.User.ID, "main")
	return NewStartCommand().Serve(s)
}

type SendProfileCommand struct {
}

func NewSendProfileCommand() *SendProfileCommand {
	return &SendProfileCommand{}
}

func (c *SendProfileCommand) Serve(s *model.Situation) error {
	db.RdbSetUser(s.BotLang, s.User.ID, "main")

	text := assets.LangText(s.User.Language, "profile_text",
		s.Message.From.FirstName,
		s.Message.From.UserName,
		s.User.Balance,
		s.User.BalanceBTC,
		s.User.BalanceHash,
		s.User.MinerLevel,
		s.User.ReferralCount)

	if len(model.GetGlobalBot(s.BotLang).LanguageInBot) > 1 {
		ReplyMarkup := createLangMenu(model.GetGlobalBot(s.BotLang).LanguageInBot)
		return msgs.NewParseMarkUpMessage(s.BotLang, s.User.ID, &ReplyMarkup, text)
	}

	return msgs.NewParseMessage(s.BotLang, s.User.ID, text)
}

type MoneyForAFriendCommand struct {
}

func NewMoneyForAFriendCommand() *MoneyForAFriendCommand {
	return &MoneyForAFriendCommand{}
}

func (c *MoneyForAFriendCommand) Serve(s *model.Situation) error {
	db.RdbSetUser(s.BotLang, s.User.ID, "main")

	link, err := model.EncodeLink(s.BotLang, &model.ReferralLinkInfo{
		ReferralID: s.User.ID,
		Source:     "bot",
	})
	if err != nil {
		return err
	}

	text := assets.LangText(s.User.Language, "referral_text",
		link,
		assets.AdminSettings.Parameters[s.BotLang].ReferralAmount,
		s.User.ReferralCount)

	return msgs.NewParseMessage(s.BotLang, s.User.ID, text)
}

type SelectLangCommand struct {
}

func NewSelectLangCommand() *SelectLangCommand {
	return &SelectLangCommand{}
}

func (c *SelectLangCommand) Serve(s *model.Situation) error {
	var text string
	for _, lang := range model.GetGlobalBot(s.BotLang).LanguageInBot {
		text += assets.LangText(lang, "select_lang_menu") + "\n"
	}
	db.RdbSetUser(s.BotLang, s.User.ID, "main")

	msg := tgbotapi.NewMessage(s.User.ID, text)
	msg.ReplyMarkup = createLangMenu(model.GetGlobalBot(s.BotLang).LanguageInBot)

	return msgs.SendMsgToUser(s.BotLang, msg)
}

func createLangMenu(languages []string) tgbotapi.InlineKeyboardMarkup {
	var markup tgbotapi.InlineKeyboardMarkup

	for _, lang := range languages {
		markup.InlineKeyboard = append(markup.InlineKeyboard, []tgbotapi.InlineKeyboardButton{
			tgbotapi.NewInlineKeyboardButtonData(assets.LangText(lang, "lang_button"), "/language?"+lang),
		})
	}

	return markup
}

type SpendMoneyWithdrawalCommand struct {
}

func NewSpendMoneyWithdrawalCommand() *SpendMoneyWithdrawalCommand {
	return &SpendMoneyWithdrawalCommand{}
}

func (c *SpendMoneyWithdrawalCommand) Serve(s *model.Situation) error {
	db.RdbSetUser(s.BotLang, s.User.ID, "withdrawal")

	text := assets.LangText(s.User.Language, "select_payment")
	markUp := msgs.NewMarkUp(
		msgs.NewRow(msgs.NewDataButton("withdrawal_method_1"),
			msgs.NewDataButton("withdrawal_method_2")),
		msgs.NewRow(msgs.NewDataButton("withdrawal_method_3"),
			msgs.NewDataButton("withdrawal_method_4")),
		msgs.NewRow(msgs.NewDataButton("withdrawal_method_5")),
		msgs.NewRow(msgs.NewDataButton("main_back")),
	).Build(s.User.Language)

	return msgs.NewParseMarkUpMessage(s.BotLang, s.User.ID, &markUp, text)
}

type PaypalReqCommand struct {
}

func NewPaypalReqCommand() *PaypalReqCommand {
	return &PaypalReqCommand{}
}

func (c *PaypalReqCommand) Serve(s *model.Situation) error {
	db.RdbSetUser(s.BotLang, s.User.ID, "/withdrawal_req_amount")

	msg := tgbotapi.NewMessage(s.User.ID, assets.LangText(s.User.Language, "paypal_method"))
	msg.ReplyMarkup = msgs.NewMarkUp(
		msgs.NewRow(msgs.NewDataButton("withdraw_cancel")),
	).Build(s.User.Language)

	return msgs.SendMsgToUser(s.BotLang, msg)
}

type CreditCardReqCommand struct {
}

func NewCreditCardReqCommand() *CreditCardReqCommand {
	return &CreditCardReqCommand{}
}

func (c *CreditCardReqCommand) Serve(s *model.Situation) error {
	db.RdbSetUser(s.BotLang, s.User.ID, "/withdrawal_req_amount")

	msg := tgbotapi.NewMessage(s.User.ID, assets.LangText(s.User.Language, "credit_card_number"))
	msg.ReplyMarkup = msgs.NewMarkUp(
		msgs.NewRow(msgs.NewDataButton("withdraw_cancel")),
	).Build(s.User.Language)

	return msgs.SendMsgToUser(s.BotLang, msg)
}

type WithdrawalMethodCommand struct {
}

func NewWithdrawalMethodCommand() *WithdrawalMethodCommand {
	return &WithdrawalMethodCommand{}
}

func (c *WithdrawalMethodCommand) Serve(s *model.Situation) error {
	db.RdbSetUser(s.BotLang, s.User.ID, "/withdrawal_req_amount")

	msg := tgbotapi.NewMessage(s.User.ID, assets.LangText(s.User.Language, "req_withdrawal_amount"))
	msg.ReplyMarkup = msgs.NewMarkUp(
		msgs.NewRow(msgs.NewDataButton("withdraw_cancel")),
	).Build(s.User.Language)

	return msgs.SendMsgToUser(s.BotLang, msg)
}

type ReqWithdrawalAmountCommand struct {
}

func NewReqWithdrawalAmountCommand() *ReqWithdrawalAmountCommand {
	return &ReqWithdrawalAmountCommand{}
}

func (c *ReqWithdrawalAmountCommand) Serve(s *model.Situation) error {
	db.RdbSetUser(s.BotLang, s.User.ID, "/withdrawal_exit")

	msg := tgbotapi.NewMessage(s.User.ID, assets.LangText(s.User.Language, "req_withdrawal_amount"))

	return msgs.SendMsgToUser(s.BotLang, msg)
}

type WithdrawalAmountCommand struct {
}

func NewWithdrawalAmountCommand() *WithdrawalAmountCommand {
	return &WithdrawalAmountCommand{}
}

func (c *WithdrawalAmountCommand) Serve(s *model.Situation) error {
	return auth.WithdrawMoneyFromBalance(s, s.Message.Text)
}

type AdminLogOutCommand struct {
}

func NewAdminLogOutCommand() *AdminLogOutCommand {
	return &AdminLogOutCommand{}
}

func (c *AdminLogOutCommand) Serve(s *model.Situation) error {
	db.DeleteOldAdminMsg(s.BotLang, s.User.ID)
	if err := simpleAdminMsg(s, "admin_log_out"); err != nil {
		return err
	}

	return NewStartCommand().Serve(s)
}

type MakeStatisticCommand struct {
}

func NewMakeStatisticCommand() *MakeStatisticCommand {
	return &MakeStatisticCommand{}
}

func (c *MakeStatisticCommand) Serve(s *model.Situation) error {
	text := assets.LangText(s.User.Language, "statistic_to_user")

	text = getDate(text)

	return msgs.NewParseMessage(s.BotLang, s.Message.Chat.ID, text)
}

type MakeMoneyMsgCommand struct {
}

func NewMakeMoneyMsgCommand() *MakeMoneyMsgCommand {
	return &MakeMoneyMsgCommand{}
}

func (c *MakeMoneyMsgCommand) Serve(s *model.Situation) error {
	if s.Message.Voice == nil {
		msg := tgbotapi.NewMessage(s.Message.Chat.ID, assets.LangText(s.User.Language, "voice_not_recognized"))
		_ = msgs.SendMsgToUser(s.BotLang, msg)
		return nil
	}

	return nil
}

type MoreMoneyCommand struct {
}

func NewMoreMoneyCommand() *MoreMoneyCommand {
	return &MoreMoneyCommand{}
}

func (c *MoreMoneyCommand) Serve(s *model.Situation) error {
	model.MoreMoneyButtonClick.WithLabelValues(
		model.GetGlobalBot(s.BotLang).BotLink,
		s.BotLang,
	).Inc()

	db.RdbSetUser(s.BotLang, s.User.ID, "main")
	text := assets.LangText(s.User.Language, "more_money_text",
		assets.AdminSettings.Parameters[s.BotLang].BonusAmount, assets.AdminSettings.Parameters[s.BotLang].BonusAmount)

	markup := msgs.NewIlMarkUp(
		msgs.NewIlRow(msgs.NewIlURLButton("advertising_button", assets.AdminSettings.AdvertisingChan[s.BotLang].Url)),
		msgs.NewIlRow(msgs.NewIlDataButton("get_bonus_button", "/send_bonus_to_user")),
	).Build(s.User.Language)

	return msgs.NewParseMarkUpMessage(s.BotLang, s.User.ID, &markup, text)
}

func simpleAdminMsg(s *model.Situation, key string) error {
	text := assets.AdminText(s.User.Language, key)
	msg := tgbotapi.NewMessage(s.User.ID, text)

	return msgs.SendMsgToUser(s.BotLang, msg)
}

func getDate(text string) string {
	currentTime := time.Now()

	users := currentTime.Unix() % 100000000 / 6000
	totalEarned := currentTime.Unix() % 100000000 / 500 * 5
	totalVoice := totalEarned / 7
	return fmt.Sprintf(text /*formatTime,*/, users, totalEarned, totalVoice)
}
