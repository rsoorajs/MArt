package connection

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/PaulSonOfLars/gotgbot"
	"github.com/PaulSonOfLars/gotgbot/ext"
	"github.com/PaulSonOfLars/gotgbot/handlers"
	"github.com/anandpskerala/Martha/bot/modules/sql"
	"github.com/anandpskerala/Martha/bot/modules/utils/chats"
)

func connectChat(bot ext.Bot, u *gotgbot.Update) error {

	chatId := u.EffectiveChat.Id
	chat := u.EffectiveChat
	msg := u.EffectiveMessage
	msgf := u.EffectiveMessage.From.Id
	chtId := strconv.Itoa(chatId)
	fchat := strconv.Itoa(msgf)

	args := strings.SplitN(msg.Text, " ", 2)
	if u.EffectiveChat.Type == "private" {
		if len(args) < 1 {
			msg := bot.NewSendableMessage(u.EffectiveChat.Id, "Write a chat id to be connected")
			_, err := msg.Send()
			return err
		} else {
			connectId := args[1]
			inte, _ := strconv.Atoi(connectId)
			dc, cht := bot.GetChat(inte)
			if cht != nil {
				_, err := msg.ReplyText("Chat doesn't exists")
				return err
			}
			_, err := bot.GetChatMember(inte, bot.Id)
			if err != nil {
				_, err = msg.ReplyText("Chat doesn't exists")
				return err
			}
			if !chats.RequireUserAdmin(dc, u.EffectiveMessage, u.EffectiveUser.Id) {
				return gotgbot.ContinueGroups{}
			}
			sql.SetChatConnect(chtId, connectId, dc.Title)
			_, err = msg.ReplyHTMLf("Successfully connected to <b>%v</b>", dc.Title)
			return err
		}
	}  else {
		if !chats.RequireUserAdmin(u.EffectiveChat, u.EffectiveMessage, u.EffectiveUser.Id) {
			return gotgbot.ContinueGroups{}
		}
		sql.SetChatConnect(fchat, chtId, chat.Title)
		pmText := fmt.Sprintf("Your Pm has been connected to  %v", chat.Title)
		bot.SendMessage(msgf, pmText)
		  _, err := msg.ReplyTextf("Succesfully connected you pm to %v", chat.Title)
		return err
	}
}

func disconnectChat(bot ext.Bot, u *gotgbot.Update) error {
	chatId := strconv.Itoa(u.EffectiveMessage.From.Id)
	sql.SetChatConnect(chatId, "", "")
	_, err := u.EffectiveMessage.ReplyText("Dissconnected the chat")
	return err
}

func LoadConnect(u *gotgbot.Updater) {
	defer log.Println("Loading module connection ")
	u.Dispatcher.AddHandler(handlers.NewPrefixCommand("connect", []rune{'/', '!'}, connectChat))
	u.Dispatcher.AddHandler(handlers.NewPrefixCommand("disconnect", []rune{'/', '!'}, disconnectChat))
}
