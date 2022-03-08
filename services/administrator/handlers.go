package administrator

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Stepan1328/miner-bot/assets"
	"github.com/Stepan1328/miner-bot/db"
	"github.com/Stepan1328/miner-bot/model"
	"github.com/Stepan1328/miner-bot/msgs"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type AdminMessagesHandlers struct {
	Handlers map[string]model.Handler
}

func (h *AdminMessagesHandlers) GetHandler(command string) model.Handler {
	return h.Handlers[command]
}

func (h *AdminMessagesHandlers) Init() {
	h.OnCommand("/make_money", NewUpdateParameterCommand())
	h.OnCommand("/change_text_url", NewSetNewTextUrlCommand())
	h.OnCommand("/advertisement_setting", NewAdvertisementSettingCommand())
	h.OnCommand("/get_new_source", NewGetNewSourceCommand())
}

func (h *AdminMessagesHandlers) OnCommand(command string, handler model.Handler) {
	h.Handlers[command] = handler
}

type UpdateParameterCommand struct {
}

func NewUpdateParameterCommand() *UpdateParameterCommand {
	return &UpdateParameterCommand{}
}

func (c *UpdateParameterCommand) Serve(s *model.Situation) error {
	if strings.Contains(s.Params.Level, "make_money?") && s.Message.Text == "← Назад к ⚙️ Заработок" {
		if err := setAdminBackButton(s.BotLang, s.User.ID, "operation_canceled"); err != nil {
			return err
		}
		db.DeleteOldAdminMsg(s.BotLang, s.User.ID)
		s.Command = "admin/make_money_setting"

		return NewMakeMoneySettingCommand().Serve(s)
	}

	partitions := strings.Split(s.Params.Level, "?")
	if len(partitions) < 2 {
		return fmt.Errorf("smth went wrong")
	}

	partition := partitions[1]

	if partition == currencyType {
		assets.AdminSettings.GetParams(s.BotLang).Currency = s.Message.Text
	} else {
		err := setNewIntParameter(s, partition)
		if err != nil {
			return err
		}
	}

	assets.SaveAdminSettings()
	err := setAdminBackButton(s.BotLang, s.User.ID, "operation_completed")
	if err != nil {
		return nil
	}
	db.DeleteOldAdminMsg(s.BotLang, s.User.ID)
	s.Command = "admin/make_money_setting"

	return NewMakeMoneySettingCommand().Serve(s)
}

func setNewIntParameter(s *model.Situation, partition string) error {
	lang := assets.AdminLang(s.User.ID)

	newAmount, err := strconv.Atoi(s.Message.Text)
	if err != nil || newAmount <= 0 {
		text := assets.AdminText(lang, "incorrect_make_money_change_input")
		return msgs.NewParseMessage(s.BotLang, s.User.ID, text)
	}

	switch partition {
	case bonusAmount:
		assets.AdminSettings.GetParams(s.BotLang).BonusAmount = newAmount
	case minWithdrawalAmount:
		assets.AdminSettings.GetParams(s.BotLang).MinWithdrawalAmount = newAmount
	case maxOfClickPDAmount:
		assets.AdminSettings.GetParams(s.BotLang).MaxOfClickPerDay = newAmount
	case referralAmount:
		assets.AdminSettings.GetParams(s.BotLang).ReferralAmount = newAmount
	}

	return nil
}

type SetNewTextUrlCommand struct {
}

func NewSetNewTextUrlCommand() *SetNewTextUrlCommand {
	return &SetNewTextUrlCommand{}
}

func (c *SetNewTextUrlCommand) Serve(s *model.Situation) error {
	capitation := strings.Split(s.Params.Level, "?")[1]
	lang := assets.AdminLang(s.User.ID)
	status := "operation_canceled"

	switch capitation {
	case "change_url":
		advertChan := getUrlAndChatID(s.Message)
		if advertChan.ChannelID == 0 {
			text := assets.AdminText(lang, "chat_id_not_update")
			return msgs.NewParseMessage(s.BotLang, s.User.ID, text)
		}

		assets.AdminSettings.UpdateAdvertChan(s.BotLang, advertChan)
	case "change_text":
		assets.AdminSettings.UpdateAdvertText(s.BotLang, s.Message.Text)
	}
	assets.SaveAdminSettings()
	status = "operation_completed"

	if err := setAdminBackButton(s.BotLang, s.User.ID, status); err != nil {
		return err
	}
	db.RdbSetUser(s.BotLang, s.User.ID, "admin")
	db.DeleteOldAdminMsg(s.BotLang, s.User.ID)

	s.Command = "admin/advertisement"
	s.Params.Level = "admin/change_url"
	return NewAdvertisementMenuCommand().Serve(s)
}

type AdvertisementSettingCommand struct {
}

func NewAdvertisementSettingCommand() *AdvertisementSettingCommand {
	return &AdvertisementSettingCommand{}
}

func (c *AdvertisementSettingCommand) Serve(s *model.Situation) error {
	s.CallbackQuery = &tgbotapi.CallbackQuery{
		Data: "admin/change_text_url?",
	}
	s.Command = "admin/advertisement"
	return NewAdvertisementMenuCommand().Serve(s)
}

func getUrlAndChatID(message *tgbotapi.Message) *assets.AdvertChannel {
	data := strings.Split(message.Text, "\n")
	if len(data) != 2 {
		return &assets.AdvertChannel{}
	}

	chatId, err := strconv.Atoi(data[1])
	if err != nil {
		return &assets.AdvertChannel{}
	}

	return &assets.AdvertChannel{
		Url:       data[0],
		ChannelID: int64(chatId),
	}
}

func CheckAdminMessage(s *model.Situation) error {
	if !ContainsInAdmin(s.User.ID) {
		return notAdmin(s.BotLang, s.User)
	}

	s.Command, s.Err = assets.GetCommandFromText(s.Message, s.User.Language, s.User.ID)
	if s.Err == nil {
		Handler := model.Bots[s.BotLang].AdminMessageHandler.
			GetHandler(s.Command)

		if Handler != nil {
			return Handler.Serve(s)
		}
	}

	s.Command = strings.TrimLeft(strings.Split(s.Params.Level, "?")[0], "admin")

	Handler := model.Bots[s.BotLang].AdminMessageHandler.
		GetHandler(s.Command)

	if Handler != nil {
		return Handler.Serve(s)
	}

	return model.ErrCommandNotConverted
}

type StartTestMailing1Command struct {
}

func NewStartTestMailing1Command() *StartTestMailing1Command {
	return &StartTestMailing1Command{}
}

func (c *StartTestMailing1Command) Serve(s *model.Situation) error {
	go db.StartTestMailing1(s.BotLang, s.User)
	return msgs.NewParseMessage(s.BotLang, s.User.ID, "тестовая рассылка запущена")
}
