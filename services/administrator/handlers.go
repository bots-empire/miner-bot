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
	h.OnCommand("/set_count", NewChangeMinerCountCommand())
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
	channel, _ := strconv.Atoi(strings.Split(s.Params.Level, "?")[2])
	lang := assets.AdminLang(s.User.ID)
	status := "operation_canceled"

	switch capitation {
	case "change_url":
		url, chatID := getUrlAndChatID(s.Message)
		if chatID == 0 {
			text := assets.AdminText(lang, "chat_id_not_update")
			return msgs.NewParseMessage(s.BotLang, s.User.ID, text)
		}
		assets.AdminSettings.UpdateAdvertChannelID(s.BotLang, chatID, channel)
		assets.AdminSettings.UpdateAdvertUrl(s.BotLang, channel, url)
		//assets.AdminSettings.UpdateAdvertChan(s.BotLang, advertChan)
	case "change_text":
		assets.AdminSettings.UpdateAdvertText(s.BotLang, s.Message.Text, channel)
	case "change_photo":
		if len(s.Message.Photo) == 0 {
			text := assets.AdminText(lang, "send_only_photo")
			return msgs.NewParseMessage(s.BotLang, s.User.ID, text)
		}
		assets.AdminSettings.UpdateAdvertPhoto(s.BotLang, channel, s.Message.Photo[0].FileID)
	case "change_video":
		if s.Message.Video == nil {
			text := assets.AdminText(lang, "send_only_video")
			return msgs.NewParseMessage(s.BotLang, s.User.ID, text)
		}
		assets.AdminSettings.UpdateAdvertVideo(s.BotLang, channel, s.Message.Video.FileID)
	}
	assets.SaveAdminSettings()
	status = "operation_completed"

	if err := setAdminBackButton(s.BotLang, s.User.ID, status); err != nil {
		return err
	}
	db.RdbSetUser(s.BotLang, s.User.ID, "admin")
	db.DeleteOldAdminMsg(s.BotLang, s.User.ID)

	callback := &tgbotapi.CallbackQuery{
		Data: "admin/change_advert_chan?" + strconv.Itoa(channel),
	}
	s.CallbackQuery = callback
	return NewAdvertisementChanMenuCommand().Serve(s)
}

type ChangeMinerCountCommand struct {
}

func NewChangeMinerCountCommand() *ChangeMinerCountCommand {
	return &ChangeMinerCountCommand{}
}

func (c *ChangeMinerCountCommand) Serve(s *model.Situation) error {
	level := db.RdbGetMinerLevelSetting(s.BotLang, s.User.ID)
	count := strings.Split(s.Params.Level, "?")[1]

	number, _ := strconv.Atoi(s.Message.Text)

	if number < 1 {
		return msgs.NewParseMessage(s.BotLang, s.User.ID, "need_positive_number")
	}

	switch count {
	case "hash":
		assets.AdminSettings.GetParams(s.BotLang).ClickAmount[level] = number
	case "price":
		assets.AdminSettings.GetParams(s.BotLang).UpgradeMinerCost[level] = number
	}

	err := msgs.NewParseMessage(s.BotLang, s.User.ID, assets.AdminText(assets.AdminLang(s.User.ID), "operation_completed"))
	if err != nil {
		return err
	}

	return sendMinerSettingMenu(s)
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

func getUrlAndChatID(message *tgbotapi.Message) (string, int64) {
	data := strings.Split(message.Text, "\n")
	if len(data) != 2 {
		return "", 0
	}

	chatId, err := strconv.Atoi(data[0])
	if err != nil {
		return "", 0
	}

	//advert := &assets.AdvertChannel{
	//	Url:       map[int]string{channel: data[1]},
	//	ChannelID: int64(chatId),
	//}

	//advert.Url[channel] = data[1]
	//advert.ChannelID = int64(chatId)

	return data[1], int64(chatId)
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
