package msgs

import (
	"github.com/Stepan1328/miner-bot/assets"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

/*
==================================================
		MarkUp
==================================================
*/

type MarkUp struct {
	Rows []Row
}

func NewMarkUp(rows ...Row) MarkUp {
	return MarkUp{
		Rows: rows,
	}
}

type Row struct {
	Buttons []Buttons
}

type Buttons interface {
	build(lang string) tgbotapi.KeyboardButton
}

func NewRow(buttons ...Buttons) Row {
	return Row{
		Buttons: buttons,
	}
}

func (m MarkUp) Build(lang string) tgbotapi.ReplyKeyboardMarkup {
	var replyMarkUp tgbotapi.ReplyKeyboardMarkup

	for _, row := range m.Rows {
		replyMarkUp.Keyboard = append(replyMarkUp.Keyboard,
			row.buildRow(lang))
	}
	replyMarkUp.ResizeKeyboard = true
	return replyMarkUp
}

func (r Row) buildRow(lang string) []tgbotapi.KeyboardButton {
	var replyRow []tgbotapi.KeyboardButton

	for _, butt := range r.Buttons {
		replyRow = append(replyRow, butt.build(lang))
	}
	return replyRow
}

type DataButton struct {
	textKey string
}

func NewDataButton(key string) DataButton {
	return DataButton{
		textKey: key,
	}
}

func (b DataButton) build(lang string) tgbotapi.KeyboardButton {
	text := assets.LangText(lang, b.textKey)
	return tgbotapi.NewKeyboardButton(text)
}

type AdminButton struct {
	textKey string
}

func NewAdminButton(key string) AdminButton {
	return AdminButton{
		textKey: key,
	}
}

func (b AdminButton) build(lang string) tgbotapi.KeyboardButton {
	text := assets.AdminText(lang, b.textKey)
	return tgbotapi.NewKeyboardButton(text)
}

/*
==================================================
		InlineMarkUp
==================================================
*/

type InlineMarkUp struct {
	Rows []InlineRow
}

func NewIlMarkUp(rows ...InlineRow) InlineMarkUp {
	return InlineMarkUp{
		Rows: rows,
	}
}

type InlineRow struct {
	Buttons []InlineButtons
}

type InlineButtons interface {
	build(lang string) tgbotapi.InlineKeyboardButton
}

func NewIlRow(buttons ...InlineButtons) InlineRow {
	return InlineRow{
		Buttons: buttons,
	}
}

func (m InlineMarkUp) Build(lang string) tgbotapi.InlineKeyboardMarkup {
	var replyMarkUp tgbotapi.InlineKeyboardMarkup

	for _, row := range m.Rows {
		replyMarkUp.InlineKeyboard = append(replyMarkUp.InlineKeyboard,
			row.buildInlineRow(lang))
	}
	return replyMarkUp
}

func (r InlineRow) buildInlineRow(lang string) []tgbotapi.InlineKeyboardButton {
	var replyRow []tgbotapi.InlineKeyboardButton

	for _, butt := range r.Buttons {
		replyRow = append(replyRow, butt.build(lang))
	}
	return replyRow
}

type InlineDataButton struct {
	textKey string
	data    string
}

func NewIlDataButton(key, data string) InlineDataButton {
	return InlineDataButton{
		textKey: key,
		data:    data,
	}
}

func (b InlineDataButton) build(lang string) tgbotapi.InlineKeyboardButton {
	text := assets.LangText(lang, b.textKey)
	return tgbotapi.NewInlineKeyboardButtonData(text, b.data)
}

type InlineURLButton struct {
	textKey string
	url     string
}

func NewIlURLButton(key, url string) InlineURLButton {
	return InlineURLButton{
		textKey: key,
		url:     url,
	}
}

func (b InlineURLButton) build(lang string) tgbotapi.InlineKeyboardButton {
	text := assets.LangText(lang, b.textKey)
	return tgbotapi.NewInlineKeyboardButtonURL(text, b.url)
}

type InlineAdminButton struct {
	textKey string
	data    string
}

func NewIlAdminButton(key, data string) InlineAdminButton {
	return InlineAdminButton{
		textKey: key,
		data:    data,
	}
}

func (b InlineAdminButton) build(lang string) tgbotapi.InlineKeyboardButton {
	text := assets.AdminText(lang, b.textKey)
	return tgbotapi.NewInlineKeyboardButtonData(text, b.data)
}
