package purge

import (
	"log"
	"time"

	"github.com/anandpskerala/Martha/bot/modules/utils/chats"
	"github.com/anandpskerala/Martha/bot/modules/utils/error_handling"
	"github.com/PaulSonOfLars/gotgbot"
	"github.com/PaulSonOfLars/gotgbot/ext"
	"github.com/PaulSonOfLars/gotgbot/handlers"
)

func purge(bot ext.Bot, u *gotgbot.Update) error {
	msg := u.EffectiveMessage
	user := u.EffectiveUser
	chat := u.EffectiveChat

	// Permission checks
	if !chats.RequireBotAdmin(chat, msg) {
		return gotgbot.EndGroups{}
	}
	if !chats.RequireUserAdmin(chat, msg, user.Id) {
		return gotgbot.EndGroups{}
	}

	if msg.ReplyToMessage != nil {
		if chats.CanDelete(chat, bot.Id) {
			msgId := msg.ReplyToMessage.MessageId
			deleteTo := msg.MessageId - 1
			for mId := deleteTo; mId > msgId-1; mId-- {
				_, err := bot.DeleteMessage(chat.Id, mId)
				if err != nil {
					if err.Error() == "Bad Request: message can't be deleted" {
						_, err := bot.SendMessage(chat.Id, "Cannot delete all messages. The messages may be too old, I might "+
							"not have delete rights, or this might not be a supergroup.")
						error_handling.HandleErr(err)
					} else if err.Error() != "Bad Request: message to delete not found" {
						error_handling.HandleErr(err)
					}
				}
			}
			_, err := msg.Delete()
			if err != nil {
				if err.Error() == "Bad Request: message can't be deleted" {
					_, err := bot.SendMessage(chat.Id, "Cannot delete all messages. The messages may be too old, I might "+
						"not have delete rights, or this might not be a supergroup.")
					error_handling.HandleErr(err)
				} else if err.Error() == "Bad Request: message to delete not found" {
					error_handling.HandleErr(err)
				}
			}
			delMsg, err := bot.SendMessage(chat.Id, "Purge complete.")
			error_handling.HandleErr(err)
			time.Sleep(2 * time.Second)
			_, err = bot.DeleteMessage(chat.Id, delMsg.MessageId)
			return err
		}
		return nil
	} else {
		_, err := msg.ReplyText("Reply to a message to select where to start purging from.")
		return err
	}
}

func delMessage(bot ext.Bot, u *gotgbot.Update) error {
	msg := u.EffectiveMessage
	user := u.EffectiveUser
	chat := u.EffectiveChat

	// Permission checks
	if !chats.RequireBotAdmin(chat, msg) {
		return gotgbot.EndGroups{}
	}
	if !chats.RequireUserAdmin(chat, msg, user.Id) {
		return gotgbot.EndGroups{}
	}

	if msg.ReplyToMessage != nil {
		if chats.CanDelete(chat, bot.Id) {
			_, err := msg.ReplyToMessage.Delete()
			error_handling.HandleErr(err)
			_, err = msg.Delete()
			return err
		}
	} else {
		_, err := msg.ReplyText("What do u want to delete?")
		return err
	}
	return nil
}

func LoadPurge(u *gotgbot.Updater) {
	defer log.Println("Loading module purge")
	u.Dispatcher.AddHandler(handlers.NewPrefixCommand("purge", []rune{'/', '!'}, purge))
	u.Dispatcher.AddHandler(handlers.NewPrefixCommand("del", []rune{'/', '!'}, delMessage))
}
