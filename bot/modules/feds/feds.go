package feds

import (
	"fmt"
	"log"
	"strconv"

	"github.com/anandpskerala/Martha/bot"
	"github.com/anandpskerala/Martha/bot/modules/sql"
	"github.com/anandpskerala/Martha/bot/modules/utils/chats"
	"github.com/anandpskerala/Martha/bot/modules/utils/error_handling"
	"github.com/anandpskerala/Martha/bot/modules/utils/extraction"
	"github.com/anandpskerala/Martha/bot/modules/utils/helpers"
	"github.com/PaulSonOfLars/gotgbot"
	"github.com/PaulSonOfLars/gotgbot/ext"
	"github.com/PaulSonOfLars/gotgbot/handlers"
	"github.com/PaulSonOfLars/gotgbot/handlers/Filters"
)

func fedBan(b ext.Bot, u *gotgbot.Update, args []string) error {
	chat := u.EffectiveChat
	user := u.EffectiveUser
	msg := u.EffectiveMessage
	fedId := sql.GetFedId(strconv.Itoa(chat.Id))

	if fedId == "" {
		_, err := msg.ReplyText("This chat is not part of any federation!")
		return err
	}

	userId, reason := extraction.ExtractUserAndText(msg, args)

	if userId == 0 {
		_, err := msg.ReplyText("Try targeting a user next time bud.")
		return err
	}

	fed := sql.GetFedInfo(fedId)

	if fed == nil {
		_, err := msg.ReplyText("Please give me a valid fed ID!")
		return err
	}

	if sql.IsUserFedAdmin(fedId, strconv.Itoa(user.Id)) == "" {
		_, err := msg.ReplyHTMLf("You aren't a federation admin for <b>%v</b>", fed.FedName)
		return err
	}

	fbannedUser := sql.GetFbanUser(fedId, strconv.Itoa(userId))

	if strconv.Itoa(userId) == fed.OwnerId {
		_, err := msg.ReplyText("Why are you trying to fban the federation owner?")
		return err
	}

	if sql.IsUserFedAdmin(fedId, strconv.Itoa(userId)) != "" {
		_, err := msg.ReplyText("Why are you trying to fban a federation admin?")
		return err
	}

	if userId == bot.BotConfig.OwnerId {
		_, err := msg.ReplyText("I'm not fbanning my owner!")
		return err
	}

	for _, id := range bot.BotConfig.SudoUsers {
		sudoId, _ := strconv.Atoi(id)
		if userId == sudoId {
			_, err := msg.ReplyText("I'm not fbanning a sudo user!")
			return err
		}
	}

	if reason == "" {
		reason = "No reason."
	}

	go sql.FbanUser(fedId, strconv.Itoa(userId), reason)
	member, _ := b.GetChat(userId)

	if fbannedUser == nil {
		_, err := msg.ReplyHTMLf("Beginning federation ban of %v in %v.", helpers.MentionHtml(member.Id, member.FirstName), fed.FedName)
		error_handling.HandleErr(err)
		go func(b ext.Bot, user *ext.Chat, userId int, federations *sql.Federation, reason string) {
			for _, chat := range sql.AllFedChats(fedId) {
				chatId, err := strconv.Atoi(chat)
				error_handling.HandleErr(err)
				_, err = b.KickChatMember(chatId, userId)
				error_handling.HandleErr(err)

				_, err = b.SendMessageHTML(chatId, fmt.Sprintf("User %v is banned in the current federation "+
					"(%v), and so has been removed."+
					"\n<b>Reason</b>: %v", helpers.MentionHtml(member.Id, member.FirstName), fed.FedName, reason))
				error_handling.HandleErr(err)
			}
		}(b, member, userId, fed, reason)

		_, err = msg.ReplyHTMLf("<b>New FedBan</b>"+
			"\n<b>Fed</b>: %v"+
			"\n<b>FedAdmin</b>: %v"+
			"\n<b>User</b>: %v"+
			"\n<b>User ID</b>: <code>%v</code>"+
			"\n<b>Reason</b>: %v", fed.FedName, helpers.MentionHtml(user.Id, user.FirstName), helpers.MentionHtml(member.Id, member.FirstName),
			member.Id, reason)
		return err
	} else {
		_, err := msg.ReplyHTMLf("<b>FedBan Reason update</b>"+
			"\n<b>Fed</b>: %v"+
			"\n<b>FedAdmin</b>: %v"+
			"\n<b>User</b>: %v"+
			"\n<b>User ID</b>: <code>%v</code>"+
			"\n<b>Previous Reason</b>: %v"+
			"\n<b>New Reason</b>: %v", fed.FedName, helpers.MentionHtml(user.Id, user.FirstName), helpers.MentionHtml(member.Id, member.FirstName),
			member.Id, fbannedUser.Reason, reason)
		return err
	}
}

func unfedban(bot ext.Bot, u *gotgbot.Update, args []string) error {
	chat := u.EffectiveChat
	user := u.EffectiveUser
	msg := u.EffectiveMessage
	fedId := sql.GetFedId(strconv.Itoa(chat.Id))

	if fedId == "" {
		_, err := msg.ReplyText("This chat is not part of any federation!")
		return err
	}

	userId, _ := extraction.ExtractUserAndText(msg, args)

	if userId == 0 {
		_, err := msg.ReplyText("Try targeting a user next time bud.")
		return err
	}

	fed := sql.GetFedInfo(fedId)

	if fed == nil {
		_, err := msg.ReplyText("Please give me a valid fed ID!")
		return err
	}

	if sql.IsUserFedAdmin(fedId, strconv.Itoa(user.Id)) == "" {
		_, err := msg.ReplyHTMLf("You aren't a federation admin for <b>%v</b>", fed.FedName)
		return err
	}

	fbannedUser := sql.GetFbanUser(fedId, strconv.Itoa(userId))

	if fbannedUser == nil {
		_, err := msg.ReplyHTMLf("This user isn't banned in the current federation, <b>%v</b>.\n(<code>%v</code>)", fed.FedName, fed.Id)
		return err
	}

	go sql.UnFbanUser(fedId, strconv.Itoa(userId))

	member, _ := bot.GetChat(userId)

	go func(bot ext.Bot, user *ext.Chat, userId int, federations *sql.Federation) {
		for _, chat := range sql.AllFedChats(fedId) {
			chatId, err := strconv.Atoi(chat)
			error_handling.HandleErr(err)
			_, err = bot.UnbanChatMember(chatId, userId)
			error_handling.HandleErr(err)
		}
	}(bot, member, userId, fed)

	_, err := msg.ReplyHTMLf("<b>New un-FedBan</b>"+
		"\n<b>Fed</b>: %v"+
		"\n<b>FedAdmin</b>: %v"+
		"\n<b>User</b>: %v"+
		"\n<b>User ID</b>: <code>%v</code>", fed.FedName, helpers.MentionHtml(user.Id, user.FirstName),
		helpers.MentionHtml(member.Id, member.FirstName),
		member.Id)
	return err
}

func fedCheckBan(bot ext.Bot, u *gotgbot.Update) error {
	user := u.EffectiveUser
	msg := u.EffectiveMessage
	chat := u.EffectiveChat

	if chats.IsUserAdmin(chat, u.EffectiveUser.Id) {
		return gotgbot.ContinueGroups{}
	}

	if !chats.IsBotAdmin(chat, nil) {
		return gotgbot.ContinueGroups{}
	}

	fed := sql.GetChatFed(strconv.Itoa(chat.Id))

	if fed == nil {
		return gotgbot.ContinueGroups{}
	}
	member := sql.GetFbanUser(fed.Id, strconv.Itoa(user.Id))

	if member != nil {
		_, err := msg.Delete()
		error_handling.HandleErr(err)
		_, err = bot.KickChatMember(chat.Id, user.Id)
		if err == nil {
			_, err = bot.SendMessageHTML(chat.Id, fmt.Sprintf("User %v is banned in the current federation "+
				"(%v), and so has been removed."+
				"\n<b>Reason</b>: %v", helpers.MentionHtml(user.Id, user.FirstName), fed.FedName, member.Reason))
			return err
		}
	}
	return gotgbot.ContinueGroups{}
}

func LoadFeds(u *gotgbot.Updater) {
	defer log.Println("Loading module feds")
	u.Dispatcher.AddHandler(handlers.NewPrefixCommand("newfed", []rune{'/', '!'}, newFed))
	u.Dispatcher.AddHandler(handlers.NewPrefixCommand("delfed", []rune{'/', '!'}, delFed))
	u.Dispatcher.AddHandler(handlers.NewPrefixCommand("chatfed", []rune{'/', '!'}, chatFed))
	u.Dispatcher.AddHandler(handlers.NewPrefixArgsCommand("joinfed", []rune{'/', '!'}, joinFed))
	u.Dispatcher.AddHandler(handlers.NewPrefixCommand("leavefed", []rune{'/', '!'}, leaveFed))

	u.Dispatcher.AddHandler(handlers.NewPrefixArgsCommand("fedpromote", []rune{'/', '!'}, fedPromote))
	u.Dispatcher.AddHandler(handlers.NewPrefixArgsCommand("feddemote", []rune{'/', '!'}, fedDemote))
	u.Dispatcher.AddHandler(handlers.NewPrefixArgsCommand("fedinfo", []rune{'/', '!'}, fedInfo))
	u.Dispatcher.AddHandler(handlers.NewPrefixArgsCommand("fedadmins", []rune{'/', '!'}, fedAdmins))
	u.Dispatcher.AddHandler(handlers.NewPrefixArgsCommand("fedban", []rune{'/', '!'}, fedBan))
	u.Dispatcher.AddHandler(handlers.NewPrefixArgsCommand("unfedban", []rune{'/', '!'}, unfedban))
	u.Dispatcher.AddHandler(handlers.NewPrefixArgsCommand("fedstat", []rune{'/', '!'}, fedStat))

	// Shorter aliases
	u.Dispatcher.AddHandler(handlers.NewPrefixArgsCommand("fpromote", []rune{'/', '!'}, fedPromote))
	u.Dispatcher.AddHandler(handlers.NewPrefixArgsCommand("fdemote", []rune{'/', '!'}, fedDemote))
	u.Dispatcher.AddHandler(handlers.NewPrefixArgsCommand("finfo", []rune{'/', '!'}, fedInfo))
	u.Dispatcher.AddHandler(handlers.NewPrefixArgsCommand("fadmins", []rune{'/', '!'}, fedAdmins))
	u.Dispatcher.AddHandler(handlers.NewPrefixArgsCommand("fban", []rune{'/', '!'}, fedBan))
	u.Dispatcher.AddHandler(handlers.NewPrefixArgsCommand("unfban", []rune{'/', '!'}, unfedban))
	u.Dispatcher.AddHandler(handlers.NewPrefixArgsCommand("fbanstat", []rune{'/', '!'}, fedStat))

	u.Dispatcher.AddHandler(handlers.NewMessage(Filters.All, fedCheckBan))
}
