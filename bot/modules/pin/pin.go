package pin

import (
	//"fmt"
	//"html"
	"log"
	//"strconv"
	//"strings"

	"github.com/anandpskerala/Martha/bot/modules/utils/chats"
	//"github.com/anandpskerala/Martha/bot/modules/utils/error_handling"
	//"github.com/anandpskerala/Martha/bot/modules/utils/extraction"
	//"github.com/anandpskerala/Martha/bot/modules/utils/helpers"
	//"github.com/anandpskerala/Martha/bot/modules/utils/string_handling"
	"github.com/PaulSonOfLars/gotgbot"
	"github.com/PaulSonOfLars/gotgbot/ext"
	"github.com/PaulSonOfLars/gotgbot/handlers"
)

func pin(bot ext.Bot, u *gotgbot.Update, args []string) error {
	user := u.EffectiveUser
	chat := u.EffectiveChat
	msg := u.EffectiveMessage

	// Check permissions
	if chat.Type == "private" {
		_, err := msg.ReplyText("This command is meant to be used in a group!")
		return err
	}
	if !chats.RequireUserAdmin(chat, msg, user.Id) {
		return gotgbot.EndGroups{}
	}
	if !chats.RequireBotAdmin(chat, msg) {
		return gotgbot.EndGroups{}
	}
	if !chats.CanPin(bot, chat) {
		return gotgbot.EndGroups{}
	}
	prevMessage := u.EffectiveMessage.ReplyToMessage
        _, err := bot.PinChatMessage(chat.Id, prevMessage.MessageId)
	return err
}

func unpin(bot ext.Bot, u *gotgbot.Update) error {
	user := u.EffectiveUser
	chat := u.EffectiveChat
	msg := u.EffectiveMessage

	// Check permissions
	if chat.Type == "private" {
		_, err := msg.ReplyText("This command is meant to be used in a group!")
		return err
	}
	if !chats.RequireUserAdmin(chat, msg, user.Id) {
		return gotgbot.EndGroups{}
	}
	if !chats.RequireBotAdmin(chat, msg) {
		return gotgbot.EndGroups{}
	}
	if !chats.CanPin(bot, chat) {
		return gotgbot.EndGroups{}
	}

	_, err := bot.UnpinChatMessage(chat.Id)
        _, err = bot.ReplyText(chat.Id, "Successfully unpinned the message", msg.MessageId)
	return err
}

func LoadPin(u *gotgbot.Updater) {
	defer log.Println("Loading module pin")
	//u.Dispatcher.AddHandler(handlers.NewPrefixArgsCommand("demote", []rune{'/', '!'}, demote))
	u.Dispatcher.AddHandler(handlers.NewPrefixArgsCommand("pin", []rune{'/', '!'}, pin))
	u.Dispatcher.AddHandler(handlers.NewPrefixCommand("unpin", []rune{'/', '!'}, unpin))
	//u.Dispatcher.AddHandler(handlers.NewPrefixCommand("invitelink", []rune{'/', '!'}, invitelink))
	//u.Dispatcher.AddHandler(handlers.NewPrefixCommand("adminlist", []rune{'/', '!'}, adminlist))
}
