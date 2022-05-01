package auth

//
//import (
//	"database/sql"
//	"time"
//
//	"github.com/Stepan1328/miner-bot/log"
//	"github.com/Stepan1328/miner-bot/model"
//	"github.com/Stepan1328/miner-bot/msgs"
//	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
//	"github.com/pkg/errors"
//)
//
//var (
//	expectedHandlerLogger = log.NewDefaultLogger().Prefix("expected handler")
//)
//
//const (
//	getMsgAlreadyToSend               = "SELECT * FROM expected_msgs WHERE delivery_time <= ?;"
//	setNewMinWithdrawal               = "UPDATE users SET min_withdrawal = ? WHERE id = ?;"
//	updateBalanceAfterFirstWithdrawal = "UPDATE users SET balance = ?, first_withdrawal = TRUE WHERE id = ?;"
//	deleteSentMessage                 = "DELETE FROM expected_msgs WHERE token = ?;"
//
//	waitTime = 5 * time.Second
//)
//
//type expectedMessage struct {
//	botLang          string
//	token            string
//	userID           int64
//	withdrawalAmount int
//	deliveryTime     int64
//	textKey          string
//	newMinWithdrawal int
//}
//
//func HandleExpectedMsgs(botLang string) {
//	for {
//		messages, err := getMsgReadyToSend(botLang)
//		if err != nil {
//			expectedHandlerLogger.Warn("error when receiving messages from the database: %s", err.Error())
//			continue
//		}
//
//		sendExpectedMsgs(messages)
//
//		deleteAlreadyDeliveredMsgs(messages)
//
//		countOfMsgSend := len(messages)
//		if countOfMsgSend != 0 {
//			expectedHandlerLogger.Ok("// %s  delivered %d messages to users", botLang, countOfMsgSend)
//		}
//		time.Sleep(waitTime)
//	}
//}
//
//func getMsgReadyToSend(botLang string) ([]*expectedMessage, error) {
//	dataBase := model.GetDB(botLang)
//	rows, err := dataBase.Query(getMsgAlreadyToSend, time.Now().Unix())
//	if err != nil {
//		return nil, errors.Wrap(err, "get msgs")
//	}
//
//	return readMsgsFromRows(rows, botLang)
//}
//
//func readMsgsFromRows(rows *sql.Rows, botLang string) ([]*expectedMessage, error) {
//	defer rows.Close()
//
//	var msgs []*expectedMessage
//
//	for rows.Next() {
//		var (
//			token                string
//			withdrawalAmount     int
//			deliveryTime, userID int64
//			textKey              string
//			newMinWithdrawal     int
//		)
//
//		if err := rows.Scan(&token, &userID, &withdrawalAmount, &deliveryTime, &textKey, &newMinWithdrawal); err != nil {
//			return nil, errors.Wrap(err, model.ErrScanSqlRow.Error())
//		}
//
//		msgs = append(msgs, &expectedMessage{
//			botLang:          botLang,
//			token:            token,
//			userID:           userID,
//			withdrawalAmount: withdrawalAmount,
//			deliveryTime:     deliveryTime,
//			textKey:          textKey,
//			newMinWithdrawal: newMinWithdrawal,
//		})
//	}
//
//	return msgs, nil
//}
//
//func sendExpectedMsgs(msgs []*expectedMessage) {
//	for _, msg := range msgs {
//		sendExpectedMsg(msg)
//
//		editMinWithdrawal(msg)
//	}
//}
//
//func sendExpectedMsg(msg *expectedMessage) {
//	var text string
//	if msg.textKey == "withdrawal_delay_text_1" {
//		text = handlerFirstLimit(msg)
//	}
//
//	msgToSend := tgbotapi.NewMessage(msg.userID, text)
//
//	_ = msgs.SendMsgToUser(msg.botLang, msgToSend)
//}
//
//func handlerFirstLimit(msg *expectedMessage) string {
//	text := model.LangText(msg.botLang, msg.textKey)
//	//text = fmt.Sprintf(text, assets.AdminSettings.Parameters[msg.botLang].SecondLimitAmount) // TODO: refactor
//
//	user, err := GetUser(msg.botLang, msg.userID)
//	if err != nil {
//		return ""
//	}
//
//	dataBase := model.GetDB(msg.botLang)
//	rows, err := dataBase.Query(updateBalanceAfterFirstWithdrawal, msg.withdrawalAmount+user.Balance, msg.userID)
//	if err != nil {
//		return ""
//	}
//	_ = rows.Close()
//
//	return text
//}
//
//func editMinWithdrawal(msg *expectedMessage) {
//	dataBase := model.GetDB(msg.botLang)
//	rows, err := dataBase.Query(setNewMinWithdrawal, msg.newMinWithdrawal, msg.userID)
//	if err != nil {
//		return
//	}
//
//	_ = rows.Close()
//}
//
//func deleteAlreadyDeliveredMsgs(msgs []*expectedMessage) {
//	for _, msg := range msgs {
//		deleteSendMsg(msg)
//	}
//}
//
//func deleteSendMsg(msg *expectedMessage) {
//	dataBase := model.GetDB(msg.botLang)
//	rows, err := dataBase.Query(deleteSentMessage, msg.token)
//	if err != nil {
//		return
//	}
//
//	_ = rows.Close()
//}
