package services

import (
	"strconv"
	"strings"

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

type CallBackHandlers struct {
	Handlers map[string]model.Handler
}

func (h *CallBackHandlers) GetHandler(command string) model.Handler {
	return h.Handlers[command]
}

func (h *CallBackHandlers) Init() {
	// Start commands
	h.OnCommand("/language", NewLanguageCommand())

	// Money commands
	h.OnCommand("/make_money_click", NewHandleClickCommand())
	h.OnCommand("/upgrade_miner_lvl", NewUpgradeMinerLvlCommand())
	h.OnCommand("/send_bonus_to_user", NewGetBonusCommand())
	h.OnCommand("/withdrawal_money", NewRecheckSubscribeCommand())
	h.OnCommand("/promotion_case", NewPromotionCaseCommand())
}

func (h *CallBackHandlers) OnCommand(command string, handler model.Handler) {
	h.Handlers[command] = handler
}

func checkCallbackQuery(s *model.Situation, logger log.Logger, sortCentre *utils.Spreader) {
	if strings.Contains(s.Params.Level, "admin") {
		if err := administrator.CheckAdminCallback(s); err != nil {
			logger.Warn("error with serve admin callback command: %s", err.Error())
		}
		return
	}

	handler := model.Bots[s.BotLang].CallbackHandler.
		GetHandler(s.Command)

	if handler != nil {
		sortCentre.ServeHandler(handler, s, func(err error) {
			logger.Warn("error with serve user callback command: %s", err.Error())
			smthWentWrong(s.BotLang, s.CallbackQuery.Message.Chat.ID, s.User.Language)
		})

		return
	}

	logger.Warn("get callback data='%s', but they didn't react in any way", s.CallbackQuery.Data)
}

type LanguageCommand struct {
}

func NewLanguageCommand() *LanguageCommand {
	return &LanguageCommand{}
}

func (c *LanguageCommand) Serve(s *model.Situation) error {
	lang := strings.Split(s.CallbackQuery.Data, "?")[1]

	level := db.GetLevel(s.BotLang, s.User.ID)
	if strings.Contains(level, "admin") {
		return nil
	}

	s.User.Language = lang

	return NewStartCommand().Serve(s)
}

type HandleClickCommand struct {
}

func NewHandleClickCommand() *HandleClickCommand {
	return &HandleClickCommand{}
}

func (c *HandleClickCommand) Serve(s *model.Situation) error {
	err, ok := auth.MakeClick(s)
	if err != nil {
		return errors.Wrap(err, "failed make click")
	}
	if ok {
		return nil
	}
	_ = msgs.SendAnswerCallback(s.BotLang, s.CallbackQuery, s.User.Language, "click_done")

	s.User, err = auth.GetUser(s.BotLang, s.User.ID)
	if err != nil {
		return nil
	}
	text, markUp := buildClickMsg(s.BotLang, s.User)
	oldMsgID := db.GetUserClickerMsgID(s.BotLang, s.User.ID)

	return msgs.NewEditMarkUpMessage(s.BotLang, s.User.ID, oldMsgID, markUp, text)
}

type UpgradeMinerLvlCommand struct {
}

func NewUpgradeMinerLvlCommand() *UpgradeMinerLvlCommand {
	return &UpgradeMinerLvlCommand{}
}

func (c *UpgradeMinerLvlCommand) Serve(s *model.Situation) error {
	nilBalance, err := auth.UpgradeMinerLevel(s)
	if err == model.ErrMaxLevelAlreadyCompleted {
		return reachedMaxMinerLvl(s)
	}
	if err != nil {
		return err
	}

	if nilBalance {
		text := assets.LangText(s.User.Language, "failed_upgrade_miner")

		return msgs.NewParseMessage(s.BotLang, s.User.ID, text)
	}

	_ = msgs.SendMsgToUser(s.BotLang, tgbotapi.NewDeleteMessage(s.User.ID, s.CallbackQuery.Message.MessageID))

	text := assets.LangText(s.User.Language, "successful_upgrade_miner",
		s.User.MinerLevel,
		assets.AdminSettings.GetParams(s.BotLang).UpgradeMinerCost[s.User.MinerLevel-2])

	return msgs.NewParseMessage(s.BotLang, s.User.ID, text)
}

type GetBonusCommand struct {
}

func NewGetBonusCommand() *GetBonusCommand {
	return &GetBonusCommand{}
}

func (c *GetBonusCommand) Serve(s *model.Situation) error {
	return auth.GetABonus(s)
}

type RecheckSubscribeCommand struct {
}

func NewRecheckSubscribeCommand() *RecheckSubscribeCommand {
	return &RecheckSubscribeCommand{}
}

func (c *RecheckSubscribeCommand) Serve(s *model.Situation) error {
	amount := strings.Split(s.CallbackQuery.Data, "?")[1]
	s.Message = &tgbotapi.Message{
		Text: amount,
	}
	if err := msgs.SendAnswerCallback(s.BotLang, s.CallbackQuery, s.User.Language, "invitation_to_subscribe"); err != nil {
		return err
	}
	amountInt, _ := strconv.Atoi(amount)

	if auth.CheckSubscribeToWithdrawal(s, amountInt) {
		db.RdbSetUser(s.BotLang, s.User.ID, "main")

		return NewStartCommand().Serve(s)
	}
	return nil
}

type PromotionCaseCommand struct {
}

func NewPromotionCaseCommand() *PromotionCaseCommand {
	return &PromotionCaseCommand{}
}

func (c *PromotionCaseCommand) Serve(s *model.Situation) error {
	cost, err := strconv.Atoi(strings.Split(s.CallbackQuery.Data, "?")[1])
	if err != nil {
		return err
	}

	if s.User.Balance < cost {
		return msgs.SendAnswerCallback(s.BotLang, s.CallbackQuery, s.User.Language, "not_enough_money")
	}

	db.RdbSetUser(s.BotLang, s.User.ID, s.CallbackQuery.Data)
	msg := tgbotapi.NewMessage(s.User.ID, assets.LangText(s.User.Language, "invitation_to_send_link_text"))
	msg.ReplyMarkup = msgs.NewMarkUp(
		msgs.NewRow(msgs.NewDataButton("withdraw_cancel")),
	).Build(s.User.Language)

	if err := msgs.SendAnswerCallback(s.BotLang, s.CallbackQuery, s.User.Language, "invitation_to_send_link"); err != nil {
		return err
	}

	return msgs.SendMsgToUser(s.BotLang, msg)
}
