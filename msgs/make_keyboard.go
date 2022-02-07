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
	Text string
}

func NewDataButton(text string) DataButton {
	return DataButton{
		Text: text,
	}
}

func (b DataButton) build(lang string) tgbotapi.KeyboardButton {
	text := assets.LangText(lang, b.Text)
	return tgbotapi.NewKeyboardButton(text)
}

type AdminButton struct {
	Text string
}

func NewAdminButton(text string) AdminButton {
	return AdminButton{
		Text: text,
	}
}

func (b AdminButton) build(lang string) tgbotapi.KeyboardButton {
	text := assets.AdminText(lang, b.Text)
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
	Text string
	Data string
}

func NewIlDataButton(text, data string) InlineDataButton {
	return InlineDataButton{
		Text: text,
		Data: data,
	}
}

func (b InlineDataButton) build(lang string) tgbotapi.InlineKeyboardButton {
	text := assets.LangText(lang, b.Text)
	return tgbotapi.NewInlineKeyboardButtonData(text, b.Data)
}

type InlineURLButton struct {
	Text string
	Url  string
}

func NewIlURLButton(text, url string) InlineURLButton {
	return InlineURLButton{
		Text: text,
		Url:  url,
	}
}

func (b InlineURLButton) build(lang string) tgbotapi.InlineKeyboardButton {
	text := assets.LangText(lang, b.Text)
	return tgbotapi.NewInlineKeyboardButtonURL(text, b.Url)
}

type InlineAdminButton struct {
	Text string
	Data string
}

func NewIlAdminButton(text, data string) InlineAdminButton {
	return InlineAdminButton{
		Text: text,
		Data: data,
	}
}

func (b InlineAdminButton) build(lang string) tgbotapi.InlineKeyboardButton {
	text := assets.AdminText(lang, b.Text)
	return tgbotapi.NewInlineKeyboardButtonData(text, b.Data)
}
